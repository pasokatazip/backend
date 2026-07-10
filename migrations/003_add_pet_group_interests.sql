-- +goose Up
CREATE TABLE IF NOT EXISTS pet_group_interests (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id) ON DELETE CASCADE,
    interest_score DECIMAL(12, 5) NOT NULL DEFAULT 0,
    last_matched_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_pet_group_interests_pet_group UNIQUE (pet_id, group_master_id),
    CONSTRAINT chk_pet_group_interests_score CHECK (interest_score >= 0)
);

CREATE INDEX IF NOT EXISTS idx_pet_group_interests_pet_score
ON pet_group_interests(pet_id, interest_score DESC);

CREATE INDEX IF NOT EXISTS idx_pet_group_interests_group_master_id
ON pet_group_interests(group_master_id);

-- +goose Down
DROP TABLE IF EXISTS pet_group_interests;
