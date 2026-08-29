-- +goose Up

-- おみやげの公開導線で「ほめる」を一度選んだかをユーザー単位で保存する。
-- 113 で日次レポート単位へ拡張する前のスキーマとして保持する。
CREATE TABLE user_souvenir_praise_flags (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    has_praised BOOLEAN NOT NULL DEFAULT FALSE,
    praised_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_user_souvenir_praise_flags_praised_at CHECK (
        (has_praised = FALSE AND praised_at IS NULL)
        OR (has_praised = TRUE AND praised_at IS NOT NULL)
    )
);

-- +goose Down

DROP TABLE IF EXISTS user_souvenir_praise_flags;
