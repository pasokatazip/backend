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
    display_name_embedding vector(1536),
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

--group_keywords
CREATE TABLE IF NOT EXISTS group_keywords (
    id SERIAL PRIMARY KEY,
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    normalized_keyword VARCHAR(100) NOT NULL,
    weight DECIMAL(5, 2) NOT NULL DEFAULT 1.00,
    match_type VARCHAR(50) NOT NULL DEFAULT 'exact_or_partial',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_group_keywords_lookup ON group_keywords (normalized_keyword, active);

--extracted_nouns
CREATE TABLE IF NOT EXISTS extracted_nouns (
    id SERIAL PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    noun_text VARCHAR(255) NOT NULL,
    normalized_noun VARCHAR(255) NOT NULL,
    noun_embedding VECTOR(384),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_extracted_nouns_post_id ON extracted_nouns(post_id);

CREATE INDEX IF NOT EXISTS idx_extracted_nouns_normalized_noun ON extracted_nouns(normalized_noun);

--noun_group_matches
CREATE TABLE IF NOT EXISTS noun_group_matches (
    id SERIAL PRIMARY KEY,
    extracted_noun_id INTEGER NOT NULL REFERENCES extracted_nouns(id) ON DELETE CASCADE,
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id) ON DELETE CASCADE,
    keyword_score DECIMAL(8, 5) NOT NULL DEFAULT 0,
    vector_score DECIMAL(8, 5) NOT NULL DEFAULT 0,
    keyword_weight DECIMAL(5, 2) NOT NULL DEFAULT 0,
    priority_score DECIMAL(8, 5) NOT NULL DEFAULT 0,
    match_score DECIMAL(8, 5) NOT NULL DEFAULT 0,
    match_reason VARCHAR(255),
    selected BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_noun_group_matches_extracted_noun_id ON noun_group_matches(extracted_noun_id);

CREATE INDEX IF NOT EXISTS idx_noun_group_matches_group_master_id ON noun_group_matches(group_master_id);

CREATE INDEX IF NOT EXISTS idx_noun_group_matches_selected ON noun_group_matches(selected);

CREATE INDEX IF NOT EXISTS idx_noun_group_matches_match_score ON noun_group_matches(match_score);
