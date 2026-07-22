-- +goose Up
ALTER TABLE
    pets
ADD
    COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'active';

UPDATE
    pets
SET
    status = 'active'
WHERE
    status IS NULL;

ALTER TABLE
    pets
ALTER COLUMN
    status
SET
    DEFAULT 'active';

ALTER TABLE
    pets
ALTER COLUMN
    status
SET
    NOT NULL;

-- +goose StatementBegin
DO $$ BEGIN IF NOT EXISTS (
    SELECT
        1
    FROM
        pg_constraint
    WHERE
        conname = 'chk_pets_status'
        AND conrelid = 'pets' :: regclass
) THEN
ALTER TABLE
    pets
ADD
    CONSTRAINT chk_pets_status CHECK (
        status IN ('active', 'departed', 'lost', 'archived')
    );

END IF;

END;

$$;

-- +goose StatementEnd
CREATE INDEX IF NOT EXISTS idx_pets_status ON pets(status);

-- db/init の初期値（50）を、既存 migration から作成したDBにも反映
ALTER TABLE
    pets
ALTER COLUMN
    energy
SET
    DEFAULT 50;

ALTER TABLE
    pets
ALTER COLUMN
    curiosity
SET
    DEFAULT 50;

ALTER TABLE
    pets
ALTER COLUMN
    sociality
SET
    DEFAULT 50;

ALTER TABLE
    pets
ALTER COLUMN
    routine
SET
    DEFAULT 50;

-- 投稿から得た群れへの関心を、ペット単位で累積
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

CREATE INDEX IF NOT EXISTS idx_pet_group_interests_pet_score ON pet_group_interests(pet_id, interest_score DESC);

CREATE INDEX IF NOT EXISTS idx_pet_group_interests_group_master_id ON pet_group_interests(group_master_id);

CREATE INDEX IF NOT EXISTS idx_pet_group_interests_last_matched_at ON pet_group_interests(last_matched_at);

-- 同じ群れにいたペット同士で伝わった関心の履歴を保存
CREATE TABLE IF NOT EXISTS pet_interest_propagations (
    id UUID PRIMARY KEY,
    recipient_pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    source_pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    source_pet_hourly_log_id UUID NOT NULL REFERENCES pet_hourly_logs(id) ON DELETE CASCADE,
    via_group_master_id INTEGER NOT NULL REFERENCES group_masters(id),
    propagated_group_master_id INTEGER NOT NULL REFERENCES group_masters(id),
    amount DECIMAL(8, 5) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_pet_interest_propagations_different_pets CHECK (recipient_pet_id <> source_pet_id),
    CONSTRAINT chk_pet_interest_propagations_amount CHECK (amount > 0),
    CONSTRAINT uq_pet_interest_propagations_source UNIQUE (
        recipient_pet_id,
        source_pet_hourly_log_id,
        propagated_group_master_id
    )
);

CREATE INDEX IF NOT EXISTS idx_pet_interest_propagations_recipient_occurred_at ON pet_interest_propagations(recipient_pet_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_pet_interest_propagations_source_hourly_log_id ON pet_interest_propagations(source_pet_hourly_log_id);

-- +goose Down
SELECT
    1;