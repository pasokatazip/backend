-- Development seed user and pet.
-- Login email: dev@example.com
-- Login password: password

INSERT INTO users (
    id,
    email,
    password,
    subsc,
    created_at
)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'dev@example.com',
    '$2y$12$EogK9gUGFLC/i9EbUDuSGu.jaVrg4E71sVbMt8fMLqrDV/CVYyWYe',
    TRUE,
    CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO UPDATE
SET
    email = EXCLUDED.email,
    password = EXCLUDED.password,
    subsc = EXCLUDED.subsc;

INSERT INTO pets (
    id,
    name,
    is_deleted,
    user_id,
    energy,
    curiosity,
    sociality,
    routine,
    current_group_master_id,
    current_stage_id,
    created_at,
    updated_at
)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'テストペット',
    FALSE,
    '00000000-0000-0000-0000-000000000001',
    50.000,
    55.000,
    45.000,
    50.000,
    (
        SELECT id
        FROM group_masters
        WHERE group_key = 'app'
    ),
    0,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    is_deleted = EXCLUDED.is_deleted,
    user_id = EXCLUDED.user_id,
    energy = EXCLUDED.energy,
    curiosity = EXCLUDED.curiosity,
    sociality = EXCLUDED.sociality,
    routine = EXCLUDED.routine,
    current_group_master_id = EXCLUDED.current_group_master_id,
    current_stage_id = EXCLUDED.current_stage_id,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO user_active_pets (
    user_id,
    pet_id,
    assigned_at
)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '11111111-1111-1111-1111-111111111111',
    CURRENT_TIMESTAMP
)
ON CONFLICT (user_id, pet_id) DO UPDATE
SET assigned_at = EXCLUDED.assigned_at;

INSERT INTO pet_experiences (
    id,
    pet_id,
    total_experience,
    feed_count,
    created_at,
    updated_at
)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    0,
    0,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO UPDATE
SET
    pet_id = EXCLUDED.pet_id,
    total_experience = EXCLUDED.total_experience,
    feed_count = EXCLUDED.feed_count,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO pet_group_joins (
    id,
    pet_id,
    group_master_id,
    joined_at,
    left_at,
    move_reason,
    created_at,
    updated_at
)
VALUES (
    '33333333-3333-3333-3333-333333333333',
    '11111111-1111-1111-1111-111111111111',
    (
        SELECT id
        FROM group_masters
        WHERE group_key = 'app'
    ),
    CURRENT_TIMESTAMP,
    NULL,
    'initial',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO NOTHING;
