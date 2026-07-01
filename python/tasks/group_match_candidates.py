from tasks.group_master import ActiveGroup
from tasks.group_match_scoring import (
    CONTEXT_REQUIRED_MATCH_SCORE,
    VECTOR_CANDIDATE_THRESHOLD,
    calculate_keyword_score,
    calculate_match_score,
    calculate_vector_score,
    is_better_candidate,
    is_context_required_match,
    sorted_candidates,
)
from tasks.group_match_types import GroupMatchCandidate


def find_keyword_candidates(
    cur,
    normalized_noun: str,
    noun_embedding: list[float],
    group_by_id: dict[int, ActiveGroup],
    context_supported_group_ids: set[int],
    noun_requires_context: bool,
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

        candidate = build_keyword_candidate(
            row=row,
            normalized_noun=normalized_noun,
            noun_embedding=noun_embedding,
            group=group,
            context_supported_group_ids=context_supported_group_ids,
            noun_requires_context=noun_requires_context,
        )
        current = best_by_group.get(candidate.group_master_id)
        if current is None or is_better_candidate(candidate, current):
            best_by_group[candidate.group_master_id] = candidate

    return sorted_candidates(best_by_group.values())


def find_vector_candidates(
    noun_embedding: list[float],
    groups: list[ActiveGroup],
    context_required_group_ids: set[int],
    context_supported_group_ids: set[int],
) -> list[GroupMatchCandidate]:
    candidates: list[GroupMatchCandidate] = []
    for group in groups:
        if (
            group.group_master_id in context_required_group_ids
            and group.group_master_id not in context_supported_group_ids
        ):
            continue

        vector_score = calculate_vector_score(noun_embedding, group.display_name_embedding)
        if vector_score < VECTOR_CANDIDATE_THRESHOLD:
            continue

        candidates.append(
            GroupMatchCandidate(
                group_master_id=group.group_master_id,
                keyword_score=0.0,
                vector_score=vector_score,
                keyword_weight=0.0,
                match_score=calculate_match_score(
                    keyword_score=0.0,
                    vector_score=vector_score,
                    keyword_weight=0.0,
                ),
                match_reason=f"vector:{group.display_name}",
            )
        )

    return sorted_candidates(candidates)


def build_keyword_candidate(
    row,
    normalized_noun: str,
    noun_embedding: list[float],
    group: ActiveGroup,
    context_supported_group_ids: set[int],
    noun_requires_context: bool,
) -> GroupMatchCandidate:
    normalized_keyword = row["normalized_keyword"]
    keyword = row["keyword"]
    match_type = row["match_type"]
    keyword_score = calculate_keyword_score(
        normalized_noun=normalized_noun,
        keyword=keyword,
        normalized_keyword=normalized_keyword,
        match_type=match_type,
    )
    vector_score = calculate_vector_score(noun_embedding, group.display_name_embedding)
    keyword_weight = float(row["weight"])
    selectable = True
    match_reason = f"keyword:{keyword}->{normalized_keyword};vector:{group.display_name}"

    requires_context = is_context_required_match(
        normalized_noun=normalized_noun,
        keyword=keyword,
        normalized_keyword=normalized_keyword,
        match_type=match_type,
        noun_requires_context=noun_requires_context,
    )
    if requires_context and row["group_master_id"] not in context_supported_group_ids:
        keyword_score = 0.0
        vector_score = 0.0
        keyword_weight = 0.0
        selectable = False
        match_reason = f"keyword:{keyword}->{normalized_keyword};context_required_but_missing"

    match_score = calculate_match_score(
        keyword_score=keyword_score,
        vector_score=vector_score,
        keyword_weight=keyword_weight,
    )
    if not selectable:
        match_score = CONTEXT_REQUIRED_MATCH_SCORE

    return GroupMatchCandidate(
        group_master_id=row["group_master_id"],
        keyword_score=keyword_score,
        vector_score=vector_score,
        keyword_weight=keyword_weight,
        match_score=match_score,
        match_reason=match_reason,
        selectable=selectable,
    )


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
