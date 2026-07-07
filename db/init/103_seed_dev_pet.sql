-- Development seed user and pet.
-- Login email: dev@example.com
-- Login password: password
INSERT INTO
    users (
        id,
        email,
        password,
        subsc,
        created_at
    )
VALUES
    (
        '00000000-0000-0000-0000-000000000001',
        'dev@example.com',
        '$2y$12$EogK9gUGFLC/i9EbUDuSGu.jaVrg4E71sVbMt8fMLqrDV/CVYyWYe',
        TRUE,
        CURRENT_TIMESTAMP
    ) ON CONFLICT (id) DO
UPDATE
SET
    email = EXCLUDED.email,
    password = EXCLUDED.password,
    subsc = EXCLUDED.subsc;

INSERT INTO
    pets (
        id,
        name,
        color,
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
VALUES
    (
        '11111111-1111-1111-1111-111111111111',
        'テストペット',
        '#FFC1CA',
        FALSE,
        '00000000-0000-0000-0000-000000000001',
        50,
        50,
        50,
        50,
        (
            SELECT
                id
            FROM
                group_masters
            WHERE
                group_key = 'app'
        ),
        0,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ) ON CONFLICT (id) DO
UPDATE
SET
    name = EXCLUDED.name,
    color = EXCLUDED.color,
    is_deleted = EXCLUDED.is_deleted,
    user_id = EXCLUDED.user_id,
    energy = EXCLUDED.energy,
    curiosity = EXCLUDED.curiosity,
    sociality = EXCLUDED.sociality,
    routine = EXCLUDED.routine,
    current_group_master_id = EXCLUDED.current_group_master_id,
    current_stage_id = EXCLUDED.current_stage_id,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO
    user_active_pets (user_id, pet_id, assigned_at)
VALUES
    (
        '00000000-0000-0000-0000-000000000001',
        '11111111-1111-1111-1111-111111111111',
        CURRENT_TIMESTAMP
    ) ON CONFLICT (user_id, pet_id) DO
UPDATE
SET
    assigned_at = EXCLUDED.assigned_at;

INSERT INTO
    pet_experiences (
        id,
        pet_id,
        total_experience,
        feed_count,
        created_at,
        updated_at
    )
VALUES
    (
        '22222222-2222-2222-2222-222222222222',
        '11111111-1111-1111-1111-111111111111',
        0,
        0,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ) ON CONFLICT (id) DO
UPDATE
SET
    pet_id = EXCLUDED.pet_id,
    total_experience = EXCLUDED.total_experience,
    feed_count = EXCLUDED.feed_count,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO
    pet_group_joins (
        id,
        pet_id,
        group_master_id,
        joined_at,
        left_at,
        move_reason,
        created_at,
        updated_at
    )
VALUES
    (
        '33333333-3333-3333-3333-333333333333',
        '11111111-1111-1111-1111-111111111111',
        (
            SELECT
                id
            FROM
                group_masters
            WHERE
                group_key = 'app'
        ),
        CURRENT_TIMESTAMP,
        NULL,
        'initial',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ) ON CONFLICT (id) DO NOTHING;

-- Hourly simulation test users and pets.
-- Login password for all users: password
-- These pets cover low/high status combinations, recent moves, and long stays.
WITH seed_hourly_users (id, email) AS (
    VALUES
    (
        '00000000-0000-0000-0000-000000000101' :: UUID,
        'hourly.low-energy@example.com'
    ),
    (
        '00000000-0000-0000-0000-000000000102' :: UUID,
        'hourly.high-curiosity@example.com'
    ),
    (
        '00000000-0000-0000-0000-000000000103' :: UUID,
        'hourly.social@example.com'
    ),
    (
        '00000000-0000-0000-0000-000000000104' :: UUID,
        'hourly.late-night@example.com'
    ),
    (
        '00000000-0000-0000-0000-000000000105' :: UUID,
        'hourly.energetic@example.com'
    ),
    (
        '00000000-0000-0000-0000-000000000106' :: UUID,
        'hourly.balanced@example.com'
    ),
    (
        '00000000-0000-0000-0000-000000000107' :: UUID,
        'hourly.recent-move@example.com'
    ),
    (
        '00000000-0000-0000-0000-000000000108' :: UUID,
        'hourly.long-stay@example.com'
    )
)
INSERT INTO
    users (
        id,
        email,
        password,
        subsc,
        created_at
    )
SELECT
    id,
    email,
    '$2y$12$EogK9gUGFLC/i9EbUDuSGu.jaVrg4E71sVbMt8fMLqrDV/CVYyWYe',
    TRUE,
    CURRENT_TIMESTAMP
FROM
    seed_hourly_users ON CONFLICT (id) DO
UPDATE
SET
    email = EXCLUDED.email,
    password = EXCLUDED.password,
    subsc = EXCLUDED.subsc;

WITH seed_hourly_pets (
    id,
    user_id,
    name,
    color,
    energy,
    curiosity,
    sociality,
    routine,
    group_key,
    current_stage_id,
    created_at,
    joined_at
) AS (
    VALUES
    -- rest_need が高く、睡眠・休息系へ寄りやすい
    (
        '11111111-1111-1111-1111-111111111101' :: UUID,
        '00000000-0000-0000-0000-000000000101' :: UUID,
        'ねむテスト',
        '#9AD7FF',
        8.0000,
        35.0000,
        30.0000,
        42.0000,
        'work',
        0,
        CURRENT_TIMESTAMP - INTERVAL '12 days',
        CURRENT_TIMESTAMP - INTERVAL '9 hours'
    ),
    -- curiosity補正が高く、新しい群れへ移動しやすい
    (
        '11111111-1111-1111-1111-111111111102' :: UUID,
        '00000000-0000-0000-0000-000000000102' :: UUID,
        'きょろテスト',
        '#B8F7C5',
        72.0000,
        95.0000,
        35.0000,
        55.0000,
        'home',
        1,
        CURRENT_TIMESTAMP - INTERVAL '18 days',
        CURRENT_TIMESTAMP - INTERVAL '6 hours'
    ),
    -- sociality が高く、交流数と群れへの愛着が出やすい
    (
        '11111111-1111-1111-1111-111111111103' :: UUID,
        '00000000-0000-0000-0000-000000000103' :: UUID,
        'わいわいテスト',
        '#FFD166',
        58.0000,
        55.0000,
        96.0000,
        60.0000,
        'chat',
        1,
        CURRENT_TIMESTAMP - INTERVAL '20 days',
        CURRENT_TIMESTAMP - INTERVAL '14 hours'
    ),
    -- routine が低く、夜ふかし・休息寄りの挙動を見やすい
    (
        '11111111-1111-1111-1111-111111111104' :: UUID,
        '00000000-0000-0000-0000-000000000104' :: UUID,
        'よふかしテスト',
        '#CDB4DB',
        38.0000,
        72.0000,
        45.0000,
        8.0000,
        'late_night',
        1,
        CURRENT_TIMESTAMP - INTERVAL '22 days',
        CURRENT_TIMESTAMP - INTERVAL '11 hours'
    ),
    -- energy が高く、運動系のdelta適用後も上限付近を確認しやすい
    (
        '11111111-1111-1111-1111-111111111105' :: UUID,
        '00000000-0000-0000-0000-000000000105' :: UUID,
        'げんきテスト',
        '#80ED99',
        97.0000,
        64.0000,
        68.0000,
        82.0000,
        'exercise',
        2,
        CURRENT_TIMESTAMP - INTERVAL '35 days',
        CURRENT_TIMESTAMP - INTERVAL '5 hours'
    ),
    -- 中央値付近の基準ケース
    (
        '11111111-1111-1111-1111-111111111106' :: UUID,
        '00000000-0000-0000-0000-000000000106' :: UUID,
        'まんなかテスト',
        '#FFC1CA',
        50.0000,
        50.0000,
        50.0000,
        50.0000,
        'app',
        0,
        CURRENT_TIMESTAMP - INTERVAL '7 days',
        CURRENT_TIMESTAMP - INTERVAL '4 hours'
    ),
    -- 直近移動ペナルティを確認するケース
    (
        '11111111-1111-1111-1111-111111111107' :: UUID,
        '00000000-0000-0000-0000-000000000107' :: UUID,
        'さっきテスト',
        '#FFAFCC',
        45.0000,
        62.0000,
        42.0000,
        48.0000,
        'cafe',
        0,
        CURRENT_TIMESTAMP - INTERVAL '5 days',
        CURRENT_TIMESTAMP - INTERVAL '1 hour'
    ),
    -- 長時間滞在で boredom が高くなるケース
    (
        '11111111-1111-1111-1111-111111111108' :: UUID,
        '00000000-0000-0000-0000-000000000108' :: UUID,
        'ながいテスト',
        '#A0C4FF',
        66.0000,
        44.0000,
        72.0000,
        74.0000,
        'reading',
        2,
        CURRENT_TIMESTAMP - INTERVAL '40 days',
        CURRENT_TIMESTAMP - INTERVAL '48 hours'
    )
)
INSERT INTO
    pets (
        id,
        name,
        color,
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
SELECT
    seed.id,
    seed.name,
    seed.color,
    FALSE,
    seed.user_id,
    seed.energy,
    seed.curiosity,
    seed.sociality,
    seed.routine,
    gm.id,
    seed.current_stage_id,
    seed.created_at,
    CURRENT_TIMESTAMP
FROM
    seed_hourly_pets seed
    INNER JOIN group_masters gm ON gm.group_key = seed.group_key ON CONFLICT (id) DO
UPDATE
SET
    name = EXCLUDED.name,
    color = EXCLUDED.color,
    is_deleted = EXCLUDED.is_deleted,
    user_id = EXCLUDED.user_id,
    energy = EXCLUDED.energy,
    curiosity = EXCLUDED.curiosity,
    sociality = EXCLUDED.sociality,
    routine = EXCLUDED.routine,
    current_group_master_id = EXCLUDED.current_group_master_id,
    current_stage_id = EXCLUDED.current_stage_id,
    updated_at = CURRENT_TIMESTAMP;

DO $$
BEGIN
    -- 古い開発DBには pets.status がない場合があるため、列がある時だけ active に戻す。
    IF EXISTS (
        SELECT
            1
        FROM
            information_schema.columns
        WHERE
            table_name = 'pets'
            AND column_name = 'status'
    ) THEN
        EXECUTE $sql$
            UPDATE pets
            SET status = 'active'
            WHERE id IN (
                '11111111-1111-1111-1111-111111111101' :: UUID,
                '11111111-1111-1111-1111-111111111102' :: UUID,
                '11111111-1111-1111-1111-111111111103' :: UUID,
                '11111111-1111-1111-1111-111111111104' :: UUID,
                '11111111-1111-1111-1111-111111111105' :: UUID,
                '11111111-1111-1111-1111-111111111106' :: UUID,
                '11111111-1111-1111-1111-111111111107' :: UUID,
                '11111111-1111-1111-1111-111111111108' :: UUID
            )
        $sql$;
    END IF;
END $$;

WITH seed_hourly_active_pets (user_id, pet_id) AS (
    VALUES
    (
        '00000000-0000-0000-0000-000000000101' :: UUID,
        '11111111-1111-1111-1111-111111111101' :: UUID
    ),
    (
        '00000000-0000-0000-0000-000000000102' :: UUID,
        '11111111-1111-1111-1111-111111111102' :: UUID
    ),
    (
        '00000000-0000-0000-0000-000000000103' :: UUID,
        '11111111-1111-1111-1111-111111111103' :: UUID
    ),
    (
        '00000000-0000-0000-0000-000000000104' :: UUID,
        '11111111-1111-1111-1111-111111111104' :: UUID
    ),
    (
        '00000000-0000-0000-0000-000000000105' :: UUID,
        '11111111-1111-1111-1111-111111111105' :: UUID
    ),
    (
        '00000000-0000-0000-0000-000000000106' :: UUID,
        '11111111-1111-1111-1111-111111111106' :: UUID
    ),
    (
        '00000000-0000-0000-0000-000000000107' :: UUID,
        '11111111-1111-1111-1111-111111111107' :: UUID
    ),
    (
        '00000000-0000-0000-0000-000000000108' :: UUID,
        '11111111-1111-1111-1111-111111111108' :: UUID
    )
)
INSERT INTO
    user_active_pets (user_id, pet_id, assigned_at)
SELECT
    user_id,
    pet_id,
    CURRENT_TIMESTAMP
FROM
    seed_hourly_active_pets ON CONFLICT (user_id, pet_id) DO
UPDATE
SET
    assigned_at = EXCLUDED.assigned_at;

WITH seed_hourly_experiences (id, pet_id, total_experience, feed_count) AS (
    VALUES
    (
        '22222222-2222-2222-2222-222222222101' :: UUID,
        '11111111-1111-1111-1111-111111111101' :: UUID,
        5,
        1
    ),
    (
        '22222222-2222-2222-2222-222222222102' :: UUID,
        '11111111-1111-1111-1111-111111111102' :: UUID,
        35,
        4
    ),
    (
        '22222222-2222-2222-2222-222222222103' :: UUID,
        '11111111-1111-1111-1111-111111111103' :: UUID,
        60,
        7
    ),
    (
        '22222222-2222-2222-2222-222222222104' :: UUID,
        '11111111-1111-1111-1111-111111111104' :: UUID,
        80,
        9
    ),
    (
        '22222222-2222-2222-2222-222222222105' :: UUID,
        '11111111-1111-1111-1111-111111111105' :: UUID,
        130,
        12
    ),
    (
        '22222222-2222-2222-2222-222222222106' :: UUID,
        '11111111-1111-1111-1111-111111111106' :: UUID,
        0,
        0
    ),
    (
        '22222222-2222-2222-2222-222222222107' :: UUID,
        '11111111-1111-1111-1111-111111111107' :: UUID,
        15,
        2
    ),
    (
        '22222222-2222-2222-2222-222222222108' :: UUID,
        '11111111-1111-1111-1111-111111111108' :: UUID,
        160,
        16
    )
)
INSERT INTO
    pet_experiences (
        id,
        pet_id,
        total_experience,
        feed_count,
        created_at,
        updated_at
    )
SELECT
    id,
    pet_id,
    total_experience,
    feed_count,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM
    seed_hourly_experiences ON CONFLICT (pet_id) DO
UPDATE
SET
    total_experience = EXCLUDED.total_experience,
    feed_count = EXCLUDED.feed_count,
    updated_at = CURRENT_TIMESTAMP;

WITH seed_hourly_pet_ids (pet_id) AS (
    VALUES
    ('11111111-1111-1111-1111-111111111101' :: UUID),
    ('11111111-1111-1111-1111-111111111102' :: UUID),
    ('11111111-1111-1111-1111-111111111103' :: UUID),
    ('11111111-1111-1111-1111-111111111104' :: UUID),
    ('11111111-1111-1111-1111-111111111105' :: UUID),
    ('11111111-1111-1111-1111-111111111106' :: UUID),
    ('11111111-1111-1111-1111-111111111107' :: UUID),
    ('11111111-1111-1111-1111-111111111108' :: UUID)
),
seed_hourly_join_ids (join_id) AS (
    VALUES
    ('33333333-3333-3333-3333-333333333101' :: UUID),
    ('33333333-3333-3333-3333-333333333102' :: UUID),
    ('33333333-3333-3333-3333-333333333103' :: UUID),
    ('33333333-3333-3333-3333-333333333104' :: UUID),
    ('33333333-3333-3333-3333-333333333105' :: UUID),
    ('33333333-3333-3333-3333-333333333106' :: UUID),
    ('33333333-3333-3333-3333-333333333107' :: UUID),
    ('33333333-3333-3333-3333-333333333108' :: UUID)
)
UPDATE
    pet_group_joins
SET
    left_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE
    pet_id IN (
        SELECT
            pet_id
        FROM
            seed_hourly_pet_ids
    )
    AND id NOT IN (
        SELECT
            join_id
        FROM
            seed_hourly_join_ids
    )
    AND left_at IS NULL;

WITH seed_hourly_joins (id, pet_id, group_key, joined_at) AS (
    VALUES
    (
        '33333333-3333-3333-3333-333333333101' :: UUID,
        '11111111-1111-1111-1111-111111111101' :: UUID,
        'work',
        CURRENT_TIMESTAMP - INTERVAL '9 hours'
    ),
    (
        '33333333-3333-3333-3333-333333333102' :: UUID,
        '11111111-1111-1111-1111-111111111102' :: UUID,
        'home',
        CURRENT_TIMESTAMP - INTERVAL '6 hours'
    ),
    (
        '33333333-3333-3333-3333-333333333103' :: UUID,
        '11111111-1111-1111-1111-111111111103' :: UUID,
        'chat',
        CURRENT_TIMESTAMP - INTERVAL '14 hours'
    ),
    (
        '33333333-3333-3333-3333-333333333104' :: UUID,
        '11111111-1111-1111-1111-111111111104' :: UUID,
        'late_night',
        CURRENT_TIMESTAMP - INTERVAL '11 hours'
    ),
    (
        '33333333-3333-3333-3333-333333333105' :: UUID,
        '11111111-1111-1111-1111-111111111105' :: UUID,
        'exercise',
        CURRENT_TIMESTAMP - INTERVAL '5 hours'
    ),
    (
        '33333333-3333-3333-3333-333333333106' :: UUID,
        '11111111-1111-1111-1111-111111111106' :: UUID,
        'app',
        CURRENT_TIMESTAMP - INTERVAL '4 hours'
    ),
    (
        '33333333-3333-3333-3333-333333333107' :: UUID,
        '11111111-1111-1111-1111-111111111107' :: UUID,
        'cafe',
        CURRENT_TIMESTAMP - INTERVAL '1 hour'
    ),
    (
        '33333333-3333-3333-3333-333333333108' :: UUID,
        '11111111-1111-1111-1111-111111111108' :: UUID,
        'reading',
        CURRENT_TIMESTAMP - INTERVAL '48 hours'
    )
)
INSERT INTO
    pet_group_joins (
        id,
        pet_id,
        group_master_id,
        joined_at,
        left_at,
        move_reason,
        created_at,
        updated_at
    )
SELECT
    seed.id,
    seed.pet_id,
    gm.id,
    seed.joined_at,
    NULL,
    'initial',
    seed.joined_at,
    CURRENT_TIMESTAMP
FROM
    seed_hourly_joins seed
    INNER JOIN group_masters gm ON gm.group_key = seed.group_key ON CONFLICT (id) DO
UPDATE
SET
    group_master_id = EXCLUDED.group_master_id,
    joined_at = EXCLUDED.joined_at,
    left_at = NULL,
    move_reason = EXCLUDED.move_reason,
    updated_at = CURRENT_TIMESTAMP;
