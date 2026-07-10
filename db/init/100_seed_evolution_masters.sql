-- Evolution stage and rule master data.
-- あかご期は固定。あまえ期で3種類に分岐し、同じ種類のなまい期へ進化する。
INSERT INTO
    evolution_stages (
        id,
        stage_key,
        stage_no,
        name,
        branch_key,
        image_url,
        created_at,
        updated_at
    )
VALUES
    (
        0,
        'akago',
        0,
        'あかご期',
        NULL,
        'https://assets.akatukii.com/assets/pets/akago.png',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        1,
        'amae_shokushu',
        1,
        'あまえ期・しょくしゅ',
        'shokushu',
        'https://assets.akatukii.com/assets/pets/amae_shokushu.png',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        11,
        'amae_yonshoku',
        1,
        'あまえ期・よんしょく',
        'yonshoku',
        'https://assets.akatukii.com/assets/pets/amae_yonshoku.png',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        12,
        'amae_nishoku',
        1,
        'あまえ期・にしょく',
        'nishoku',
        'https://assets.akatukii.com/assets/pets/amae_nishoku.png',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        2,
        'namai_shokushu',
        2,
        'なまい期・しょくしゅ',
        'shokushu',
        'https://assets.akatukii.com/assets/pets/namai_shokushu.png',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        21,
        'namai_yonshoku',
        2,
        'なまい期・よんしょく',
        'yonshoku',
        'https://assets.akatukii.com/assets/pets/namai_yonshoku.png',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        22,
        'namai_nishoku',
        2,
        'なまい期・にしょく',
        'nishoku',
        'https://assets.akatukii.com/assets/pets/namai_nishoku.png',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ) ON CONFLICT (id) DO
UPDATE
SET
    stage_key = EXCLUDED.stage_key,
    stage_no = EXCLUDED.stage_no,
    name = EXCLUDED.name,
    branch_key = EXCLUDED.branch_key,
    image_url = EXCLUDED.image_url,
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
        0,
        11,
        30,
        0,
        3,
        NULL,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        0,
        12,
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
    ),
    (
        11,
        21,
        100,
        1,
        10,
        NULL,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        12,
        22,
        100,
        1,
        10,
        NULL,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ) ON CONFLICT (from_stage_id, to_stage_id) DO
UPDATE
SET
    required_experience = EXCLUDED.required_experience,
    required_days_since_last_evolution = EXCLUDED.required_days_since_last_evolution,
    required_feed_count = EXCLUDED.required_feed_count,
    appearance_part = EXCLUDED.appearance_part,
    updated_at = CURRENT_TIMESTAMP;