def build_context_nouns(
    normalized_noun: str,
    post_normalized_nouns: list[str] | None,
) -> set[str]:
    if not post_normalized_nouns:
        return set()

    return {
        noun
        for noun in post_normalized_nouns
        if noun and noun != normalized_noun
    }


def find_context_supported_group_ids(cur, context_nouns: set[str]) -> set[int]:
    if not context_nouns:
        return set()

    context_noun_list = list(context_nouns)
    cur.execute(
        """
        SELECT DISTINCT gk.group_master_id
        FROM group_keywords gk
        INNER JOIN group_masters gm
            ON gm.id = gk.group_master_id
        WHERE gk.active = TRUE
            AND gm.active = TRUE
            AND gk.match_type <> 'requires_context'
            AND (
                gk.normalized_keyword = ANY(%s::text[])
                OR gk.keyword = ANY(%s::text[])
                OR (
                    gk.match_type IN ('partial', 'exact_or_partial')
                    AND EXISTS (
                        SELECT 1
                        FROM unnest(%s::text[]) AS context_noun(value)
                        WHERE
                            context_noun.value LIKE '%%' || gk.normalized_keyword || '%%'
                            OR gk.normalized_keyword LIKE '%%' || context_noun.value || '%%'
                            OR context_noun.value LIKE '%%' || gk.keyword || '%%'
                            OR gk.keyword LIKE '%%' || context_noun.value || '%%'
                    )
                )
            )
        """,
        (context_noun_list, context_noun_list, context_noun_list),
    )

    return {row["group_master_id"] for row in cur.fetchall()}


# 曖昧語の管理はコードではなく group_keywords.match_type に寄せる。
def has_context_required_keyword(cur, normalized_noun: str) -> bool:
    cur.execute(
        """
        SELECT EXISTS (
            SELECT 1
            FROM group_keywords gk
            INNER JOIN group_masters gm
                ON gm.id = gk.group_master_id
            WHERE gk.active = TRUE
                AND gm.active = TRUE
                AND gk.match_type = 'requires_context'
                AND (
                    gk.normalized_keyword = %s
                    OR gk.keyword = %s
                )
        ) AS requires_context
        """,
        (normalized_noun, normalized_noun),
    )
    row = cur.fetchone()
    return bool(row and row["requires_context"])


def find_context_required_group_ids(
    cur,
    normalized_noun: str,
    noun_requires_context: bool,
) -> set[int]:
    cur.execute(
        build_context_required_group_query(noun_requires_context),
        build_context_required_group_params(normalized_noun, noun_requires_context),
    )

    return {row["group_master_id"] for row in cur.fetchall()}


def build_context_required_group_query(noun_requires_context: bool) -> str:
    context_required_partial_condition = ""
    if noun_requires_context:
        context_required_partial_condition = """
                OR (
                    gk.match_type IN ('partial', 'exact_or_partial')
                    AND (
                        %s LIKE '%%' || gk.normalized_keyword || '%%'
                        OR gk.normalized_keyword LIKE '%%' || %s || '%%'
                        OR %s LIKE '%%' || gk.keyword || '%%'
                        OR gk.keyword LIKE '%%' || %s || '%%'
                    )
                )
        """

    return f"""
        SELECT DISTINCT gk.group_master_id
        FROM group_keywords gk
        INNER JOIN group_masters gm
            ON gm.id = gk.group_master_id
        WHERE gk.active = TRUE
            AND gm.active = TRUE
            AND (
                (
                    gk.match_type = 'requires_context'
                    AND (
                        gk.normalized_keyword = %s
                        OR gk.keyword = %s
                    )
                )
                {context_required_partial_condition}
            )
        """


def build_context_required_group_params(
    normalized_noun: str,
    noun_requires_context: bool,
) -> tuple[str, ...]:
    params = [normalized_noun, normalized_noun]
    if noun_requires_context:
        params.extend(
            [
                normalized_noun,
                normalized_noun,
                normalized_noun,
                normalized_noun,
            ]
        )

    return tuple(params)
