import copy
import datetime as dt
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

import automated_test_posts as runner


class FakeAPI:
    def __init__(self):
        self.history = []
        self.creates = 0
        self.fail_after_save = False

    def posts(self, pid):
        return self.history

    def create(self, pid, content):
        self.creates += 1
        row = {"id": "test-post", "petid": pid, "content": content,
               "createdat": "2026-09-10T09:00:00+09:00"}
        self.history.append(row)
        if self.fail_after_save:
            raise TimeoutError("応答前に切断")
        return {"id": row["id"], "pet_id": pid, "content": content}


class RunnerTests(unittest.TestCase):
    def setUp(self):
        self.now = dt.datetime(2026, 9, 10, 9, tzinfo=runner.JST)
        self.api = FakeAPI()
        self.state = {}
        self.saved = []

    def send(self, apply=True, slot="morning"):
        return runner.send_one(self.api, 1, self.now, slot, self.state,
                               lambda: self.saved.append(copy.deepcopy(self.state)), apply)

    def test_schedule_boundaries_in_japan(self):
        for hour, expected in [(0,None),(8,None),(9,"morning"),(20,"morning"),(21,"evening"),(23,"evening")]:
            with self.subTest(hour=hour):
                self.assertEqual(runner.current_slot(self.now.replace(hour=hour)), expected)
        self.assertEqual(runner.current_slot(dt.datetime(2026,9,10,12,tzinfo=dt.timezone.utc)), "evening")

    def test_exactly_twenty_targets_and_short_varying_posts(self):
        self.assertEqual(len(runner.TOPICS), 20)
        self.assertEqual(len({runner.pet_id(n) for n in range(1,21)}), 20)
        for n in range(1,21):
            texts = {runner.content_for(n,self.now.date()+dt.timedelta(days=d),s)
                     for d in range(3) for s in ("morning","evening")}
            self.assertEqual(len(texts), 3)
            self.assertTrue(all(len(t)<255 for t in texts))
            self.assertTrue(all("自動テスト" not in t and "【" not in t and "2026-" not in t for t in texts))
        for n in (0,21):
            with self.assertRaises(ValueError):
                runner.pet_id(n)

    def test_api_timestamp_fractional_precision(self):
        for fraction in ("", ".1", ".12", ".123", ".3554", ".01712", ".123456", ".123456789"):
            with self.subTest(fraction=fraction):
                value = runner.parse_timestamp(f"2026-09-09T17:04:22{fraction}+00:00")
                self.assertEqual(value.astimezone(runner.JST).date(), self.now.date())
        self.assertEqual(runner.parse_timestamp("2026-09-10T00:00:00Z"), self.now)
        with self.assertRaises(ValueError):
            runner.parse_timestamp("2026-09-10T00:00:00")

    def test_dry_run_never_posts_or_persists(self):
        self.assertEqual(self.send(False), "would_post")
        self.assertEqual(self.api.creates, 0)
        self.assertEqual(self.saved, [])

    def test_repeat_does_not_post_again(self):
        self.assertEqual(self.send(), "posted")
        self.assertEqual(self.send(), "already_posted")
        self.assertEqual(self.api.creates, 1)
        self.assertEqual(next(iter(self.saved[0].values()))["status"], "pending")

    def test_lost_local_state_reconciles_from_server(self):
        self.send()
        self.state.clear()
        self.assertEqual(self.send(), "already_posted")
        self.assertEqual(self.api.creates, 1)

    def test_timeout_after_commit_reconciles_without_repost(self):
        self.api.fail_after_save = True
        with self.assertRaises(TimeoutError):
            self.send()
        self.assertEqual(self.send(), "already_posted")
        self.assertEqual(self.api.creates, 1)

    def test_unknown_result_requires_review(self):
        self.api.fail_after_save = True
        with self.assertRaises(TimeoutError):
            self.send()
        self.api.history.clear()
        with self.assertRaisesRegex(RuntimeError,"再送を停止"):
            self.send()
        self.assertEqual(self.api.creates, 1)

    def test_any_two_posts_today_reach_daily_limit(self):
        self.api.history = [{"content":"既存投稿", "createdat":"2026-09-09T15:01:00Z"}] * 2
        self.assertEqual(self.send(), "daily_limit")
        self.assertEqual(self.api.creates, 0)

    def test_yesterdays_posts_do_not_count(self):
        self.api.history = [{"content":"昨日", "createdat":"2026-09-09T14:59:00Z"}] * 2
        self.assertEqual(self.send(), "posted")

    def test_duplicate_remote_posts_are_reported(self):
        self.send()
        self.api.history *= 2
        with self.assertRaisesRegex(RuntimeError,"複数"):
            self.send()

    def test_edited_body_is_recognized_by_saved_post_id(self):
        self.send()
        self.api.history[0]["content"] = "本文を修正しました。"
        self.assertEqual(self.send(), "already_posted")
        self.assertEqual(self.api.creates, 1)

    def test_identical_text_on_another_day_is_not_a_duplicate(self):
        self.api.history = [{"id":"older", "content":runner.content_for(1,self.now.date(),"morning"),
                             "createdat":"2026-09-07T09:00:00+09:00"}]
        self.assertEqual(self.send(), "posted")

    def test_same_text_in_morning_does_not_block_evening(self):
        self.api.history = [{"id":"morning", "content":runner.content_for(1,self.now.date(),"evening"),
                             "createdat":"2026-09-10T09:00:00+09:00"}]
        self.now = self.now.replace(hour=21)
        self.assertEqual(self.send(slot="evening"), "posted")

    def test_initial_early_morning_post_is_recognized_without_state(self):
        self.api.history = [{"id":"early", "content":runner.content_for(1,self.now.date(),"morning"),
                             "createdat":"2026-09-10T02:00:00+09:00"}]
        self.assertEqual(self.send(), "already_posted")
        self.assertEqual(self.api.creates, 0)

    def test_get_retries_but_post_does_not(self):
        api = runner.API()
        with patch.object(api, "request", side_effect=[TimeoutError(), None]) as request, patch.object(runner.time,"sleep"):
            self.assertEqual(api.posts(runner.pet_id(1)), [])
            self.assertEqual(request.call_count, 2)
        with patch.object(api, "request", side_effect=TimeoutError()) as request:
            with self.assertRaises(TimeoutError):
                api.create(runner.pet_id(1), "料理をした")
            self.assertEqual(request.call_count, 1)

    def test_wrong_pet_response_is_rejected(self):
        api = runner.API()
        with patch.object(api,"request",return_value=[{"PetID":runner.pet_id(2),"Content":"料理"}]):
            with self.assertRaises(ValueError):
                api.posts(runner.pet_id(1))

    def test_state_written_as_complete_json(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory)/"state.json"
            runner.save_state(path, {"a":{"status":"pending"}})
            runner.save_state(path, {"a":{"status":"confirmed"}})
            self.assertEqual(json.loads(path.read_text())["a"]["status"], "confirmed")
            self.assertFalse(path.with_suffix(".tmp").exists())


if __name__ == "__main__":
    unittest.main()
