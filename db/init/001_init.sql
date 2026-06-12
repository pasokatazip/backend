--pg-vecter
CREATE EXTENSION IF NOT EXISTS vector;

--users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    subsc BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--posts
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL,
    content VARCHAR(255) NOT NULL,
    content_embedding VECTOR(384),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--group_masters
CREATE TABLE IF NOT EXISTS group_masters (
    id SERIAL PRIMARY KEY,
    group_key VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    category VARCHAR(100),
    min_pet_count INTEGER NOT NULL DEFAULT 0,
    energy_delta DECIMAL(5, 4) NOT NULL DEFAULT 0,
    curiosity_delta DECIMAL(5, 4) NOT NULL DEFAULT 0,
    sociality_delta DECIMAL(5, 4) NOT NULL DEFAULT 0,
    routine_delta DECIMAL(5, 4) NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_group_masters_active ON group_masters(active);

CREATE INDEX IF NOT EXISTS idx_group_masters_category ON group_masters(category);