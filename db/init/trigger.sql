--postトリガー
--contentが入るとPythonにイベント発火する

CREATE OR REPLACE FUNCTION notify_post_created()
RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('post_created', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_post_created ON posts;

CREATE TRIGGER trigger_post_created
AFTER INSERT ON posts
FOR EACH ROW
EXECUTE FUNCTION notify_post_created();