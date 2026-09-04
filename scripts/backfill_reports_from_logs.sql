-- inspect_report_recovery.sqlで保存済みログと補完件数を確認してから実行する。
-- 既存レポートは上書きせず、行動ログが残る時間だけを補完する。
-- 行動ログがない時間の行動・交流・おみやげは推測しない。
-- 通常のreports INSERTと同様に、DBのおみやげ紐付け・最低保証トリガーは実行される。
BEGIN;
SET LOCAL statement_timeout = '30s';
SET LOCAL lock_timeout = '5s';
SET LOCAL TIME ZONE 'UTC';

WITH recovery_window AS (
    -- 終了は排他的。実行中の毎時処理と重ならないよう、現在の正時より前までを対象にする。
    SELECT TIMESTAMPTZ '2026-09-03 18:00:00+09' AS started_at,
           date_trunc('hour', CURRENT_TIMESTAMP) AS ended_at
), inserted AS (
    INSERT INTO reports (
        id, user_id, pet_id, hour_slot, gossip, group_master_id, previous_group_master_id,
        moved, behavior_type, behavior_label, interaction_count,
        energy_delta, curiosity_delta, sociality_delta, routine_delta,
        reason_json, rumor, created_at
    )
    SELECT
        md5(l.pet_id::TEXT || ':' || l.simulated_at::TEXT)::UUID,
        p.user_id, l.pet_id,
        EXTRACT(HOUR FROM l.simulated_at AT TIME ZONE 'Asia/Tokyo')::INTEGER,
        LEFT(NULLIF(l.report_material, ''), 255),
        l.group_master_id, NULL, NOT l.stayed,
        CASE WHEN l.stayed THEN 'stayed' ELSE 'moved' END,
        LEFT(COALESCE(NULLIF(l.ambient_event, ''),
            CASE WHEN l.stayed THEN '群れでゆっくり過ごした' ELSE '別の群れへ移動した' END), 255),
        l.interaction_count,
        ROUND(l.energy_delta_applied)::INTEGER, ROUND(l.curiosity_delta_applied)::INTEGER,
        ROUND(l.sociality_delta_applied)::INTEGER, ROUND(l.routine_delta_applied)::INTEGER,
        jsonb_build_object('source', 'hourly_log_backfill', 'simulated_at', l.simulated_at),
        '[]'::JSONB, l.simulated_at
    FROM pet_hourly_logs l
    JOIN pets p ON p.id = l.pet_id
    CROSS JOIN recovery_window w
    WHERE l.simulated_at >= w.started_at AND l.simulated_at < w.ended_at
      AND NOT EXISTS (
          SELECT 1 FROM reports r WHERE r.pet_id = l.pet_id AND r.created_at = l.simulated_at
      )
    -- 同じペットのおみやげが時系列に沿って紐付くよう古い時間から処理する。
    ORDER BY l.pet_id, l.simulated_at
    ON CONFLICT (pet_id, created_at) DO NOTHING
    RETURNING created_at
)
SELECT created_at AT TIME ZONE 'Asia/Tokyo' AS hour_jst,
       COUNT(*) AS inserted_reports
FROM inserted
GROUP BY created_at
ORDER BY created_at;
COMMIT;
