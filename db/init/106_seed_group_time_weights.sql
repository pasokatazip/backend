-- group_masters の時間帯ごとの行きやすさ
-- すべての群れをまず 1.00（中立）で更新し、時間帯の特徴がある群れだけを上書きする。
UPDATE group_masters
SET
    morning_weight = 1.00,
    afternoon_weight = 1.00,
    night_weight = 1.00;

-- 朝・昼・夜の移動候補スコアへ倍率として利用する。
WITH seed_group_time_weights (
    group_key,
    morning_weight,
    afternoon_weight,
    night_weight
) AS (
    VALUES
        ('morning', 2.00, 0.80, 0.60),
        ('meal', 1.50, 1.50, 1.20),
        ('cooking', 1.40, 1.40, 1.10),
        ('cleaning', 1.50, 1.20, 0.60),
        ('laundry', 1.50, 1.20, 0.60),
        ('gardening', 1.60, 1.30, 0.50),
        ('walk', 1.60, 1.30, 0.70),
        ('work', 1.30, 1.50, 0.70),
        ('meeting', 1.20, 1.50, 0.60),
        ('study', 1.20, 1.50, 0.90),
        ('shopping', 1.10, 1.50, 0.90),
        ('cafe', 1.10, 1.50, 1.10),
        ('park', 1.30, 1.50, 0.70),
        ('late_night', 0.40, 0.60, 2.00),
        ('alcohol', 0.40, 0.70, 1.90),
        ('bath', 0.70, 1.00, 1.80),
        ('sleep', 0.50, 0.70, 2.00),
        ('game', 0.70, 1.10, 1.60),
        ('anime', 0.60, 1.00, 1.60),
        ('movie', 0.60, 1.00, 1.70),
        ('music', 0.80, 1.20, 1.50),
        ('streaming', 0.70, 1.20, 1.60)
)
UPDATE group_masters AS group_master
SET
    morning_weight = seed.morning_weight,
    afternoon_weight = seed.afternoon_weight,
    night_weight = seed.night_weight
FROM seed_group_time_weights AS seed
WHERE group_master.group_key = seed.group_key;
