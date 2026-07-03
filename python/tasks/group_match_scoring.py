import math

from tasks.group_match_types import GroupMatchCandidate

EXACT_KEYWORD_SCORE = 1.0
PARTIAL_KEYWORD_SCORE = 0.7
VECTOR_CANDIDATE_THRESHOLD = 0.45
CONTEXT_REQUIRED_MATCH_SCORE = 0.0
KEYWORD_SCORE_FACTOR = 0.57
VECTOR_SCORE_FACTOR = 0.40
KEYWORD_WEIGHT_FACTOR = 0.03


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


# group_keywords.weight は微調整に留め、名詞一致とベクトル類似を主軸にする。
def calculate_match_score(
    keyword_score: float,
    vector_score: float,
    keyword_weight: float,
) -> float:
    return round(
        keyword_score * KEYWORD_SCORE_FACTOR
        + vector_score * VECTOR_SCORE_FACTOR
        + keyword_weight * KEYWORD_WEIGHT_FACTOR,
        5,
    )


def is_context_required_match(
    normalized_noun: str,
    keyword: str,
    normalized_keyword: str,
    match_type: str,
    noun_requires_context: bool,
) -> bool:
    if match_type == "requires_context":
        return True

    if not noun_requires_context:
        return False

    if normalized_noun == normalized_keyword or normalized_noun == keyword:
        return True

    return match_type in ("partial", "exact_or_partial") and (
        normalized_noun in normalized_keyword
        or normalized_keyword in normalized_noun
        or normalized_noun in keyword
        or keyword in normalized_noun
    )


def select_best_candidate_index(candidates: list[GroupMatchCandidate]) -> int | None:
    selectable_indexes = [
        index
        for index, candidate in enumerate(candidates)
        if candidate.selectable
    ]
    if not selectable_indexes:
        return None

    return max(
        selectable_indexes,
        key=lambda index: candidate_sort_key(candidates[index]),
    )


def sorted_candidates(candidates) -> list[GroupMatchCandidate]:
    return sorted(candidates, key=candidate_sort_key, reverse=True)


def is_better_candidate(candidate: GroupMatchCandidate, current: GroupMatchCandidate) -> bool:
    return candidate_sort_key(candidate) > candidate_sort_key(current)


def candidate_sort_key(candidate: GroupMatchCandidate) -> tuple[float, float, float, float, float]:
    return (
        1.0 if candidate.selectable else 0.0,
        candidate.match_score,
        candidate.keyword_score,
        candidate.vector_score,
        candidate.keyword_weight,
    )
