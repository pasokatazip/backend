-- pgweb の pet_yoyo で実行する。再実行しても既存の成長・投稿をリセットしない。
-- パスワードはログイン不能な印。通知・課金登録は作成しない。
BEGIN;
SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';
DO $$
DECLARE
    n INTEGER;
    uid UUID;
    pid UUID;
    gid INTEGER;
    mail TEXT;
    topics TEXT[] := ARRAY['game','cooking','walk','reading','music','drawing','programming','cafe','photography','gardening','exercise','movie','anime','crafts','travel','pet','cleaning','board_game','language_learning','sleep'];
    colors TEXT[] := ARRAY['#FFC1CA','#9AD7FF','#B8F7C5','#FFD166','#CDB4DB'];
BEGIN
    -- 同じ初期化を同時に実行しても、1ユーザー1匹の割り当てを守る。
    PERFORM pg_advisory_xact_lock(hashtextextended('yoyo-test-accounts-v1', 0));
    IF NOT EXISTS (SELECT 1 FROM evolution_stages WHERE id = 0) THEN
        RAISE EXCEPTION '初期進化段階 id=0 がありません';
    END IF;
    FOR n IN 1..20 LOOP
        uid := ('77000000-0000-4000-8000-' || lpad(n::TEXT,12,'0'))::UUID;
        pid := ('77100000-0000-4000-8000-' || lpad(n::TEXT,12,'0'))::UUID;
        mail := 'yoyo-test-' || lpad(n::TEXT,2,'0') || '@example.invalid';
        SELECT id INTO gid FROM group_masters WHERE group_key = topics[n] AND active;
        IF gid IS NULL THEN
            RAISE EXCEPTION '群れ % がありません', topics[n];
        END IF;
        IF EXISTS (SELECT 1 FROM users WHERE (id=uid OR email=mail) AND NOT (id=uid AND email=mail AND password='!automated-test-account-disabled')) THEN
            RAISE EXCEPTION 'テストユーザー % と既存データが衝突しています', n;
        END IF;
        IF EXISTS (SELECT 1 FROM pets WHERE id=pid AND user_id<>uid) THEN
            RAISE EXCEPTION 'テストペット % と既存データが衝突しています', n;
        END IF;
        INSERT INTO users (id,email,password,subsc)
        VALUES (uid,mail,'!automated-test-account-disabled',FALSE)
        ON CONFLICT (id) DO NOTHING;
        INSERT INTO pets (id,user_id,name,color,status,current_group_master_id,current_stage_id)
        VALUES (pid,uid,'テスト' || lpad(n::TEXT,2,'0'),colors[1+(n-1)%5],'active',gid,0)
        ON CONFLICT (id) DO NOTHING;
        INSERT INTO user_active_pets (user_id,pet_id)
        SELECT uid,pid WHERE NOT EXISTS (SELECT 1 FROM user_active_pets WHERE user_id=uid)
          AND EXISTS (SELECT 1 FROM pets WHERE id=pid AND status='active' AND NOT is_deleted)
        ON CONFLICT (user_id) DO NOTHING;
        INSERT INTO pet_experiences (id,pet_id)
        VALUES (('77200000-0000-4000-8000-' || lpad(n::TEXT,12,'0'))::UUID,pid)
        ON CONFLICT (pet_id) DO NOTHING;
        INSERT INTO pet_group_joins (id,pet_id,group_master_id,move_reason)
        SELECT ('77300000-0000-4000-8000-' || lpad(n::TEXT,12,'0'))::UUID,pid,gid,'initial'
        WHERE NOT EXISTS (SELECT 1 FROM pet_group_joins WHERE pet_id=pid)
          AND EXISTS (SELECT 1 FROM pets WHERE id=pid AND status='active' AND NOT is_deleted AND current_group_master_id=gid)
        ON CONFLICT (id) DO NOTHING;
    END LOOP;
END $$;
COMMIT;

SELECT u.email,p.id AS pet_id,p.name,gm.display_name,pe.feed_count
FROM users u
JOIN user_active_pets ap ON ap.user_id=u.id
JOIN pets p ON p.id=ap.pet_id
JOIN pet_experiences pe ON pe.pet_id=p.id
LEFT JOIN group_masters gm ON gm.id=p.current_group_master_id
WHERE u.id::TEXT LIKE '77000000-0000-4000-8000-%'
ORDER BY u.email;
