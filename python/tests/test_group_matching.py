import unittest

from tasks.group_master import ActiveGroup
from tasks.group_match_candidates import build_keyword_candidate, find_vector_candidates, merge_candidates
from tasks.group_match_context import build_context_nouns
from tasks.group_match_scoring import calculate_keyword_score, select_best_candidate_index
from tasks.noun_extract import extract_nouns, should_ignore_normalized_noun


class GroupMatchingRegressionTests(unittest.TestCase):
    def candidate(self, group_id, noun, match_type, support=()):
        return build_keyword_candidate(
            row=dict(group_master_id=group_id, keyword=noun, normalized_keyword=noun,
                     match_type=match_type, weight=1.0),
            normalized_noun=noun, noun_embedding=[1.0, 0.0],
            group=ActiveGroup(group_id, noun, [1.0, 0.0]),
            context_supported_group_ids=set(support),
        )

    def test_canonical_photo_and_music_remain_selectable_without_context(self):
        # 同じ名詞でも、関連分野の文脈必須キーワードは本来の群れを妨げない。
        for noun in ('写真', '音楽'):
            with self.subTest(noun=noun):
                canonical = self.candidate(1, noun, 'exact_or_partial')
                ambiguous = self.candidate(2, noun, 'requires_context')
                candidates = merge_candidates([ambiguous, canonical], [])
                self.assertTrue(canonical.selectable)
                self.assertFalse(ambiguous.selectable)
                self.assertEqual(candidates[select_best_candidate_index(candidates)].group_master_id, 1)

    def test_context_required_keyword_unlocks_only_with_support_for_its_group(self):
        self.assertFalse(self.candidate(2, '写真', 'requires_context', support=[1]).selectable)
        self.assertTrue(self.candidate(2, '写真', 'requires_context', support=[2]).selectable)

    def test_vector_cannot_bypass_explicit_context_requirement(self):
        groups = [ActiveGroup(1, '写真', [1.0, 0.0]), ActiveGroup(2, 'スマホ', [1.0, 0.0])]
        self.assertEqual([c.group_master_id for c in find_vector_candidates([1.0, 0.0], groups, {2}, set())], [1])
        self.assertEqual(len(find_vector_candidates([1.0, 0.0], groups, {2}, {2})), 2)

    def test_short_generic_noun_does_not_match_more_specific_keyword(self):
        for noun, keyword in [('続き', '晴天続き'), ('写真', '青空写真'), ('続き', '機種変更手続き')]:
            with self.subTest(noun=noun, keyword=keyword):
                self.assertEqual(calculate_keyword_score(noun, keyword, keyword, 'exact_or_partial'), 0)

    def test_compound_noun_still_matches_contained_keyword(self):
        self.assertEqual(calculate_keyword_score('スマホゲーム', 'ゲーム', 'ゲーム', 'exact_or_partial'), .7)
        self.assertEqual(calculate_keyword_score('写真', '写真', '写真', 'exact_or_partial'), 1)
        self.assertEqual(calculate_keyword_score('スマホゲーム', 'ゲーム', 'ゲーム', 'exact'), 0)

    def test_empty_keywords_never_match_every_noun(self):
        self.assertEqual(calculate_keyword_score('写真', '', '', 'partial'), 0)
        self.assertEqual(calculate_keyword_score('', '', '', 'exact'), 0)

    def test_noun_cannot_provide_its_own_context(self):
        self.assertEqual(build_context_nouns('写真', ['写真', '写真']), set())
        self.assertEqual(build_context_nouns('写真', ['写真', 'カメラ']), {'カメラ'})

    def test_dates_and_sequence_numbers_are_excluded_by_real_tokenizer(self):
        nouns = extract_nouns('2026-09-10 朝 20。ゲームの新しいステージを楽しんだ。')
        values = {n.normalized_noun for n in nouns}
        self.assertIn('ゲーム', values)
        self.assertFalse(any(value.isdecimal() for value in values))

    def test_names_containing_numbers_are_not_blanket_excluded(self):
        for name in ['3d', 'ff14', 'iphone16']:
            self.assertFalse(should_ignore_normalized_noun(name))
        for number in ['2026', '09', '２０']:
            self.assertTrue(should_ignore_normalized_noun(number))

    def test_empty_or_date_only_post_has_no_group_nouns(self):
        self.assertEqual(extract_nouns(''), [])
        self.assertEqual(extract_nouns('2026-09-10 朝 20'), [])

    def test_continuation_does_not_add_an_unrelated_topic(self):
        values = {n.normalized_noun for n in extract_nouns('映画の続きを観た。')}
        self.assertIn('映画', values)
        self.assertNotIn('続き', values)

    def test_interview_practice_requires_interview_context(self):
        self.assertFalse(self.candidate(2, '練習', 'requires_context').selectable)
        self.assertTrue(self.candidate(2, '練習', 'requires_context', support=[2]).selectable)


if __name__ == '__main__':
    unittest.main()
