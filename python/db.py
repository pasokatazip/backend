# db.py

import os
import psycopg2
from psycopg2.extras import RealDictCursor

DATABASE_URL = os.getenv("DATABASE_URL")

if not DATABASE_URL:
    raise ValueError("DATABASE_URL is not set")


def get_connection():
    return psycopg2.connect(
        DATABASE_URL,
        cursor_factory=RealDictCursor,
    )


def get_listener_connection():
    conn = get_connection()
    try:
        # ensure notifications are delivered promptly
        conn.autocommit = True
    except Exception:
        pass
    return conn