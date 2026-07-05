INSERT INTO experience_caps (
    id,
    cap_type,
    max_experience,
    active,
    created_at,
    updated_at
)
VALUES
(
    '44444444-4444-4444-4444-444444444441',
    'daily',
    100,
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
),
(
    '44444444-4444-4444-4444-444444444442',
    'weekly',
    500,
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO UPDATE
SET
    cap_type = EXCLUDED.cap_type,
    max_experience = EXCLUDED.max_experience,
    active = EXCLUDED.active,
    updated_at = CURRENT_TIMESTAMP;