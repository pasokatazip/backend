from dataclasses import dataclass

from tasks.embedding import create_embedding, to_pgvector


@dataclass(frozen=True)
class ActiveGroup:
    group_master_id: int
    display_name: str
    display_name_embedding: list[float]


# active な群れを取得し、表示名の embedding が未作成なら作成して保存する。
def find_active_groups(cur) -> list[ActiveGroup]:
    cur.execute(
        """
        SELECT
            id,
            display_name,
            display_name_embedding
        FROM group_masters
        WHERE active = TRUE
        ORDER BY id
        """
    )

    groups: list[ActiveGroup] = []
    for row in cur.fetchall():
        embedding = parse_pgvector(row["display_name_embedding"])
        if embedding is None:
            embedding = create_embedding(row["display_name"])
            cur.execute(
                """
                UPDATE group_masters
                SET display_name_embedding = %s
                WHERE id = %s
                """,
                (to_pgvector(embedding), row["id"]),
            )

        groups.append(
            ActiveGroup(
                group_master_id=row["id"],
                display_name=row["display_name"],
                display_name_embedding=embedding,
            )
        )

    return groups


# DBから取得した pgvector 文字列を Python の float 配列に戻す。
def parse_pgvector(value) -> list[float] | None:
    if value is None:
        return None

    if isinstance(value, list):
        return [float(item) for item in value]

    vector_text = str(value).strip()
    if not vector_text:
        return None

    vector_text = vector_text.removeprefix("[").removesuffix("]")
    return [float(item) for item in vector_text.split(",") if item]
