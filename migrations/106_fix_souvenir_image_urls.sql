-- +goose Up
-- MinIOの管理画面URLを、ブラウザから直接表示できる公開画像URLへ置換する。
UPDATE souvenir_masters
SET
    image_url = REPLACE(
        image_url,
        'https://minio.akatukii.com/browser/assets/Souvenirs%2F',
        'https://assets.akatukii.com/assets/Souvenirs/'
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE POSITION(
    'https://minio.akatukii.com/browser/assets/Souvenirs%2F'
    IN image_url
) = 1;

-- +goose Down
-- 公開画像URLを、マイグレーション適用前のMinIO管理画面URLへ戻す。
UPDATE souvenir_masters
SET
    image_url = REPLACE(
        image_url,
        'https://assets.akatukii.com/assets/Souvenirs/',
        'https://minio.akatukii.com/browser/assets/Souvenirs%2F'
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE POSITION(
    'https://assets.akatukii.com/assets/Souvenirs/'
    IN image_url
) = 1;
