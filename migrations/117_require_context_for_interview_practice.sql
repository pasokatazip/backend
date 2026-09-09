-- +goose Up
-- 「プログラミングの練習」だけで面接へ興味が加算されないようにする。
-- 群れ単位の文脈判定へ修正するPythonコードと同じリリースで適用する。
UPDATE group_keywords gk
SET match_type = 'requires_context', updated_at = CURRENT_TIMESTAMP
FROM group_masters gm
WHERE gm.id = gk.group_master_id
  AND gm.group_key = 'interview'
  AND gk.normalized_keyword = '練習'
  AND gk.match_type = 'exact_or_partial';

-- +goose Down
UPDATE group_keywords gk
SET match_type = 'exact_or_partial', updated_at = CURRENT_TIMESTAMP
FROM group_masters gm
WHERE gm.id = gk.group_master_id
  AND gm.group_key = 'interview'
  AND gk.normalized_keyword = '練習'
  AND gk.match_type = 'requires_context';
