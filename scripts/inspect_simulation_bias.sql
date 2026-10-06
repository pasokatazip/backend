-- 自動テスト20匹の直近7日間を対象に、群れとステータス変化の偏りを確認する。
-- pgwebで実行し、アルゴリズム変更前後の同じ指標を比較する。

-- 群れマスタ自体のdelta分布。平均が0から大きく離れていないか確認する。
WITH deltas AS (
    SELECT status_name, delta
    FROM group_masters
    CROSS JOIN LATERAL (
        VALUES
            ('energy', energy_delta),
            ('curiosity', curiosity_delta),
            ('sociality', sociality_delta),
            ('routine', routine_delta)
    ) AS status(status_name, delta)
    WHERE active = TRUE
)
SELECT
    status_name,
    COUNT(*) FILTER (WHERE delta < 0) AS negative_groups,
    COUNT(*) FILTER (WHERE delta = 0) AS neutral_groups,
    COUNT(*) FILTER (WHERE delta > 0) AS positive_groups,
    ROUND(AVG(delta), 4) AS average_delta,
    MIN(delta) AS minimum_delta,
    MAX(delta) AS maximum_delta
FROM deltas
GROUP BY status_name
ORDER BY status_name;

WITH cohort AS (
    SELECT (
        '77100000-0000-4000-8000-' || LPAD(number::TEXT, 12, '0')
    )::UUID AS pet_id
    FROM generate_series(1, 20) AS number
),
logs AS (
    SELECT hourly_log.*
    FROM pet_hourly_logs AS hourly_log
    INNER JOIN cohort ON cohort.pet_id = hourly_log.pet_id
    WHERE hourly_log.simulated_at >= CURRENT_TIMESTAMP - INTERVAL '7 days'
      AND hourly_log.simulated_at < CURRENT_TIMESTAMP
),
group_counts AS (
    SELECT
        group_master_id,
        COUNT(*) AS observations,
        COUNT(DISTINCT pet_id) AS pets
    FROM logs
    GROUP BY group_master_id
),
group_shares AS (
    SELECT
        group_counts.*,
        observations::NUMERIC / SUM(observations) OVER () AS share
    FROM group_counts
)
SELECT
    group_master.group_key,
    group_master.category,
    group_shares.observations,
    group_shares.pets,
    ROUND(group_shares.share, 4) AS share
FROM group_shares
INNER JOIN group_masters AS group_master ON group_master.id = group_shares.group_master_id
ORDER BY group_shares.observations DESC, group_master.group_key;

WITH cohort AS (
    SELECT (
        '77100000-0000-4000-8000-' || LPAD(number::TEXT, 12, '0')
    )::UUID AS pet_id
    FROM generate_series(1, 20) AS number
),
logs AS (
    SELECT hourly_log.*
    FROM pet_hourly_logs AS hourly_log
    INNER JOIN cohort ON cohort.pet_id = hourly_log.pet_id
    WHERE hourly_log.simulated_at >= CURRENT_TIMESTAMP - INTERVAL '7 days'
      AND hourly_log.simulated_at < CURRENT_TIMESTAMP
),
current_status AS (
    SELECT pet.*
    FROM pets AS pet
    INNER JOIN cohort ON cohort.pet_id = pet.id
)
SELECT
    status.name,
    status.net_delta,
    status.positive_hours,
    status.negative_hours,
    status.average_current,
    status.minimum_current,
    status.maximum_current
FROM (
    SELECT
        'energy' AS name,
        SUM(energy_delta_applied) AS net_delta,
        COUNT(*) FILTER (WHERE energy_delta_applied > 0) AS positive_hours,
        COUNT(*) FILTER (WHERE energy_delta_applied < 0) AS negative_hours,
        (SELECT AVG(energy) FROM current_status) AS average_current,
        (SELECT MIN(energy) FROM current_status) AS minimum_current,
        (SELECT MAX(energy) FROM current_status) AS maximum_current
    FROM logs
    UNION ALL
    SELECT
        'curiosity',
        SUM(curiosity_delta_applied),
        COUNT(*) FILTER (WHERE curiosity_delta_applied > 0),
        COUNT(*) FILTER (WHERE curiosity_delta_applied < 0),
        (SELECT AVG(curiosity) FROM current_status),
        (SELECT MIN(curiosity) FROM current_status),
        (SELECT MAX(curiosity) FROM current_status)
    FROM logs
    UNION ALL
    SELECT
        'sociality',
        SUM(sociality_delta_applied),
        COUNT(*) FILTER (WHERE sociality_delta_applied > 0),
        COUNT(*) FILTER (WHERE sociality_delta_applied < 0),
        (SELECT AVG(sociality) FROM current_status),
        (SELECT MIN(sociality) FROM current_status),
        (SELECT MAX(sociality) FROM current_status)
    FROM logs
    UNION ALL
    SELECT
        'routine',
        SUM(routine_delta_applied),
        COUNT(*) FILTER (WHERE routine_delta_applied > 0),
        COUNT(*) FILTER (WHERE routine_delta_applied < 0),
        (SELECT AVG(routine) FROM current_status),
        (SELECT MIN(routine) FROM current_status),
        (SELECT MAX(routine) FROM current_status)
    FROM logs
) AS status
ORDER BY status.name;
