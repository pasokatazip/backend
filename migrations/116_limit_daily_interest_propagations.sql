-- +goose Up

-- 群れの気配の伝播は、受け取るペットごとにJSTの同日内2回までにする。
-- 同じペットの同日伝播を直列化し、毎時ジョブが並行しても上限を超えないようにする。
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION enforce_daily_interest_propagation_limit()
RETURNS trigger AS $$
DECLARE
    daily_start TIMESTAMPTZ;
    daily_count INTEGER;
BEGIN
    daily_start := (
        (NEW.occurred_at AT TIME ZONE 'Asia/Tokyo')::DATE::TIMESTAMP
        AT TIME ZONE 'Asia/Tokyo'
    );

    PERFORM pg_advisory_xact_lock(
        hashtextextended(
            'interest-propagation:' || NEW.recipient_pet_id::TEXT || ':' || daily_start::TEXT,
            0
        )
    );

    SELECT COUNT(*)
    INTO daily_count
    FROM pet_interest_propagations
    WHERE recipient_pet_id = NEW.recipient_pet_id
      AND occurred_at >= daily_start
      AND occurred_at < daily_start + INTERVAL '1 day';

    -- 上限到達はジョブ全体の失敗にせず、この伝播だけをスキップする。
    IF daily_count >= 2 THEN
        RETURN NULL;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trigger_enforce_daily_interest_propagation_limit
    ON pet_interest_propagations;

CREATE TRIGGER trigger_enforce_daily_interest_propagation_limit
BEFORE INSERT ON pet_interest_propagations
FOR EACH ROW
EXECUTE FUNCTION enforce_daily_interest_propagation_limit();

-- +goose Down

DROP TRIGGER IF EXISTS trigger_enforce_daily_interest_propagation_limit
    ON pet_interest_propagations;

-- +goose StatementBegin
DROP FUNCTION IF EXISTS enforce_daily_interest_propagation_limit();
-- +goose StatementEnd
