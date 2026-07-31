-- +goose Up
ALTER TABLE reports ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id);

UPDATE reports AS report
SET user_id = pet.user_id
FROM pets AS pet
WHERE pet.id = report.pet_id
  AND report.user_id IS NULL;

ALTER TABLE reports ALTER COLUMN user_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_reports_user_pet_created_at
    ON reports(user_id, pet_id, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_reports_user_pet_created_at;
ALTER TABLE reports DROP COLUMN IF EXISTS user_id;
