import select
from db import get_listener_connection
from listeners.post import process_post


def main():
    conn = get_listener_connection()

    with conn.cursor() as cur:
        cur.execute("LISTEN post_created;")

    print("worker started")

    while True:
        if select.select([conn], [], [], 10) == ([], [], []):
            continue

        conn.poll() 

        while conn.notifies:
            notify = conn.notifies.pop(0)
            post_id = notify.payload
            process_post(post_id)


if __name__ == "__main__":
    main()