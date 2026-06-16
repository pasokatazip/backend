import select
import logging
import traceback


logging.basicConfig(
    level=logging.INFO,
    format="[%(asctime)s] %(levelname)s: %(message)s",
)

from db import get_listener_connection
from listeners.post import process_post


def main():
    try:
        conn = get_listener_connection()
    except Exception:
        logging.exception("Failed to get listener connection")
        return

    try:
        with conn.cursor() as cur:
            cur.execute("LISTEN post_created;")
            logging.info("LISTEN registered for post_created")
    except Exception:
        logging.exception("Failed to register LISTEN; closing connection")
        try:
            conn.close()
        except Exception:
            pass
        return

    logging.info("worker started")

    try:
        while True:
            ready = select.select([conn], [], [], 10)
            if ready == ([], [], []):
                continue

            try:
                conn.poll()
            except Exception:
                logging.exception("conn.poll() failed")
                break

            while getattr(conn, 'notifies', []):
                notify = conn.notifies.pop(0)
                post_id = notify.payload
                logging.info("received notify payload=%s", post_id)
                try:
                    process_post(post_id)
                except Exception:
                    logging.exception("process_post failed for %s", post_id)
    except KeyboardInterrupt:
        logging.info("worker interrupted by user")
    except Exception:
        logging.exception("Unhandled exception in worker loop")
    finally:
        try:
            conn.close()
        except Exception:
            pass


if __name__ == "__main__":
    main()