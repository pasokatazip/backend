from db import get_connection
from tasks.embedding import create_embedding, to_pgvector
from tasks.group_match import create_noun_group_matches
from tasks.noun_extract import extract_nouns
from tasks.pet_movement import choose_adopted_group_for_post, move_pet_to_adopted_group
import logging
import traceback

logger = logging.getLogger(__name__)


def process_post(post_id: str):
    logger.info("process_post start: %s", post_id)
    conn = get_connection()

    try:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, pet_id, content
                FROM posts
                WHERE id = %s
                """,
                (post_id,),
            )
            post = cur.fetchone()
            if post is None:
                logger.warning("post not found: %s", post_id)
                return

            pet_id = post["pet_id"]
            content = post["content"]
            logger.info("fetched post id=%s content_len=%d", post_id, len(content or ""))

            logger.info("creating embedding for post %s", post_id)
            embedding = create_embedding(content)
            logger.info("embedding created for post %s length=%d", post_id, len(embedding))

            pg_embedding = to_pgvector(embedding)

            logger.info("updating DB embedding for post %s", post_id)
            cur.execute(
                """
                UPDATE posts
                SET content_embedding = %s
                WHERE id = %s
                """,
                (pg_embedding, post_id),
            )

            logger.info("extracting nouns for post %s", post_id)
            extracted_nouns = extract_nouns(content)
            logger.info(
                "extracted nouns for post %s count=%d",
                post_id,
                len(extracted_nouns),
            )

            logger.info("refreshing extracted_nouns for post %s", post_id)
            cur.execute(
                """
                DELETE FROM extracted_nouns
                WHERE post_id = %s
                """,
                (post_id,),
            )

            post_normalized_nouns = [
                noun.normalized_noun
                for noun in extracted_nouns
            ]

            for noun in extracted_nouns:
                noun_embedding = create_embedding(noun.normalized_noun)
                pg_noun_embedding = to_pgvector(noun_embedding)

                cur.execute(
                    """
                    INSERT INTO extracted_nouns (
                        post_id,
                        noun_text,
                        normalized_noun,
                        noun_embedding
                    )
                    VALUES (%s, %s, %s, %s)
                    RETURNING id
                    """,
                    (
                        post_id,
                        noun.noun_text,
                        noun.normalized_noun,
                        pg_noun_embedding,
                    ),
                )
                extracted_noun = cur.fetchone()
                create_noun_group_matches(
                    cur=cur,
                    extracted_noun_id=extracted_noun["id"],
                    normalized_noun=noun.normalized_noun,
                    noun_embedding=noun_embedding,
                    post_normalized_nouns=post_normalized_nouns,
                )

            adopted_group = choose_adopted_group_for_post(cur, post_id)
            if adopted_group is None:
                logger.info("no adopted group found for post %s", post_id)
            else:
                move_pet_to_adopted_group(
                    cur=cur,
                    pet_id=pet_id,
                    adopted_group=adopted_group,
                )

        conn.commit()
        logger.info("process_post completed and committed: %s", post_id)

    except Exception:
        try:
            conn.rollback()
        except Exception:
            logger.exception("rollback failed")
        logger.exception("worker failed for post %s", post_id)
        logger.debug(traceback.format_exc())

    finally:
        try:
            conn.close()
        except Exception:
            logger.exception("failed to close connection for post %s", post_id)
