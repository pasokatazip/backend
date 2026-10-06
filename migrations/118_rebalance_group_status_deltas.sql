-- +goose Up
-- 群れ効果の強弱順は保ったまま、各状態が一方向に増え続けない範囲へ再配分する。
-- 全群れの平均deltaを0付近にし、選択アルゴリズムだけに補正を任せない。
UPDATE group_masters
SET
    energy_delta = CASE energy_delta
        WHEN -0.1400 THEN -0.0300 WHEN -0.1300 THEN -0.0280 WHEN -0.1200 THEN -0.0260
        WHEN -0.1100 THEN -0.0240 WHEN -0.1000 THEN -0.0220 WHEN -0.0900 THEN -0.0200
        WHEN -0.0800 THEN -0.0180 WHEN -0.0300 THEN -0.0050
        WHEN 0.0800 THEN 0.0500 WHEN 0.0900 THEN 0.0550 WHEN 0.1000 THEN 0.0600
        WHEN 0.1100 THEN 0.0650 WHEN 0.1200 THEN 0.0700 WHEN 0.1300 THEN 0.0750
        WHEN 0.1400 THEN 0.0800 ELSE energy_delta
    END,
    curiosity_delta = CASE curiosity_delta
        WHEN -0.0800 THEN -0.0600 WHEN 0.0600 THEN -0.0500 WHEN 0.0800 THEN -0.0400
        WHEN 0.0900 THEN -0.0300 WHEN 0.1000 THEN -0.0200 WHEN 0.1100 THEN 0.0000
        WHEN 0.1200 THEN 0.0200 WHEN 0.1300 THEN 0.0400 WHEN 0.1400 THEN 0.0600
        ELSE curiosity_delta
    END,
    sociality_delta = CASE sociality_delta
        WHEN -0.0800 THEN -0.0400 WHEN 0.0500 THEN -0.0300 WHEN 0.0800 THEN 0.0000
        WHEN 0.0900 THEN 0.0100 WHEN 0.1000 THEN 0.0200 WHEN 0.1200 THEN 0.0400
        WHEN 0.1300 THEN 0.0500 WHEN 0.1400 THEN 0.0600 ELSE sociality_delta
    END,
    routine_delta = CASE routine_delta
        WHEN -0.1400 THEN -0.0800 WHEN -0.1200 THEN -0.0700 WHEN -0.1000 THEN -0.0600
        WHEN -0.0900 THEN -0.0500 WHEN -0.0800 THEN -0.0400 WHEN -0.0100 THEN -0.0100
        WHEN 0.0800 THEN 0.0000 WHEN 0.0900 THEN 0.0100 WHEN 0.1000 THEN 0.0200
        WHEN 0.1100 THEN 0.0300 WHEN 0.1200 THEN 0.0400 WHEN 0.1300 THEN 0.0500
        WHEN 0.1400 THEN 0.0600 ELSE routine_delta
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE EXISTS (
    -- 新規環境では調整済みseedが入るため、再変換しない。
    SELECT 1
    FROM group_masters AS current_master
    WHERE current_master.curiosity_delta > 0.0800
       OR current_master.sociality_delta > 0.0800
       OR current_master.routine_delta > 0.0800
       OR current_master.energy_delta < -0.0300
);

-- +goose Down
-- Upの対応はすべて1対1のため、元の値へ戻せる。
UPDATE group_masters
SET
    energy_delta = CASE energy_delta
        WHEN -0.0300 THEN -0.1400 WHEN -0.0280 THEN -0.1300 WHEN -0.0260 THEN -0.1200
        WHEN -0.0240 THEN -0.1100 WHEN -0.0220 THEN -0.1000 WHEN -0.0200 THEN -0.0900
        WHEN -0.0180 THEN -0.0800 WHEN -0.0050 THEN -0.0300
        WHEN 0.0500 THEN 0.0800 WHEN 0.0550 THEN 0.0900 WHEN 0.0600 THEN 0.1000
        WHEN 0.0650 THEN 0.1100 WHEN 0.0700 THEN 0.1200 WHEN 0.0750 THEN 0.1300
        WHEN 0.0800 THEN 0.1400 ELSE energy_delta
    END,
    curiosity_delta = CASE curiosity_delta
        WHEN -0.0600 THEN -0.0800 WHEN -0.0500 THEN 0.0600 WHEN -0.0400 THEN 0.0800
        WHEN -0.0300 THEN 0.0900 WHEN -0.0200 THEN 0.1000 WHEN 0.0000 THEN 0.1100
        WHEN 0.0200 THEN 0.1200 WHEN 0.0400 THEN 0.1300 WHEN 0.0600 THEN 0.1400
        ELSE curiosity_delta
    END,
    sociality_delta = CASE sociality_delta
        WHEN -0.0400 THEN -0.0800 WHEN -0.0300 THEN 0.0500 WHEN 0.0000 THEN 0.0800
        WHEN 0.0100 THEN 0.0900 WHEN 0.0200 THEN 0.1000 WHEN 0.0400 THEN 0.1200
        WHEN 0.0500 THEN 0.1300 WHEN 0.0600 THEN 0.1400 ELSE sociality_delta
    END,
    routine_delta = CASE routine_delta
        WHEN -0.0800 THEN -0.1400 WHEN -0.0700 THEN -0.1200 WHEN -0.0600 THEN -0.1000
        WHEN -0.0500 THEN -0.0900 WHEN -0.0400 THEN -0.0800 WHEN -0.0100 THEN -0.0100
        WHEN 0.0000 THEN 0.0800 WHEN 0.0100 THEN 0.0900 WHEN 0.0200 THEN 0.1000
        WHEN 0.0300 THEN 0.1100 WHEN 0.0400 THEN 0.1200 WHEN 0.0500 THEN 0.1300
        WHEN 0.0600 THEN 0.1400 ELSE routine_delta
    END,
    updated_at = CURRENT_TIMESTAMP;
