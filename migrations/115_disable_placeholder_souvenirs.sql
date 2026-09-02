-- +goose Up
-- おみやげ未設定の群れ用に作られていたプレースホルダーを取得対象から外す。
-- 既存の取得履歴から参照されている可能性があるため、レコードは削除しない。
UPDATE souvenir_masters
SET
    active = FALSE,
    updated_at = CURRENT_TIMESTAMP
WHERE display_name = '謎のお土産';

-- +goose Down
UPDATE souvenir_masters
SET
    active = TRUE,
    updated_at = CURRENT_TIMESTAMP
WHERE display_name = '謎のお土産';
