import logging
from uuid import uuid4

logger = logging.getLogger(__name__)


def add_selected_group_interests_for_post(cur, pet_id: str, post_id: str) -> int:
    """投稿内で選ばれた群れ候補のスコアを、ペットの興味として累積する。"""
    cur.execute(
        """
        SELECT
            ngm.group_master_id,
            SUM(ngm.match_score) AS interest_score
        FROM noun_group_matches ngm
        INNER JOIN extracted_nouns en
            ON en.id = ngm.extracted_noun_id
        INNER JOIN group_masters gm
            ON gm.id = ngm.group_master_id
        WHERE en.post_id = %s
            AND ngm.selected = TRUE
            AND gm.active = TRUE
        GROUP BY ngm.group_master_id
        HAVING SUM(ngm.match_score) > 0
        ORDER BY ngm.group_master_id
        """,
        (post_id,),
    )
    interests = cur.fetchall()

    for interest in interests:
        cur.execute(
            """
            INSERT INTO pet_group_interests (
                id,
                pet_id,
                group_master_id,
                interest_score,
                last_matched_at,
                created_at,
                updated_at
            )
            VALUES (%s, %s, %s, %s, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
            ON CONFLICT (pet_id, group_master_id) DO UPDATE
            SET
                interest_score = pet_group_interests.interest_score + EXCLUDED.interest_score,
                last_matched_at = EXCLUDED.last_matched_at,
                updated_at = EXCLUDED.updated_at
            """,
            (
                str(uuid4()),
                pet_id,
                interest["group_master_id"],
                float(interest["interest_score"]),
            ),
        )

    logger.info(
        "updated pet group interests pet_id=%s post_id=%s group_count=%d",
        pet_id,
        post_id,
        len(interests),
    )
    return len(interests)
