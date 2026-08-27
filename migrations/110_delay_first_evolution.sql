-- +goose Up

-- アカウント作成直後の連続投稿だけでは進化しないよう、
-- あかご期からの初回進化に7日間の経過を必須とする。
UPDATE evolution_rules
SET
    required_days_since_last_evolution = 7,
    updated_at = CURRENT_TIMESTAMP
WHERE from_stage_id = 0;

-- +goose Down

UPDATE evolution_rules
SET
    required_days_since_last_evolution = 0,
    updated_at = CURRENT_TIMESTAMP
WHERE from_stage_id = 0;
