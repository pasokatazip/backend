-- +goose Up
-- 同じ群れにいた他ユーザーの投稿本文を、レポートごとに最大2件保存する。
ALTER TABLE reports
ADD COLUMN IF NOT EXISTS rumor JSONB NOT NULL DEFAULT '[]'::JSONB;

-- migration適用前に作られたレポートにも、対応する毎時ログから噂を補完する。
UPDATE reports AS report
SET rumor = COALESCE(rumor_posts.items, '[]'::JSONB)
FROM pet_hourly_logs AS hourly_log
INNER JOIN pets AS reporting_pet ON reporting_pet.id = hourly_log.pet_id
LEFT JOIN LATERAL (
    SELECT jsonb_agg(candidate.content ORDER BY candidate.created_at DESC) AS items
    FROM (
        SELECT post.content, post.created_at
        FROM pet_hourly_logs AS nearby_log
        INNER JOIN pets AS nearby_pet ON nearby_pet.id = nearby_log.pet_id
        INNER JOIN posts AS post ON post.pet_id = nearby_pet.id
        WHERE nearby_log.simulated_at = hourly_log.simulated_at
            AND nearby_log.group_master_id = hourly_log.group_master_id
            AND nearby_pet.user_id <> reporting_pet.user_id
            AND post.created_at <= hourly_log.simulated_at
        ORDER BY post.created_at DESC
        LIMIT 2
    ) AS candidate
) AS rumor_posts ON TRUE
WHERE report.pet_id = hourly_log.pet_id
    AND report.created_at = hourly_log.simulated_at
    AND report.rumor = '[]'::JSONB;

-- +goose Down
ALTER TABLE reports DROP COLUMN IF EXISTS rumor;
