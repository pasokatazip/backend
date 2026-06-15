from db import get_connection
from tasks.embedding import create_embedding, to_pgvector


def process_post(post_id: str):
    conn = get_connection()

    try:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, content
                FROM posts
                WHERE id = %s
                """,
                (post_id,),
            )
            post = cur.fetchone()

            if post is None:
                return

            content = post["content"]

            embedding = create_embedding(content)
            pg_embedding = to_pgvector(embedding)

            cur.execute(
                """
                UPDATE posts
                SET content_embedding = %s
                WHERE id = %s
                """,
                (pg_embedding, post_id),
            )

        conn.commit()

    except Exception as e:
        conn.rollback()
        print("worker failed:", e)

    finally:
        conn.close()