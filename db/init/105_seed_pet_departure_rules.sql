-- Default pet departure rule.
-- ペットは最低30日過ごし、ステージ2到達から7日経っていた場合に旅立つ。
INSERT INTO
    pet_departure_rules (
        rule_key,
        min_age_days,
        required_stage_id,
        grace_days_min,
        grace_days_max,
        active,
        created_at,
        updated_at
    )
VALUES
    (
        'default_adult_departure',
        30,
        2,
        7,
        7,
        TRUE,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ) ON CONFLICT (rule_key) DO
UPDATE
SET
    min_age_days = EXCLUDED.min_age_days,
    required_stage_id = EXCLUDED.required_stage_id,
    grace_days_min = EXCLUDED.grace_days_min,
    grace_days_max = EXCLUDED.grace_days_max,
    active = EXCLUDED.active,
    updated_at = CURRENT_TIMESTAMP;
