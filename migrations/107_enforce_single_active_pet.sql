-- +goose Up

-- 旧スキーマでは (user_id, pet_id) の複合主キーだったため、
-- 同じユーザーに複数のアクティブペットを登録できた。
-- 最新の割り当てだけを残し、それ以前のペットは履歴として archived にする。
-- 所有者が違う行や、すでに非アクティブなペットの古い関連も先に除外する。
DELETE FROM user_active_pets uap
USING pets p
WHERE uap.pet_id = p.id
  AND (
      uap.user_id <> p.user_id
      OR p.is_deleted = TRUE
      OR p.status <> 'active'
  );

WITH ranked_active_pets AS (
    SELECT
        user_id,
        pet_id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id
            ORDER BY assigned_at DESC, pet_id DESC
        ) AS active_order
    FROM user_active_pets
), superseded_pets AS (
    SELECT pet_id
    FROM ranked_active_pets
    WHERE active_order > 1
)
UPDATE pets
SET
    status = 'archived',
    updated_at = CURRENT_TIMESTAMP
WHERE id IN (SELECT pet_id FROM superseded_pets);

WITH ranked_active_pets AS (
    SELECT
        user_id,
        pet_id,
        ROW_NUMBER() OVER (
            PARTITION BY user_id
            ORDER BY assigned_at DESC, pet_id DESC
        ) AS active_order
    FROM user_active_pets
)
DELETE FROM user_active_pets
WHERE (user_id, pet_id) IN (
    SELECT user_id, pet_id
    FROM ranked_active_pets
    WHERE active_order > 1
);

ALTER TABLE user_active_pets
    DROP CONSTRAINT user_active_pets_pkey,
    ADD CONSTRAINT user_active_pets_pkey PRIMARY KEY (user_id),
    ADD CONSTRAINT user_active_pets_pet_id_key UNIQUE (pet_id);

-- +goose Down

ALTER TABLE user_active_pets
    DROP CONSTRAINT user_active_pets_pet_id_key,
    DROP CONSTRAINT user_active_pets_pkey,
    ADD CONSTRAINT user_active_pets_pkey PRIMARY KEY (user_id, pet_id);
