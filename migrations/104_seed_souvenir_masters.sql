-- +goose Up
WITH specified_souvenirs (group_key, display_name, image_url) AS (
    VALUES
        (
            'camping',
            '炭',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fcamping.png'
        ),
        (
            'cold',
            '干からびたシート',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fcold.png'
        ),
        (
            'cooking',
            '食べかけ即席麺',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fcooking.png'
        ),
        (
            'drawing',
            'スケッチブック',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fdrawing.png'
        ),
        (
            'dream',
            'バクのまくら',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fdream.png'
        ),
        (
            'exercise',
            'うさぎなわとび',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fexercise.png'
        ),
        (
            'gadget',
            'どっかのネジ',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fgadget.png'
        ),
        (
            'oshi_katsu',
            'らぶらぶチェキ',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Foshi_katsu.png'
        ),
        (
            'philosophy',
            'イデア的りんご',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fphilosophy.png'
        ),
        (
            'reflection',
            'かしおり',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Freflection.png'
        ),
        (
            'sticker',
            'はかせシール',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fsticker.png'
        ),
        (
            'youtube',
            'はじまりスライム',
            'https://minio.akatukii.com/browser/assets/Souvenirs%2Fyoutube.png'
        )
)
INSERT INTO
    souvenir_masters (
        souvenir_key,
        display_name,
        description,
        image_url,
        group_master_id,
        active
    )
SELECT
    group_masters.group_key || '_souvenir',
    COALESCE(specified_souvenirs.display_name, '謎のお土産'),
    NULL,
    specified_souvenirs.image_url,
    group_masters.id,
    TRUE
FROM
    group_masters
    LEFT JOIN specified_souvenirs ON specified_souvenirs.group_key = group_masters.group_key ON CONFLICT (group_master_id) DO
UPDATE
SET
    souvenir_key = EXCLUDED.souvenir_key,
    display_name = EXCLUDED.display_name,
    description = NULL,
    image_url = EXCLUDED.image_url,
    active = TRUE,
    updated_at = CURRENT_TIMESTAMP;

-- +goose Down
-- おみやげはユーザーの取得履歴から参照されるため、ロールバック時には削除しない。
SELECT
    1;