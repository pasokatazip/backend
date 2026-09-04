-- 事前確認済みの40匹・680件だけを補完する。失敗時はDO文全体がロールバックされる。
DO $reconstruct$
DECLARE
    candidate_count INTEGER;
    inserted_count INTEGER;
    batch_count INTEGER;
    trigger_state "char";
BEGIN
    PERFORM set_config('lock_timeout', '5s', TRUE);
    -- トリガーの一時停止中に他のreports INSERTが入り込まないよう、同じトランザクションでロックする。
    LOCK TABLE public.reports IN ACCESS EXCLUSIVE MODE;

    SELECT tgenabled INTO STRICT trigger_state
    FROM pg_trigger
    WHERE tgrelid = 'public.reports'::regclass
      AND tgname = 'trigger_attach_souvenir_to_report';
    IF trigger_state <> 'O' THEN
        RAISE EXCEPTION 'Unexpected souvenir trigger state: %', trigger_state;
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgrelid = 'public.reports'::regclass
          AND NOT tgisinternal AND tgname <> 'trigger_attach_souvenir_to_report'
    ) THEN
        RAISE EXCEPTION 'Unexpected reports trigger: review side effects before reconstruction';
    END IF;

    CREATE TEMP TABLE report_reconstruction_candidates ON COMMIT DROP AS
    WITH target_hours AS (
        SELECT generate_series(
            TIMESTAMPTZ '2026-09-03 18:00:00+09',
            TIMESTAMPTZ '2026-09-04 10:00:00+09', INTERVAL '1 hour'
        ) AS at
    ), eligible AS (
        SELECT p.id AS pet_id, p.user_id, h.at
        FROM public.pets p
        JOIN public.user_active_pets a ON a.pet_id = p.id
        CROSS JOIN target_hours h
        WHERE p.is_deleted = FALSE AND p.status = 'active' AND p.created_at <= h.at
          AND NOT EXISTS (
              SELECT 1 FROM public.reports r WHERE r.pet_id = p.id AND r.created_at = h.at
          )
    )
    SELECT e.*, s.id AS source_report_id, s.group_master_id,
           s.created_at AS source_created_at,
           LEFT(COALESCE(NULLIF(s.gossip, ''),
               gm.display_name || 'の近くで、のんびり過ごしていたようです。'), 255) AS gossip,
           LEFT(COALESCE(NULLIF(s.behavior_label, ''), '群れでゆっくり過ごした'), 255) AS behavior_label
    FROM eligible e
    JOIN LATERAL (
        SELECT r.* FROM public.reports r
        WHERE r.pet_id = e.pet_id
          AND r.created_at < TIMESTAMPTZ '2026-09-03 18:00:00+09'
          AND COALESCE(r.reason_json->>'source', '') <> 'historical_report_reconstruction'
        ORDER BY (r.hour_slot = EXTRACT(HOUR FROM e.at AT TIME ZONE 'Asia/Tokyo')) DESC,
                 r.created_at DESC, r.id
        LIMIT 1
    ) s ON TRUE
    JOIN public.group_masters gm ON gm.id = s.group_master_id;

    SELECT COUNT(*) INTO candidate_count FROM report_reconstruction_candidates;
    SELECT COUNT(*) INTO batch_count FROM public.reports
    WHERE reason_json->>'reconstruction_batch' = 'incident-20260903-1800-through-20260904-1000';
    -- 同じ補完の再実行は何も追加しない。
    IF candidate_count = 0 AND batch_count = 680 THEN
        RETURN;
    END IF;
    IF candidate_count <> 680 OR batch_count <> 0 THEN
        RAISE EXCEPTION 'Preview changed: candidates=%, existing_batch=%', candidate_count, batch_count;
    END IF;

    -- 補完は表示データだけ。おみやげ付与・紐付けは実行しない。
    -- FK・UNIQUEなどの整合性制約は通常どおり有効なまま維持する。
    ALTER TABLE public.reports DISABLE TRIGGER trigger_attach_souvenir_to_report;

    INSERT INTO public.reports (
        id, user_id, pet_id, hour_slot, gossip, group_master_id, previous_group_master_id,
        moved, behavior_type, behavior_label, interaction_count,
        energy_delta, curiosity_delta, sociality_delta, routine_delta,
        reason_json, rumor, created_at
    )
    SELECT
        md5('historical-report-reconstruction:' || pet_id::TEXT || ':' ||
            EXTRACT(EPOCH FROM at)::BIGINT::TEXT)::UUID,
        user_id, pet_id, EXTRACT(HOUR FROM at AT TIME ZONE 'Asia/Tokyo')::INTEGER,
        gossip, group_master_id, NULL,
        FALSE, 'reconstructed', behavior_label, 0,
        0, 0, 0, 0,
        jsonb_build_object(
            'source', 'historical_report_reconstruction',
            'source_report_id', source_report_id,
            'source_created_at', source_created_at,
            'effects_applied', FALSE,
            'reconstructed_at', CURRENT_TIMESTAMP,
            'reconstruction_batch', 'incident-20260903-1800-through-20260904-1000'
        ),
        '[]'::JSONB, at
    FROM report_reconstruction_candidates
    ORDER BY pet_id, at
    ON CONFLICT (pet_id, created_at) DO NOTHING;

    GET DIAGNOSTICS inserted_count = ROW_COUNT;
    ALTER TABLE public.reports ENABLE TRIGGER trigger_attach_souvenir_to_report;
    IF inserted_count <> 680 THEN
        RAISE EXCEPTION 'Unexpected inserted count: %', inserted_count;
    END IF;
END;
$reconstruct$;
