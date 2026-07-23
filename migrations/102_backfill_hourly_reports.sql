-- +goose Up
-- reports は毎時の行動ログを画面表示向けに整形した記録として扱う。
-- 既存ログも補完し、以後は (pet_id, created_at) で重複を防止する。
CREATE UNIQUE INDEX IF NOT EXISTS uq_reports_pet_created_at ON reports(pet_id, created_at);

INSERT INTO
    reports (
        id,
        pet_id,
        hour_slot,
        gossip,
        group_master_id,
        previous_group_master_id,
        moved,
        behavior_type,
        behavior_label,
        interaction_count,
        energy_delta,
        curiosity_delta,
        sociality_delta,
        routine_delta,
        reason_json,
        created_at
    )
SELECT
    -- 拡張機能に依存せず、毎時ログから再現可能なUUIDを生成する。
    md5(
        hourly_log.pet_id :: TEXT || ':' || hourly_log.simulated_at :: TEXT
    ) :: UUID,
    hourly_log.pet_id,
    EXTRACT(
        HOUR
        FROM
            hourly_log.simulated_at AT TIME ZONE 'Asia/Tokyo'
    ) :: INTEGER,
    LEFT(NULLIF(hourly_log.report_material, ''), 255),
    hourly_log.group_master_id,
    NULL,
    NOT hourly_log.stayed,
    CASE
        WHEN hourly_log.stayed THEN 'stayed'
        ELSE 'moved'
    END,
    LEFT(
        COALESCE(
            NULLIF(hourly_log.ambient_event, ''),
            '群れでゆっくり過ごした'
        ),
        255
    ),
    hourly_log.interaction_count,
    ROUND(hourly_log.energy_delta_applied) :: INTEGER,
    ROUND(hourly_log.curiosity_delta_applied) :: INTEGER,
    ROUND(hourly_log.sociality_delta_applied) :: INTEGER,
    ROUND(hourly_log.routine_delta_applied) :: INTEGER,
    jsonb_build_object(
        'source',
        'hourly_simulation',
        'simulated_at',
        hourly_log.simulated_at
    ),
    hourly_log.simulated_at
FROM
    pet_hourly_logs AS hourly_log ON CONFLICT (pet_id, created_at) DO NOTHING;

-- +goose Down
DROP INDEX IF EXISTS uq_reports_pet_created_at;