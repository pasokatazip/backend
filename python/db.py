# db.py

import os
import psycopg2
from psycopg2.extras import RealDictCursor
from urllib.parse import parse_qsl, urlencode, urlsplit, urlunsplit

DATABASE_URL = os.getenv("DATABASE_URL")

if not DATABASE_URL:
    raise ValueError("DATABASE_URL is not set")


def normalize_database_url_for_psycopg(database_url: str) -> str:
    parts = urlsplit(database_url)
    query_params = [
        (key, value)
        for key, value in parse_qsl(parts.query, keep_blank_values=True)
        if key.lower() != "timezone"
    ]
    return urlunsplit(
        (
            parts.scheme,
            parts.netloc,
            parts.path,
            urlencode(query_params),
            parts.fragment,
        )
    )


def get_connection():
    return psycopg2.connect(
        normalize_database_url_for_psycopg(DATABASE_URL),
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
