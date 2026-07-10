-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION notify_post_created()
RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('post_created', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trigger_post_created ON posts;

CREATE TRIGGER trigger_post_created
AFTER INSERT ON posts
FOR EACH ROW
EXECUTE FUNCTION notify_post_created();

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

DROP TRIGGER IF EXISTS trigger_attach_souvenir_to_report ON reports;

CREATE TRIGGER trigger_attach_souvenir_to_report
AFTER INSERT ON reports
FOR EACH ROW
EXECUTE FUNCTION attach_souvenir_to_report();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_attach_souvenir_to_report ON reports;

-- +goose StatementBegin
DROP FUNCTION IF EXISTS attach_souvenir_to_report();
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trigger_post_created ON posts;

-- +goose StatementBegin
DROP FUNCTION IF EXISTS notify_post_created();
-- +goose StatementEnd
