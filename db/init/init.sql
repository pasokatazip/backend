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

-- posts
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    pet_id  UUID NOT NULL,
    content VARCHAR(255) NOT NULL,
    content_embedding VECTOR(384),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
)

-- pets
CREATE TABLE pets (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    user_id UUID NOT NULL REFERENCES users(id),
    energy INT DEFAULT 0,
    curiosity INT DEFAULT 0,
    sociality INT DEFAULT 0,
    routine INT DEFAULT 0,
    current_group_master_id UUID,
    current_stage_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- USER_ACTIVE_PETS(
CREATE TABLE user_active_pets (
    user_id UUID NOT NULL REFERENCES users(id),
    pet_id UUID NOT NULL REFERENCES pets(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, pet_id)
);

-- PET_EXPERIENCES
CREATE TABLE pet_experiences (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    total_experience BIGINT NOT NULL DEFAULT 0,
    feed_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- PET_EXPERIENCES_EVENTS
CREATE TABLE pet_experiences_events (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    experience_delta INT NOT NULL,
    feed_count INT DEFAULT 0,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
