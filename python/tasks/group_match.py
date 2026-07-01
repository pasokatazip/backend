import logging

from tasks.group_master import find_active_groups
from tasks.group_match_candidates import (
    find_keyword_candidates,
    find_vector_candidates,
    merge_candidates,
)
from tasks.group_match_context import (
    build_context_nouns,
    find_context_required_group_ids,
    find_context_supported_group_ids,
    has_context_required_keyword,
)
from tasks.group_match_scoring import select_best_candidate_index
from tasks.group_match_types import GroupMatchCandidate

logger = logging.getLogger(__name__)


# 抽出名詞1件に対して、群れ候補を作成して noun_group_matches に保存する入口。
def create_noun_group_matches(
    cur,
    extracted_noun_id: int,
    normalized_noun: str,
    noun_embedding: list[float],
    post_normalized_nouns: list[str] | None = None,
) -> list[GroupMatchCandidate]:
    logger.info(
        "create_noun_group_matches start extracted_noun_id=%s normalized_noun=%s",
        extracted_noun_id,
        normalized_noun,
    )

    groups = find_active_groups(cur)
    group_by_id = {group.group_master_id: group for group in groups}
    context_nouns = build_context_nouns(normalized_noun, post_normalized_nouns)
    context_supported_group_ids = find_context_supported_group_ids(cur, context_nouns)
    noun_requires_context = has_context_required_keyword(cur, normalized_noun)
    context_required_group_ids = find_context_required_group_ids(
        cur,
        normalized_noun,
        noun_requires_context,
    )

    candidates = merge_candidates(
        find_keyword_candidates(
            cur=cur,
            normalized_noun=normalized_noun,
            noun_embedding=noun_embedding,
            group_by_id=group_by_id,
            context_supported_group_ids=context_supported_group_ids,
            noun_requires_context=noun_requires_context,
        ),
        find_vector_candidates(
            noun_embedding=noun_embedding,
            groups=groups,
            context_required_group_ids=context_required_group_ids,
            context_supported_group_ids=context_supported_group_ids,
        ),
    )

    if not candidates:
        logger.info(
            "create_noun_group_matches no candidates extracted_noun_id=%s",
            extracted_noun_id,
        )
        return []

    save_noun_group_matches(cur, extracted_noun_id, candidates)

    best_index = select_best_candidate_index(candidates)
    logger.info(
        "create_noun_group_matches done extracted_noun_id=%s candidate_count=%d selected_group_master_id=%s",
        extracted_noun_id,
        len(candidates),
        candidates[best_index].group_master_id if best_index is not None else None,
    )
    return candidates


def save_noun_group_matches(
    cur,
    extracted_noun_id: int,
    candidates: list[GroupMatchCandidate],
) -> None:
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
                best_index is not None and index == best_index,
            ),
        )
