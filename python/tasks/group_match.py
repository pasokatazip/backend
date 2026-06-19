from dataclasses import dataclass
import logging
import math

from tasks.group_master import ActiveGroup, find_active_groups

logger = logging.getLogger(__name__)


EXACT_KEYWORD_SCORE = 1.0
PARTIAL_KEYWORD_SCORE = 0.7
VECTOR_CANDIDATE_THRESHOLD = 0.45


@dataclass(frozen=True)
class GroupMatchCandidate:
    group_master_id: int
    keyword_score: float
    vector_score: float
    keyword_weight: float
    match_score: float
    match_reason: str


# 抽出名詞1件に対して、群れ候補を作成して noun_group_matches に保存する入口。
def create_noun_group_matches(
    cur,
    extracted_noun_id: int,
    normalized_noun: str,
    noun_embedding: list[float],
) -> list[GroupMatchCandidate]:
    logger.info(
        "create_noun_group_matches start extracted_noun_id=%s normalized_noun=%s",
        extracted_noun_id,
        normalized_noun,
    )

    groups = find_active_groups(cur)
    group_by_id = {group.group_master_id: group for group in groups}
    keyword_candidates = find_keyword_candidates(cur, normalized_noun, noun_embedding, group_by_id)
    vector_candidates = find_vector_candidates(noun_embedding, groups)
    candidates = merge_candidates(keyword_candidates, vector_candidates)

    if not candidates:
        logger.info(
            "create_noun_group_matches no candidates extracted_noun_id=%s",
            extracted_noun_id,
        )
        return []

    best_index = select_best_candidate_index(candidates)
    for index, candidate in enumerate(candidates):
        cur.execute(
            """
            INSERT INTO noun_group_matches (
                extracted_noun_id,
                group_master_id,
                keyword_score,
                vector_score,
                keyword_weight,
                match_score,
                match_reason,
                selected
            )
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
            """,
            (
                extracted_noun_id,
                candidate.group_master_id,
                candidate.keyword_score,
                candidate.vector_score,
                candidate.keyword_weight,
                candidate.match_score,
                candidate.match_reason,
                index == best_index,
            ),
        )

    logger.info(
        "create_noun_group_matches done extracted_noun_id=%s candidate_count=%d selected_group_master_id=%s",
        extracted_noun_id,
        len(candidates),
        candidates[best_index].group_master_id,
    )
    return candidates


# group_keywords の完全一致・部分一致から候補群れを作る。
def find_keyword_candidates(
    cur,
    normalized_noun: str,
    noun_embedding: list[float],
    group_by_id: dict[int, ActiveGroup],
) -> list[GroupMatchCandidate]:
    cur.execute(
        """
        SELECT
            gk.group_master_id,
            gk.keyword,
            gk.normalized_keyword,
            gk.weight,
            gk.match_type
        FROM group_keywords gk
        INNER JOIN group_masters gm
            ON gm.id = gk.group_master_id
        WHERE gk.active = TRUE
            AND gm.active = TRUE
            AND (
                gk.normalized_keyword = %s
                OR gk.keyword = %s
                OR (
                    gk.match_type IN ('partial', 'exact_or_partial')
                    AND (
                        %s LIKE '%%' || gk.normalized_keyword || '%%'
                        OR gk.normalized_keyword LIKE '%%' || %s || '%%'
                        OR %s LIKE '%%' || gk.keyword || '%%'
                        OR gk.keyword LIKE '%%' || %s || '%%'
                    )
                )
            )
        ORDER BY gk.weight DESC, gk.id
        """,
        (
            normalized_noun,
            normalized_noun,
            normalized_noun,
            normalized_noun,
            normalized_noun,
            normalized_noun,
        ),
    )

    best_by_group: dict[int, GroupMatchCandidate] = {}
    for row in cur.fetchall():
        group = group_by_id.get(row["group_master_id"])
        if group is None:
            continue

        candidate = build_keyword_candidate(row, normalized_noun, noun_embedding, group)
        current = best_by_group.get(candidate.group_master_id)
        if current is None or is_better_candidate(candidate, current):
            best_by_group[candidate.group_master_id] = candidate

    return sorted_candidates(best_by_group.values())


# embedding 類似度だけで候補群れを作る。キーワード未登録の表現を拾うために使う。
def find_vector_candidates(
    noun_embedding: list[float],
    groups: list[ActiveGroup],
) -> list[GroupMatchCandidate]:
    candidates: list[GroupMatchCandidate] = []
    for group in groups:
        vector_score = calculate_vector_score(noun_embedding, group.display_name_embedding)
        if vector_score < VECTOR_CANDIDATE_THRESHOLD:
            continue

        match_score = calculate_match_score(
            keyword_score=0.0,
            vector_score=vector_score,
            keyword_weight=0.0,
        )
        candidates.append(
            GroupMatchCandidate(
                group_master_id=group.group_master_id,
                keyword_score=0.0,
                vector_score=vector_score,
                keyword_weight=0.0,
                match_score=match_score,
                match_reason=f"vector:{group.display_name}",
            )
        )

    return sorted_candidates(candidates)


# keyword で見つかった1行を、vector_score も含む候補データに変換する。
def build_keyword_candidate(
    row,
    normalized_noun: str,
    noun_embedding: list[float],
    group: ActiveGroup,
) -> GroupMatchCandidate:
    normalized_keyword = row["normalized_keyword"]
    keyword = row["keyword"]
    keyword_score = calculate_keyword_score(
        normalized_noun=normalized_noun,
        keyword=keyword,
        normalized_keyword=normalized_keyword,
        match_type=row["match_type"],
    )
    vector_score = calculate_vector_score(noun_embedding, group.display_name_embedding)
    keyword_weight = float(row["weight"])
    match_score = calculate_match_score(
        keyword_score=keyword_score,
        vector_score=vector_score,
        keyword_weight=keyword_weight,
    )

    return GroupMatchCandidate(
        group_master_id=row["group_master_id"],
        keyword_score=keyword_score,
        vector_score=vector_score,
        keyword_weight=keyword_weight,
        match_score=match_score,
        match_reason=f"keyword:{keyword}->{normalized_keyword};vector:{group.display_name}",
    )


# keyword 候補と vector 候補をまとめ、同じ群れは最も良い候補だけ残す。
def merge_candidates(
    keyword_candidates: list[GroupMatchCandidate],
    vector_candidates: list[GroupMatchCandidate],
) -> list[GroupMatchCandidate]:
    best_by_group: dict[int, GroupMatchCandidate] = {}
    for candidate in keyword_candidates + vector_candidates:
        current = best_by_group.get(candidate.group_master_id)
        if current is None or is_better_candidate(candidate, current):
            best_by_group[candidate.group_master_id] = candidate

    return sorted_candidates(best_by_group.values())


# 名詞と keyword の一致の強さを数値化する。
def calculate_keyword_score(
    normalized_noun: str,
    keyword: str,
    normalized_keyword: str,
    match_type: str,
) -> float:
    if normalized_noun == normalized_keyword or normalized_noun == keyword:
        return EXACT_KEYWORD_SCORE

    if match_type in ("partial", "exact_or_partial") and (
        normalized_noun in normalized_keyword
        or normalized_keyword in normalized_noun
        or normalized_noun in keyword
        or keyword in normalized_noun
    ):
        return PARTIAL_KEYWORD_SCORE

    return 0.0


# 名詞 embedding と群れ embedding の cosine 類似度を 0.0 から 1.0 に丸めて返す。
def calculate_vector_score(noun_embedding: list[float], group_embedding: list[float]) -> float:
    if not noun_embedding or not group_embedding or len(noun_embedding) != len(group_embedding):
        return 0.0

    dot = sum(noun_value * group_value for noun_value, group_value in zip(noun_embedding, group_embedding))
    noun_norm = math.sqrt(sum(noun_value * noun_value for noun_value in noun_embedding))
    group_norm = math.sqrt(sum(group_value * group_value for group_value in group_embedding))
    if noun_norm == 0 or group_norm == 0:
        return 0.0

    cosine_similarity = dot / (noun_norm * group_norm)
    return round(max(0.0, min(1.0, cosine_similarity)), 5)


# keyword_score / vector_score / keyword_weight を合成して最終スコアを作る。
def calculate_match_score(
    keyword_score: float,
    vector_score: float,
    keyword_weight: float,
) -> float:
    return round(
        keyword_score * 0.55
        + vector_score * 0.35
        + keyword_weight * 0.10,
        5,
    )


# candidates の中で selected=true にする候補の位置を決める。
def select_best_candidate_index(candidates: list[GroupMatchCandidate]) -> int:
    return max(
        range(len(candidates)),
        key=lambda index: candidate_sort_key(candidates[index]),
    )


# 候補一覧を、移動先として強い順に並べる。
def sorted_candidates(candidates) -> list[GroupMatchCandidate]:
    return sorted(candidates, key=candidate_sort_key, reverse=True)


# 既存候補より新しい候補のほうが良いかを判定する。
def is_better_candidate(candidate: GroupMatchCandidate, current: GroupMatchCandidate) -> bool:
    return candidate_sort_key(candidate) > candidate_sort_key(current)


# 候補を比較するときの優先順位を定義する。
def candidate_sort_key(candidate: GroupMatchCandidate) -> tuple[float, float, float, float]:
    return (
        candidate.match_score,
        candidate.keyword_score,
        candidate.vector_score,
        candidate.keyword_weight,
    )
