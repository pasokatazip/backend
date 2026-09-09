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
                            (gk.normalized_keyword <> '' AND POSITION(gk.normalized_keyword IN context_noun.value) > 0)
                            OR (gk.keyword <> '' AND POSITION(gk.keyword IN context_noun.value) > 0)
                    )
                )
            )
        """,
        (context_noun_list, context_noun_list, context_noun_list),
    )

    return {row["group_master_id"] for row in cur.fetchall()}


# 文脈必須の設定は該当する群れにだけ適用し、通常キーワードへ波及させない。
def find_context_required_group_ids(cur, normalized_noun: str) -> set[int]:
    cur.execute(
        """
        SELECT DISTINCT gk.group_master_id
        FROM group_keywords gk
        INNER JOIN group_masters gm ON gm.id = gk.group_master_id
        WHERE gk.active = TRUE AND gm.active = TRUE
            AND gk.match_type = 'requires_context'
            AND (gk.normalized_keyword = %s OR gk.keyword = %s)
        """,
        (normalized_noun, normalized_noun),
    )
    return {row["group_master_id"] for row in cur.fetchall()}
