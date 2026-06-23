from dataclasses import dataclass
import logging
from uuid import uuid4

logger = logging.getLogger(__name__)

MOVE_REASON_FEED_MATCH = "feed_match"
STATUS_MIN = 0
STATUS_MAX = 100


@dataclass(frozen=True)
class AdoptedGroup:
    group_master_id: int
    match_score: float
    extracted_noun_id: int
    noun_text: str
    normalized_noun: str


@dataclass(frozen=True)
class PetGroupJoin:
    id: str
    pet_id: str
    group_master_id: int
    move_reason: str


# 投稿に紐づく pet_id が実在するか確認する。
def pet_exists(cur, pet_id: str) -> bool:
    cur.execute(
        """
        SELECT EXISTS (
            SELECT 1
            FROM pets
            WHERE id = %s
        ) AS exists
        """,
        (pet_id,),
    )
    row = cur.fetchone()
    return bool(row and row["exists"])


# 投稿内で selected=true になった候補から、最終的に移動する群れを1つ選ぶ。
def choose_adopted_group_for_post(cur, post_id: str) -> AdoptedGroup | None:
    cur.execute(
        """
        SELECT
            ngm.group_master_id,
            ngm.match_score,
            ngm.keyword_score,
            ngm.vector_score,
            ngm.keyword_weight,
            en.id AS extracted_noun_id,
            en.noun_text,
            en.normalized_noun
        FROM noun_group_matches ngm
        INNER JOIN extracted_nouns en
            ON en.id = ngm.extracted_noun_id
        INNER JOIN group_masters gm
            ON gm.id = ngm.group_master_id
        WHERE en.post_id = %s
            AND ngm.selected = TRUE
            AND gm.active = TRUE
        ORDER BY
            ngm.match_score DESC,
            ngm.keyword_score DESC,
            ngm.vector_score DESC,
            ngm.keyword_weight DESC,
            ngm.id ASC
        LIMIT 1
        """,
        (post_id,),
    )
    row = cur.fetchone()
    if row is None:
        return None

    return AdoptedGroup(
        group_master_id=row["group_master_id"],
        match_score=float(row["match_score"]),
        extracted_noun_id=row["extracted_noun_id"],
        noun_text=row["noun_text"],
        normalized_noun=row["normalized_noun"],
    )


# 採用群れにペットを移動し、pet_group_joins に移動ログを残す。
def move_pet_to_adopted_group(
    cur,
    pet_id: str,
    adopted_group: AdoptedGroup,
    move_reason: str = MOVE_REASON_FEED_MATCH,
) -> PetGroupJoin:
    join_id = str(uuid4())

    # いま開いている参加レコードを閉じる。
    cur.execute(
        """
        UPDATE pet_group_joins
        SET
            left_at = CURRENT_TIMESTAMP,
            updated_at = CURRENT_TIMESTAMP
        WHERE pet_id = %s
            AND left_at IS NULL
        """,
        (pet_id,),
    )

    # 新しい群れへの参加を記録する。
    cur.execute(
        """
        INSERT INTO pet_group_joins (
            id,
            pet_id,
            group_master_id,
            joined_at,
            move_reason,
            created_at,
            updated_at
        )
        VALUES (
            %s,
            %s,
            %s,
            CURRENT_TIMESTAMP,
            %s,
            CURRENT_TIMESTAMP,
            CURRENT_TIMESTAMP
        )
        """,
        (
            join_id,
            pet_id,
            adopted_group.group_master_id,
            move_reason,
        ),
    )

    # pets に現在地を持たせ、画面表示や次回シミュレーションの起点にする。
    cur.execute(
        """
        UPDATE pets
        SET
            current_group_master_id = %s,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = %s
        """,
        (
            adopted_group.group_master_id,
            pet_id,
        ),
    )

    apply_group_status_delta(cur, pet_id, adopted_group.group_master_id)

    logger.info(
        "moved pet to adopted group pet_id=%s group_master_id=%s match_score=%.5f normalized_noun=%s",
        pet_id,
        adopted_group.group_master_id,
        adopted_group.match_score,
        adopted_group.normalized_noun,
    )

    return PetGroupJoin(
        id=join_id,
        pet_id=pet_id,
        group_master_id=adopted_group.group_master_id,
        move_reason=move_reason,
    )


# group_masters の delta をペットの現在ステータスに反映する。
def apply_group_status_delta(cur, pet_id: str, group_master_id: int) -> None:
    cur.execute(
        """
        UPDATE pets p
        SET
            energy = GREATEST(
                %s,
                LEAST(%s, p.energy + gm.energy_delta)
            ),
            curiosity = GREATEST(
                %s,
                LEAST(%s, p.curiosity + gm.curiosity_delta)
            ),
            sociality = GREATEST(
                %s,
                LEAST(%s, p.sociality + gm.sociality_delta)
            ),
            routine = GREATEST(
                %s,
                LEAST(%s, p.routine + gm.routine_delta)
            ),
            updated_at = CURRENT_TIMESTAMP
        FROM group_masters gm
        WHERE p.id = %s
            AND gm.id = %s
        """,
        (
            STATUS_MIN,
            STATUS_MAX,
            STATUS_MIN,
            STATUS_MAX,
            STATUS_MIN,
            STATUS_MAX,
            STATUS_MIN,
            STATUS_MAX,
            pet_id,
            group_master_id,
        ),
    )
