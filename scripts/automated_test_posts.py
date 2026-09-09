#!/usr/bin/env python3
"""登録済み20匹専用。投稿APIを通して経験値・進化処理も実行する。"""
import argparse
import datetime as dt
import fcntl
import json
import os
from pathlib import Path
import re
import sys
import time
import urllib.error
import urllib.request

API_URL = "https://api.akatukii.com"
JST = dt.timezone(dt.timedelta(hours=9))
STATE_DIR = Path(__file__).resolve().parent / ".automated-test-posts"

# seed_automated_test_accounts.sql の番号順。日付に応じて言い回しを変える。
TOPICS = (
    ("ゲーム", "ゲームの続きを少し進めた", "ゲームの新しいステージを楽しんだ", "ゲームの攻略を考えた"),
    ("料理", "野菜のスープを料理した", "料理のレシピを試した", "料理の下ごしらえをした"),
    ("散歩", "公園を散歩した", "散歩で小さな花を見つけた", "川沿いの散歩を楽しんだ"),
    ("読書", "読書で物語の続きを楽しんだ", "読書のために本を選んだ", "読書で好きな一節を見つけた"),
    ("音楽", "好きな音楽を聴いた", "音楽のプレイリストを作った", "新しい音楽を探した"),
    ("お絵描き", "お絵描きで花を描いた", "お絵描きの色を選んだ", "お絵描きで風景を描いた"),
    ("プログラミング", "プログラミングで小さな機能を作った", "プログラミングの練習をした", "プログラミングで動作を確かめた"),
    ("カフェ", "カフェでコーヒーを飲んだ", "カフェでゆっくり過ごした", "カフェのケーキを楽しんだ"),
    ("写真", "空の写真を撮った", "写真の構図を考えた", "花の写真を整理した"),
    ("栽培", "栽培している植物に水をあげた", "栽培しているハーブを眺めた", "栽培している芽が少し伸びた"),
    ("運動", "軽い運動をした", "運動で体を伸ばした", "運動のあとゆっくり休んだ"),
    ("映画", "映画の続きを観た", "映画の好きな場面を思い出した", "次に観たい映画を選んだ"),
    ("アニメ", "アニメを一話観た", "アニメの続きを楽しんだ", "好きなアニメの場面を思い出した"),
    ("手芸", "手芸で小さな飾りを作った", "手芸の糸を選んだ", "手芸の続きを少し進めた"),
    ("旅行", "旅行の行き先を考えた", "旅行の地図を眺めた", "旅行で見たい景色を探した"),
    ("ペット", "ペットのおもちゃを片づけた", "ペットとゆっくり遊んだ", "ペットの寝顔を眺めた"),
    ("掃除", "机の掃除をした", "部屋の掃除を少し進めた", "棚の掃除をした"),
    ("ボードゲーム", "ボードゲームのルールを読んだ", "ボードゲームで遊んだ", "ボードゲームの作戦を考えた"),
    ("語学", "語学の単語を練習した", "語学の例文を読んだ", "語学の発音を練習した"),
    ("睡眠", "睡眠のために寝具を整えた", "睡眠の前にゆっくり過ごした", "睡眠をとってのんびり過ごした"),
)


def pet_id(number):
    if not 1 <= number <= 20:
        raise ValueError("テストアカウント番号は1〜20です")
    return f"77100000-0000-4000-8000-{number:012d}"


def current_slot(now):
    hour = now.astimezone(JST).hour
    return "evening" if hour >= 21 else "morning" if hour >= 9 else None


def parse_timestamp(value):
    # macOS標準Python 3.9は小数秒4/5桁を直接解析できないため6桁にそろえる。
    match = re.fullmatch(r"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,9}))?(Z|[+-]\d{2}:\d{2})", value)
    if match is None:
        raise ValueError("APIの日時にはISO形式の日時とタイムゾーンが必要です")
    fraction = (match.group(2) or "0")[:6].ljust(6, "0")
    zone = "+00:00" if match.group(3) == "Z" else match.group(3)
    return dt.datetime.fromisoformat(f"{match.group(1)}.{fraction}{zone}")


def content_for(number, day, slot):
    variant = (day.toordinal() + number + (slot == "evening")) % 3 + 1
    label = "朝" if slot == "morning" else "夜"
    return f"【自動テスト {day.isoformat()} {label} {number:02d}】{TOPICS[number-1][variant]}。"


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise RuntimeError("想定外のAPIリダイレクトを停止しました")


class API:
    def __init__(self):
        self.opener = urllib.request.build_opener(NoRedirect())

    def request(self, path, payload=None):
        body = None if payload is None else json.dumps(payload, ensure_ascii=False).encode()
        request = urllib.request.Request(
            API_URL + path, data=body,
            headers={"Content-Type": "application/json", "User-Agent": "YoyoTestAccounts/1.0"},
        )
        with self.opener.open(request, timeout=25) as response:
            if response.status != (200 if payload is None else 201):
                raise RuntimeError(f"想定外のHTTP応答: {response.status}")
            return json.load(response)

    def posts(self, pid):
        # GETだけを再試行する。送信結果不明のPOSTを自動再送してはいけない。
        for attempt in range(3):
            try:
                result = self.request("/posts/" + pid)
                if result is None:
                    return []
                if not isinstance(result, list):
                    raise ValueError("投稿一覧が配列ではありません")
                normalized = []
                for row in result:
                    row = {key.replace("_", "").lower(): value for key, value in row.items()}
                    if row.get("petid") != pid or not isinstance(row.get("content"), str):
                        raise ValueError("投稿一覧のペットIDまたは本文が不正です")
                    parse_timestamp(row["createdat"])
                    normalized.append(row)
                return normalized
            except (urllib.error.URLError, TimeoutError):
                if attempt == 2:
                    raise
                time.sleep(2 ** attempt)

    def create(self, pid, content):
        row = self.request("/posts?pet_id=" + pid, {"content": content})
        if row.get("pet_id") != pid or row.get("content") != content or not row.get("id"):
            raise ValueError("投稿応答が送信内容と一致しません。再送せず確認してください")
        return row


def save_state(path, state):
    # POSTの前にpendingを永続化し、プロセスが落ちても盲目的な再送を防ぐ。
    temporary = path.with_suffix(".tmp")
    with temporary.open("w", encoding="utf-8") as file:
        json.dump(state, file, ensure_ascii=False, indent=2)
        file.flush()
        os.fsync(file.fileno())
    os.replace(temporary, path)


def send_one(api, number, now, slot, state, persist, apply):
    day = now.astimezone(JST).date()
    pid = pet_id(number)
    content = content_for(number, day, slot)
    key = f"{pid}/{day.isoformat()}/{slot}"
    history = api.posts(pid)
    matches = [post for post in history if post["content"] == content]
    if len(matches) > 1:
        raise RuntimeError(f"テスト{number:02d}: 同じ枠の投稿が複数あります")
    if matches:
        if apply:
            state[key] = {"status": "confirmed", "post_id": matches[0]["id"]}
            persist()
        return "already_posted"
    # ローカル記録とAPIが矛盾したら、人が確認するまで停止する。
    if key in state:
        raise RuntimeError(f"テスト{number:02d}: 前回の送信結果が未確認です。再送を停止しました")
    today_count = sum(
        parse_timestamp(post["createdat"]).astimezone(JST).date() == day
        for post in history
    )
    if today_count >= 2:
        return "daily_limit"
    if not apply:
        return "would_post"
    state[key] = {"status": "pending", "attempted_at": now.isoformat(), "content": content}
    persist()
    response = api.create(pid, content)
    state[key] = {"status": "confirmed", "post_id": response["id"]}
    persist()
    return "posted"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--apply", action="store_true", help="実際に投稿する。省略時は確認のみ")
    parser.add_argument("--slot", choices=("morning", "evening"), help="初回確認用。通常は現在時刻から選ぶ")
    args = parser.parse_args()
    now = dt.datetime.now(JST)
    slot = args.slot or current_slot(now)
    if slot is None:
        print(json.dumps({"status": "not_due", "next": "09:00 Asia/Tokyo"}))
        return 0
    STATE_DIR.mkdir(mode=0o700, parents=True, exist_ok=True)
    with (STATE_DIR / "run.lock").open("a") as lock:
        try:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            print("別の投稿処理が実行中です", file=sys.stderr)
            return 1
        state_path = STATE_DIR / "state.json"
        state = json.loads(state_path.read_text()) if state_path.exists() else {}
        if not isinstance(state, dict):
            raise ValueError("投稿状態ファイルが壊れています")
        api = API()
        results, errors = {}, []
        for number in range(1, 21):
            try:
                # バッチが9時/21時/日付をまたいだ場合は次回に任せる。
                live_now = dt.datetime.now(JST)
                if live_now.date() != now.date() or (not args.slot and current_slot(live_now) != slot):
                    raise RuntimeError("実行中に投稿枠が切り替わりました")
                results[f"{number:02d}"] = send_one(
                    api, number, now, slot, state,
                    lambda: save_state(state_path, state), args.apply,
                )
            except (OSError, ValueError, RuntimeError, KeyError, TypeError) as error:
                errors.append(f"テスト{number:02d}: {type(error).__name__}: {error}")
        print(json.dumps({"date": now.date().isoformat(), "slot": slot,
                          "apply": args.apply, "results": results, "errors": errors}, ensure_ascii=False))
        return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
