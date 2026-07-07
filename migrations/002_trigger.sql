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

-- +goose Down
DROP TRIGGER IF EXISTS trigger_post_created ON posts;

-- +goose StatementBegin
DROP FUNCTION IF EXISTS notify_post_created();
-- +goose StatementEnd