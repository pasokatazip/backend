-- +goose Up

-- 毎日出力されるレポートごとに「ほめる」状態を持てるよう、
-- ユーザー単位の主キーから (user_id, report_date) の複合主キーへ拡張する。
ALTER TABLE user_souvenir_praise_flags
    ADD COLUMN IF NOT EXISTS report_date DATE;

-- 112 を先に適用済みの環境にレコードがある場合の救済。
-- 旧データにはレポート対象日がないため、選択時の JST 日付を使う。
UPDATE user_souvenir_praise_flags
SET report_date = (
    COALESCE(praised_at, created_at) AT TIME ZONE 'Asia/Tokyo'
)::DATE
WHERE report_date IS NULL;

ALTER TABLE user_souvenir_praise_flags
    ALTER COLUMN report_date SET NOT NULL;

ALTER TABLE user_souvenir_praise_flags
    DROP CONSTRAINT IF EXISTS user_souvenir_praise_flags_pkey;

ALTER TABLE user_souvenir_praise_flags
    ADD CONSTRAINT user_souvenir_praise_flags_pkey
    PRIMARY KEY (user_id, report_date);

-- +goose Down

-- ユーザー単位に戻す際は、最も早い「ほめる」履歴のみ残す。
WITH ranked_flags AS (
    SELECT
        ctid,
        ROW_NUMBER() OVER (
            PARTITION BY user_id
            ORDER BY praised_at ASC NULLS LAST, report_date ASC
        ) AS flag_order
    FROM user_souvenir_praise_flags
)
DELETE FROM user_souvenir_praise_flags AS flag
USING ranked_flags AS ranked
WHERE flag.ctid = ranked.ctid
  AND ranked.flag_order > 1;

ALTER TABLE user_souvenir_praise_flags
    DROP CONSTRAINT IF EXISTS user_souvenir_praise_flags_pkey;

ALTER TABLE user_souvenir_praise_flags
    DROP COLUMN IF EXISTS report_date;

ALTER TABLE user_souvenir_praise_flags
    ADD CONSTRAINT user_souvenir_praise_flags_pkey PRIMARY KEY (user_id);
