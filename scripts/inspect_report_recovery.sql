-- 本番DBで実行できる読み取り専用の欠損調査。
-- 対象は9月3日18時以降、現在の正時より前まで。必要なら開始・終了を変更する。
BEGIN READ ONLY;
SET LOCAL statement_timeout = '30s';
WITH recovery_window AS (
    SELECT TIMESTAMPTZ '2026-09-03 18:00:00+09' AS started_at,
           date_trunc('hour', CURRENT_TIMESTAMP) AS ended_at
), hours AS (
    SELECT generate_series(started_at, ended_at - INTERVAL '1 hour', INTERVAL '1 hour') AS at
    FROM recovery_window
), logs AS (
    SELECT l.simulated_at AS at, COUNT(*) AS log_count,
           COUNT(*) FILTER (WHERE r.id IS NULL) AS recoverable_reports
    FROM pet_hourly_logs l
    JOIN pets p ON p.id = l.pet_id
    LEFT JOIN reports r ON r.pet_id = l.pet_id AND r.created_at = l.simulated_at
    CROSS JOIN recovery_window w
    WHERE l.simulated_at >= w.started_at AND l.simulated_at < w.ended_at
    GROUP BY l.simulated_at
), report_counts AS (
    SELECT date_trunc('hour', r.created_at) AS at, COUNT(*) AS report_count
    FROM reports r CROSS JOIN recovery_window w
    WHERE r.created_at >= w.started_at AND r.created_at < w.ended_at
    GROUP BY date_trunc('hour', r.created_at)
)
SELECT h.at AT TIME ZONE 'Asia/Tokyo' AS hour_jst,
       COALESCE(l.log_count, 0) AS hourly_logs,
       COALESCE(r.report_count, 0) AS reports,
       COALESCE(l.recoverable_reports, 0) AS recoverable_reports
FROM hours h
LEFT JOIN logs l ON l.at = h.at
LEFT JOIN report_counts r ON r.at = h.at
ORDER BY h.at;
COMMIT;
