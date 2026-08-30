-- +goose Up

-- reports.rumor は表示文しか保持しないため、同じ投稿由来の噂を受け取ったか判定できない。
-- 受取ユーザーと噂元投稿の組み合わせを保存し、ペット交代後も同じ噂の再取得をDB制約で防ぐ。
CREATE TABLE IF NOT EXISTS user_rumor_receipts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    received_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_rumor_receipts_user_source UNIQUE (
        user_id,
        source_post_id
    )
);

CREATE INDEX IF NOT EXISTS idx_user_rumor_receipts_source_post_id
    ON user_rumor_receipts(source_post_id);

CREATE INDEX IF NOT EXISTS idx_user_rumor_receipts_report_id
    ON user_rumor_receipts(report_id);

-- 過去のレポートに保存済みの噂も、可能な範囲で元投稿へ対応付ける。
-- 同じ噂が複数レポートにある場合は、最初に受け取ったレポートを履歴として残す。
WITH receipt_candidates AS (
    SELECT
        report.user_id,
        candidate.source_post_id,
        report.id AS report_id,
        report.created_at AS received_at
    FROM reports AS report
    INNER JOIN pets AS reporting_pet ON reporting_pet.id = report.pet_id
    CROSS JOIN LATERAL (
        SELECT
            post.id AS source_post_id,
            post.created_at
        FROM pet_hourly_logs AS nearby_log
        INNER JOIN pets AS nearby_pet ON nearby_pet.id = nearby_log.pet_id
        INNER JOIN posts AS post ON post.pet_id = nearby_pet.id
        WHERE nearby_log.simulated_at = report.created_at
            AND nearby_log.group_master_id = report.group_master_id
            AND nearby_pet.user_id <> reporting_pet.user_id
            AND post.created_at <= report.created_at
            AND EXISTS (
                SELECT 1
                FROM jsonb_array_elements_text(report.rumor) AS rumor_item(content)
                WHERE rumor_item.content = post.content
            )
        ORDER BY post.created_at DESC, post.id
        LIMIT 2
    ) AS candidate
),
first_receipts AS (
    SELECT DISTINCT ON (user_id, source_post_id)
        user_id,
        source_post_id,
        report_id,
        received_at
    FROM receipt_candidates
    ORDER BY user_id, source_post_id, received_at, report_id
)
INSERT INTO user_rumor_receipts (
    id, user_id, source_post_id, report_id, received_at, created_at
)
SELECT
    md5(user_id::TEXT || ':' || source_post_id::TEXT)::UUID,
    user_id,
    source_post_id,
    report_id,
    received_at,
    received_at
FROM first_receipts
ON CONFLICT (user_id, source_post_id) DO NOTHING;

-- +goose Down

DROP TABLE IF EXISTS user_rumor_receipts;
