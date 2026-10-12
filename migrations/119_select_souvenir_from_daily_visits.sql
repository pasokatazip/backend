-- +goose Up

-- 旧トリガーはレポートの群れに有効なおみやげがない場合、
-- souvenir_masters.id が最も小さい群れを任意に選んでいた。
-- その日に実際に訪れた群れではない、レポート最低保証のおみやげを補正する。
WITH correction_candidates AS (
    SELECT
        pet_souvenir.id AS pet_souvenir_id,
        candidate.souvenir_master_id,
        candidate.group_master_id,
        candidate.group_display_name
    FROM pet_souvenirs AS pet_souvenir
    INNER JOIN reports AS report ON report.id = pet_souvenir.report_id
    INNER JOIN LATERAL (
        SELECT
            souvenir_master.id AS souvenir_master_id,
            souvenir_master.group_master_id,
            group_master.display_name AS group_display_name
        FROM pet_hourly_logs AS hourly_log
        INNER JOIN souvenir_masters AS souvenir_master
            ON souvenir_master.group_master_id = hourly_log.group_master_id
           AND souvenir_master.active = TRUE
           AND souvenir_master.display_name <> '謎のお土産'
        INNER JOIN group_masters AS group_master
            ON group_master.id = hourly_log.group_master_id
        WHERE hourly_log.pet_id = pet_souvenir.pet_id
          AND (hourly_log.simulated_at AT TIME ZONE 'Asia/Tokyo')::DATE = pet_souvenir.found_on
        ORDER BY
            CASE WHEN hourly_log.group_master_id = report.group_master_id THEN 0 ELSE 1 END,
            hourly_log.simulated_at DESC,
            hourly_log.id DESC
        LIMIT 1
    ) AS candidate ON TRUE
    WHERE pet_souvenir.pet_hourly_log_id IS NULL
      AND pet_souvenir.report_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1
          FROM pet_hourly_logs AS source_visit
          WHERE source_visit.pet_id = pet_souvenir.pet_id
            AND (source_visit.simulated_at AT TIME ZONE 'Asia/Tokyo')::DATE = pet_souvenir.found_on
            AND source_visit.group_master_id = pet_souvenir.source_group_master_id
      )
)
UPDATE pet_souvenirs AS pet_souvenir
SET
    souvenir_master_id = candidate.souvenir_master_id,
    source_group_master_id = candidate.group_master_id,
    note = candidate.group_display_name || 'で、小さなおみやげを見つけたようです。',
    updated_at = CURRENT_TIMESTAMP
FROM correction_candidates AS candidate
WHERE pet_souvenir.id = candidate.pet_souvenir_id;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.attach_souvenir_to_report()
RETURNS trigger AS $$
DECLARE
    target_souvenir_id UUID;
    target_master_id INTEGER;
    target_group_master_id INTEGER;
    target_group_display_name VARCHAR(100);
    daily_count INTEGER;
    report_created_at TIMESTAMPTZ;
    report_date DATE;
BEGIN
    report_created_at := COALESCE(NEW.created_at, CURRENT_TIMESTAMP);
    report_date := (report_created_at AT TIME ZONE 'Asia/Tokyo')::DATE;

    PERFORM pg_advisory_xact_lock(
        hashtextextended(NEW.pet_id::TEXT || ':' || report_date::TEXT, 0)
    );

    -- 毎時抽選でその日に取得済みのおみやげを最優先で紐付ける。
    SELECT pet_souvenir.id
    INTO target_souvenir_id
    FROM pet_souvenirs AS pet_souvenir
    WHERE pet_souvenir.pet_id = NEW.pet_id
      AND pet_souvenir.found_on = report_date
      AND pet_souvenir.report_id IS NULL
      AND pet_souvenir.found_at <= report_created_at
    ORDER BY pet_souvenir.found_at, pet_souvenir.id
    LIMIT 1
    FOR UPDATE SKIP LOCKED;

    IF target_souvenir_id IS NOT NULL THEN
        UPDATE pet_souvenirs
        SET
            report_id = NEW.id,
            reported_at = report_created_at,
            updated_at = report_created_at
        WHERE id = target_souvenir_id;

        RETURN NEW;
    END IF;

    SELECT COUNT(*)
    INTO daily_count
    FROM pet_souvenirs AS pet_souvenir
    WHERE pet_souvenir.pet_id = NEW.pet_id
      AND pet_souvenir.found_on = report_date;

    IF daily_count >= 1 THEN
        RETURN NEW;
    END IF;

    -- 最低保証は、その日に実際に訪れた群れだけを候補にする。
    -- レポートの群れにおみやげがあれば優先し、なければ直近の訪問群れを使う。
    SELECT
        souvenir_master.id,
        souvenir_master.group_master_id,
        group_master.display_name
    INTO
        target_master_id,
        target_group_master_id,
        target_group_display_name
    FROM pet_hourly_logs AS hourly_log
    INNER JOIN souvenir_masters AS souvenir_master
        ON souvenir_master.group_master_id = hourly_log.group_master_id
       AND souvenir_master.active = TRUE
       AND souvenir_master.display_name <> '謎のお土産'
    INNER JOIN group_masters AS group_master
        ON group_master.id = hourly_log.group_master_id
    WHERE hourly_log.pet_id = NEW.pet_id
      AND (hourly_log.simulated_at AT TIME ZONE 'Asia/Tokyo')::DATE = report_date
      AND hourly_log.simulated_at <= report_created_at
    ORDER BY
        CASE WHEN hourly_log.group_master_id = NEW.group_master_id THEN 0 ELSE 1 END,
        hourly_log.simulated_at DESC,
        hourly_log.id DESC
    LIMIT 1;

    -- 当日の訪問群れに有効なおみやげがなければ、事実と異なる履歴は作らない。
    IF target_master_id IS NULL THEN
        RETURN NEW;
    END IF;

    INSERT INTO pet_souvenirs (
        id,
        pet_id,
        souvenir_master_id,
        report_id,
        found_at,
        found_on,
        source_group_master_id,
        note,
        reported_at,
        created_at,
        updated_at
    ) VALUES (
        NEW.id,
        NEW.pet_id,
        target_master_id,
        NEW.id,
        report_created_at,
        report_date,
        target_group_master_id,
        target_group_display_name || 'で、小さなおみやげを見つけたようです。',
        report_created_at,
        report_created_at,
        report_created_at
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down

-- 補正した取得履歴は、誤ったマスタに戻さない。
-- Downでは、トリガーの選択方法だけを119適用前へ戻す。
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.attach_souvenir_to_report()
RETURNS trigger AS $$
DECLARE
    target_souvenir_id UUID;
    target_master_id INTEGER;
    target_group_master_id INTEGER;
    daily_count INTEGER;
    report_created_at TIMESTAMPTZ;
    report_date DATE;
BEGIN
    report_created_at := COALESCE(NEW.created_at, CURRENT_TIMESTAMP);
    report_date := (report_created_at AT TIME ZONE 'Asia/Tokyo')::DATE;

    PERFORM pg_advisory_xact_lock(
        hashtextextended(NEW.pet_id::TEXT || ':' || report_date::TEXT, 0)
    );

    SELECT pet_souvenir.id
    INTO target_souvenir_id
    FROM pet_souvenirs AS pet_souvenir
    WHERE pet_souvenir.pet_id = NEW.pet_id
      AND pet_souvenir.found_on = report_date
      AND pet_souvenir.report_id IS NULL
      AND pet_souvenir.found_at <= report_created_at
    ORDER BY pet_souvenir.found_at, pet_souvenir.id
    LIMIT 1
    FOR UPDATE SKIP LOCKED;

    IF target_souvenir_id IS NOT NULL THEN
        UPDATE pet_souvenirs
        SET
            report_id = NEW.id,
            reported_at = report_created_at,
            updated_at = report_created_at
        WHERE id = target_souvenir_id;

        RETURN NEW;
    END IF;

    SELECT COUNT(*)
    INTO daily_count
    FROM pet_souvenirs AS pet_souvenir
    WHERE pet_souvenir.pet_id = NEW.pet_id
      AND pet_souvenir.found_on = report_date;

    IF daily_count >= 1 THEN
        RETURN NEW;
    END IF;

    SELECT souvenir_master.id, souvenir_master.group_master_id
    INTO target_master_id, target_group_master_id
    FROM souvenir_masters AS souvenir_master
    WHERE souvenir_master.active = TRUE
      AND souvenir_master.display_name <> '謎のお土産'
    ORDER BY
        CASE WHEN souvenir_master.group_master_id = NEW.group_master_id THEN 0 ELSE 1 END,
        souvenir_master.id
    LIMIT 1;

    IF target_master_id IS NULL THEN
        RETURN NEW;
    END IF;

    INSERT INTO pet_souvenirs (
        id,
        pet_id,
        souvenir_master_id,
        report_id,
        found_at,
        found_on,
        source_group_master_id,
        note,
        reported_at,
        created_at,
        updated_at
    ) VALUES (
        NEW.id,
        NEW.pet_id,
        target_master_id,
        NEW.id,
        report_created_at,
        report_date,
        target_group_master_id,
        'レポートといっしょに、小さなおみやげを持ち帰りました。',
        report_created_at,
        report_created_at,
        report_created_at
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
