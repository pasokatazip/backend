-- +goose Up

-- 旧トリガーがレポートごとに生成した過剰なお土産を、
-- ペット・JST暦日ごとに3件までに整理する。
-- 毎時抽選で実際に取得した履歴を優先し、同条件では古いものを残す。
WITH ranked_souvenirs AS (
    SELECT
        id,
        ROW_NUMBER() OVER (
            PARTITION BY pet_id, found_on
            ORDER BY
                (report_id IS NOT NULL) DESC,
                (pet_hourly_log_id IS NOT NULL) DESC,
                found_at,
                id
        ) AS souvenir_order
    FROM pet_souvenirs
)
DELETE FROM pet_souvenirs ps
USING ranked_souvenirs ranked
WHERE ps.id = ranked.id
  AND ranked.souvenir_order > 3;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_daily_souvenir_limit()
RETURNS trigger AS $$
DECLARE
    daily_count INTEGER;
BEGIN
    PERFORM pg_advisory_xact_lock(
        hashtextextended(NEW.pet_id::TEXT || ':' || NEW.found_on::TEXT, 0)
    );

    SELECT COUNT(*)
    INTO daily_count
    FROM pet_souvenirs
    WHERE pet_id = NEW.pet_id
      AND found_on = NEW.found_on;

    IF daily_count >= 3 THEN
        RETURN NULL;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trigger_enforce_daily_souvenir_limit ON pet_souvenirs;

CREATE TRIGGER trigger_enforce_daily_souvenir_limit
BEFORE INSERT ON pet_souvenirs
FOR EACH ROW
EXECUTE FUNCTION enforce_daily_souvenir_limit();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION attach_souvenir_to_report()
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

    SELECT ps.id
    INTO target_souvenir_id
    FROM pet_souvenirs ps
    WHERE ps.pet_id = NEW.pet_id
      AND ps.found_on = report_date
      AND ps.report_id IS NULL
      AND ps.found_at <= report_created_at
    ORDER BY ps.found_at, ps.id
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
    FROM pet_souvenirs ps
    WHERE ps.pet_id = NEW.pet_id
      AND ps.found_on = report_date;

    IF daily_count >= 1 THEN
        RETURN NEW;
    END IF;

    SELECT sm.id, sm.group_master_id
    INTO target_master_id, target_group_master_id
    FROM souvenir_masters sm
    WHERE sm.active = TRUE
    ORDER BY
        CASE WHEN sm.group_master_id = NEW.group_master_id THEN 0 ELSE 1 END,
        sm.id
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

DROP TRIGGER IF EXISTS trigger_attach_souvenir_to_report ON reports;

CREATE TRIGGER trigger_attach_souvenir_to_report
AFTER INSERT ON reports
FOR EACH ROW
EXECUTE FUNCTION attach_souvenir_to_report();

-- +goose Down

DROP TRIGGER IF EXISTS trigger_enforce_daily_souvenir_limit ON pet_souvenirs;

-- +goose StatementBegin
DROP FUNCTION IF EXISTS enforce_daily_souvenir_limit();
-- +goose StatementEnd

-- Downでは旧仕様（レポートごとに1件保証）の関数に戻す。
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION attach_souvenir_to_report()
RETURNS trigger AS $$
DECLARE
    target_souvenir_id UUID;
    target_master_id INTEGER;
    target_group_master_id INTEGER;
    report_created_at TIMESTAMPTZ;
    report_date DATE;
BEGIN
    report_created_at := COALESCE(NEW.created_at, CURRENT_TIMESTAMP);
    report_date := (report_created_at AT TIME ZONE 'Asia/Tokyo')::DATE;

    SELECT ps.id
    INTO target_souvenir_id
    FROM pet_souvenirs ps
    WHERE ps.pet_id = NEW.pet_id
      AND ps.report_id IS NULL
      AND ps.found_at <= report_created_at
    ORDER BY
        CASE WHEN ps.found_on = report_date THEN 0 ELSE 1 END,
        ps.found_at DESC,
        ps.id
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

    SELECT sm.id, sm.group_master_id
    INTO target_master_id, target_group_master_id
    FROM souvenir_masters sm
    WHERE sm.active = TRUE
    ORDER BY
        CASE WHEN sm.group_master_id = NEW.group_master_id THEN 0 ELSE 1 END,
        sm.id
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
