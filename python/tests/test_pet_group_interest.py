import unittest

from tasks.pet_group_interest import add_selected_group_interests_for_post


class FakeCursor:
    def __init__(self, interests):
        self.interests = interests
        self.executions = []

    def execute(self, query, params):
        self.executions.append((query, params))

    def fetchall(self):
        return self.interests


class PetGroupInterestTest(unittest.TestCase):
    def test_adds_aggregated_selected_scores_to_each_group(self):
        cursor = FakeCursor(
            [
                {"group_master_id": 3, "interest_score": 1.25},
                {"group_master_id": 8, "interest_score": 0.75},
            ]
        )

        count = add_selected_group_interests_for_post(
            cursor,
            pet_id="pet-id",
            post_id="post-id",
        )

        self.assertEqual(count, 2)
        self.assertEqual(cursor.executions[0][1], ("post-id",))
        self.assertEqual(len(cursor.executions), 3)

        first_upsert_params = cursor.executions[1][1]
        self.assertEqual(first_upsert_params[1:], ("pet-id", 3, 1.25))

        second_upsert_params = cursor.executions[2][1]
        self.assertEqual(second_upsert_params[1:], ("pet-id", 8, 0.75))

    def test_skips_upsert_when_no_selected_group_exists(self):
        cursor = FakeCursor([])

        count = add_selected_group_interests_for_post(
            cursor,
            pet_id="pet-id",
            post_id="post-id",
        )

        self.assertEqual(count, 0)
        self.assertEqual(len(cursor.executions), 1)


if __name__ == "__main__":
    unittest.main()
