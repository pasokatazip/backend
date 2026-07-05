

-- Evolution stage and rule master data.
-- 投稿を食べた回数と累計経験値で、段階的に進化できる最低限の設定

INSERT INTO
    evolution_stages (
        id,
        stage_no,
        name,
        description,
        created_at,
        updated_at
    )
VALUES
    (
        0,
        0,
        'たまご',
        'まだ世界の気配を集めはじめたばかりの段階。',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        1,
        1,
        'こども',
        '少しずつ群れの気配になじみはじめた段階。',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        2,
        2,
        'おとな',
        '日々の気配をたくさん持ち帰れるようになった段階。',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    )
ON CONFLICT (id) DO UPDATE
SET
    stage_no = EXCLUDED.stage_no,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO
    evolution_rules (
        from_stage_id,
        to_stage_id,
        required_experience,
        required_days_since_last_evolution,
        required_feed_count,
        appearance_part,
        created_at,
        updated_at
    )
VALUES
    (
        0,
        1,
        30,
        0,
        3,
        NULL,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        1,
        2,
        100,
        1,
        10,
        NULL,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    )
ON CONFLICT (from_stage_id, to_stage_id) DO UPDATE
SET
    required_experience = EXCLUDED.required_experience,
    required_days_since_last_evolution = EXCLUDED.required_days_since_last_evolution,
    required_feed_count = EXCLUDED.required_feed_count,
    appearance_part = EXCLUDED.appearance_part,
    updated_at = CURRENT_TIMESTAMP;