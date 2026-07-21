-- +goose Up
ALTER TABLE pets
ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'active';

ALTER TABLE pets
ADD CONSTRAINT chk_pets_status CHECK (
    status IN ('active', 'departed', 'lost', 'archived')
);

CREATE INDEX IF NOT EXISTS idx_pets_status ON pets(status);

-- +goose Down
DROP INDEX IF EXISTS idx_pets_status;
ALTER TABLE pets DROP CONSTRAINT IF EXISTS chk_pets_status;
ALTER TABLE pets DROP COLUMN IF EXISTS status;
