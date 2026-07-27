-- +goose Up
-- 頭痛の群れと、その群れに紐づく履歴・おみやげ・レポートを削除する。
-- 現在頭痛の群れにいるペットだけは、削除せず疲れの群れへ移す。

-- レポートを削除する前に、旅立ち履歴の参照を外す。
UPDATE pet_departures
SET farewell_report_id = NULL,
    updated_at = CURRENT_TIMESTAMP
WHERE farewell_report_id IN (
    SELECT reports.id
    FROM reports
    INNER JOIN group_masters
        ON group_masters.id = reports.group_master_id
        OR group_masters.id = reports.previous_group_master_id
    WHERE group_masters.group_key = 'headache'
);

-- おみやげは群れ・レポート・毎時ログのいずれかが頭痛の群れに関係するものを削除する。
DELETE FROM pet_souvenirs
WHERE souvenir_master_id IN (
        SELECT souvenir_masters.id
        FROM souvenir_masters
        INNER JOIN group_masters ON group_masters.id = souvenir_masters.group_master_id
        WHERE group_masters.group_key = 'headache'
    )
    OR source_group_master_id IN (
        SELECT id FROM group_masters WHERE group_key = 'headache'
    )
    OR report_id IN (
        SELECT reports.id
        FROM reports
        INNER JOIN group_masters
            ON group_masters.id = reports.group_master_id
            OR group_masters.id = reports.previous_group_master_id
        WHERE group_masters.group_key = 'headache'
    )
    OR pet_hourly_log_id IN (
        SELECT pet_hourly_logs.id
        FROM pet_hourly_logs
        WHERE pet_hourly_logs.group_master_id IN (
                SELECT id FROM group_masters WHERE group_key = 'headache'
            )
            OR pet_hourly_logs.pet_group_join_id IN (
                SELECT id
                FROM pet_group_joins
                WHERE group_master_id IN (
                    SELECT id FROM group_masters WHERE group_key = 'headache'
                )
            )
    );

-- 削除する毎時ログを参照する経験値イベントも残さない。
DELETE FROM pet_experience_events
WHERE source_type = 'hourly_log'
    AND source_id IN (
        SELECT pet_hourly_logs.id
        FROM pet_hourly_logs
        WHERE pet_hourly_logs.group_master_id IN (
                SELECT id FROM group_masters WHERE group_key = 'headache'
            )
            OR pet_hourly_logs.pet_group_join_id IN (
                SELECT id
                FROM pet_group_joins
                WHERE group_master_id IN (
                    SELECT id FROM group_masters WHERE group_key = 'headache'
                )
            )
    );

DELETE FROM pet_interest_propagations
WHERE via_group_master_id IN (
        SELECT id FROM group_masters WHERE group_key = 'headache'
    )
    OR propagated_group_master_id IN (
        SELECT id FROM group_masters WHERE group_key = 'headache'
    )
    OR source_pet_hourly_log_id IN (
        SELECT pet_hourly_logs.id
        FROM pet_hourly_logs
        WHERE pet_hourly_logs.group_master_id IN (
                SELECT id FROM group_masters WHERE group_key = 'headache'
            )
            OR pet_hourly_logs.pet_group_join_id IN (
                SELECT id
                FROM pet_group_joins
                WHERE group_master_id IN (
                    SELECT id FROM group_masters WHERE group_key = 'headache'
                )
            )
    );

DELETE FROM reports
WHERE group_master_id IN (
        SELECT id FROM group_masters WHERE group_key = 'headache'
    )
    OR previous_group_master_id IN (
        SELECT id FROM group_masters WHERE group_key = 'headache'
    );

DELETE FROM pet_hourly_logs
WHERE group_master_id IN (
        SELECT id FROM group_masters WHERE group_key = 'headache'
    )
    OR pet_group_join_id IN (
        SELECT id
        FROM pet_group_joins
        WHERE group_master_id IN (
            SELECT id FROM group_masters WHERE group_key = 'headache'
        )
    );

DELETE FROM pet_group_joins
WHERE group_master_id IN (
    SELECT id FROM group_masters WHERE group_key = 'headache'
);

-- ペット自体は保持し、削除対象の群れを参照しない状態にする。
UPDATE pets
SET current_group_master_id = (
        SELECT id FROM group_masters WHERE group_key = 'fatigue'
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE current_group_master_id IN (
    SELECT id FROM group_masters WHERE group_key = 'headache'
);

-- group_keywords / noun_group_matches / pet_group_interests / souvenir_masters は
-- group_masters への ON DELETE CASCADE により同時に削除される。
DELETE FROM group_masters
WHERE group_key = 'headache';

-- +goose Down
-- 削除済みの履歴やマスタは復元できない。
SELECT 1;
