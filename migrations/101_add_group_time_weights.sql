-- +goose Up
-- 時間帯ごとの群れ選択スコアに掛ける倍率。
-- 1.00 は中立、1.00 より大きいほどその時間帯で選ばれやすい。
ALTER TABLE
    group_masters
ADD
    COLUMN IF NOT EXISTS morning_weight DECIMAL(5, 2) NOT NULL DEFAULT 1.00,
ADD
    COLUMN IF NOT EXISTS afternoon_weight DECIMAL(5, 2) NOT NULL DEFAULT 1.00,
ADD
    COLUMN IF NOT EXISTS night_weight DECIMAL(5, 2) NOT NULL DEFAULT 1.00;

-- +goose Down
SELECT
    1;