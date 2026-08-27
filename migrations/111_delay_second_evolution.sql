-- +goose Up

-- あまえ期からなまい期への進化には、前回進化から11日間の経過を必須とする。
-- 分岐ごとにステージIDは異なるため、stage_no を使って対象ルールを限定する。
UPDATE evolution_rules AS rule
SET
    required_days_since_last_evolution = 11,
    updated_at = CURRENT_TIMESTAMP
FROM evolution_stages AS from_stage,
     evolution_stages AS to_stage
WHERE rule.from_stage_id = from_stage.id
  AND rule.to_stage_id = to_stage.id
  AND from_stage.stage_no = 1
  AND to_stage.stage_no = 2;

-- +goose Down

UPDATE evolution_rules AS rule
SET
    required_days_since_last_evolution = 1,
    updated_at = CURRENT_TIMESTAMP
FROM evolution_stages AS from_stage,
     evolution_stages AS to_stage
WHERE rule.from_stage_id = from_stage.id
  AND rule.to_stage_id = to_stage.id
  AND from_stage.stage_no = 1
  AND to_stage.stage_no = 2;
