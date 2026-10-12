//go:build integration

package persistence

import (
	"context"
	"database/sql"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// 実際のPostgreSQLで検証し、スタブでは検出できないパラメータの型推論も通す。
// TEMP TABLEと単一接続に限定し、接続先の既存テーブルには書き込まない。
func simulationTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("integration tests require TEST_DATABASE_URL")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `
		SET search_path TO pg_temp;
		SET statement_timeout TO '5s';
		CREATE TEMP TABLE pets (id UUID PRIMARY KEY, is_deleted BOOLEAN, status TEXT);
		CREATE TEMP TABLE user_active_pets (pet_id UUID UNIQUE);
		CREATE TEMP TABLE pet_hourly_logs (
			pet_id UUID, group_master_id INTEGER, simulated_at TIMESTAMPTZ
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestBackfillReportsFromLogsPostgres(t *testing.T) {
	db := simulationTestDB(t)
	_, err := db.Exec(`
		ALTER TABLE pets ADD COLUMN user_id UUID;
		ALTER TABLE pet_hourly_logs
			ADD COLUMN report_material TEXT,
			ADD COLUMN ambient_event TEXT,
			ADD COLUMN stayed BOOLEAN DEFAULT TRUE,
			ADD COLUMN interaction_count INTEGER DEFAULT 0,
			ADD COLUMN energy_delta_applied NUMERIC DEFAULT 0,
			ADD COLUMN curiosity_delta_applied NUMERIC DEFAULT 0,
			ADD COLUMN sociality_delta_applied NUMERIC DEFAULT 0,
			ADD COLUMN routine_delta_applied NUMERIC DEFAULT 0;
		CREATE TEMP TABLE reports (
			id UUID PRIMARY KEY, user_id UUID NOT NULL, pet_id UUID NOT NULL,
			hour_slot INTEGER CHECK (hour_slot BETWEEN 0 AND 23), gossip VARCHAR(255),
			group_master_id INTEGER, previous_group_master_id INTEGER,
			moved BOOLEAN, behavior_type TEXT, behavior_label VARCHAR(255),
			interaction_count INTEGER, energy_delta INTEGER, curiosity_delta INTEGER,
			sociality_delta INTEGER, routine_delta INTEGER, reason_json JSONB, rumor JSONB,
			created_at TIMESTAMPTZ, UNIQUE (pet_id, created_at)
		);
		INSERT INTO pets VALUES (
			'00000000-0000-0000-0000-000000000001', false, 'departed',
			'00000000-0000-0000-0000-000000000002'
		);
		INSERT INTO pet_hourly_logs (pet_id, group_master_id, simulated_at, report_material)
		SELECT id, 1, at, '群れでゆっくり過ごしました' FROM pets CROSS JOIN (
			VALUES ('2026-09-03 17:00:00+09'::timestamptz),
			       ('2026-09-03 18:00:00+09'::timestamptz),
			       ('2026-09-03 19:00:00+09'::timestamptz),
			       ('2026-09-04 00:00:00+09'::timestamptz)
		) times(at);
		INSERT INTO reports (id, user_id, pet_id, hour_slot, gossip, rumor, created_at)
		SELECT id, user_id, id, 19, '既存レポートは保持する', '[]'::jsonb,
			'2026-09-03 19:00:00+09'::timestamptz FROM pets;
	`)
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile(filepath.Join("..", "..", "..", "scripts", "backfill_reports_from_logs.sql"))
	if err != nil {
		t.Fatal(err)
	}
	// 実行日によって結果が変わらないよう、補完対象期間の終了だけを固定する。
	query := strings.ReplaceAll(string(script), "date_trunc('hour', CURRENT_TIMESTAMP)",
		"TIMESTAMPTZ '2026-09-04 00:00:00+09'")
	for attempt := 0; attempt < 2; attempt++ {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
		var total, recovered, unchanged int
		err := db.QueryRow(`SELECT COUNT(*),
			COUNT(*) FILTER (WHERE hour_slot = 18
				AND gossip = '群れでゆっくり過ごしました'
				AND reason_json->>'source' = 'hourly_log_backfill' AND rumor = '[]'::jsonb),
			COUNT(*) FILTER (WHERE hour_slot = 19 AND gossip = '既存レポートは保持する')
			FROM reports`).Scan(&total, &recovered, &unchanged)
		if err != nil || total != 2 || recovered != 1 || unchanged != 1 {
			t.Fatalf("attempt=%d total=%d recovered=%d unchanged=%d err=%v",
				attempt, total, recovered, unchanged, err)
		}
	}
	// 旅立ったペットも履歴を復元でき、ログのない20〜23時や対象期間外は生成しない。
	var logCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM pet_hourly_logs").Scan(&logCount); err != nil || logCount != 4 {
		t.Fatalf("backfill changed hourly logs: count=%d err=%v", logCount, err)
	}
}

func TestRecentGroupVisitsPostgres(t *testing.T) {
	db := simulationTestDB(t)
	repo := NewPetSimulationRepository(db)
	simulatedAt := time.Date(2026, 9, 3, 18, 0, 0, 0, time.FixedZone("JST", 9*60*60))

	// データが空でもSQLの準備段階で失敗しないことを確認する。
	visits, err := repo.FindRecentGroupVisitCountsForSimulation(context.Background(), simulatedAt)
	if err != nil || len(visits) != 0 {
		t.Fatalf("empty database: visits=%v err=%v", visits, err)
	}

	_, err = db.Exec(`
		INSERT INTO pets VALUES
			('00000000-0000-0000-0000-000000000001', false, 'active'),
			('00000000-0000-0000-0000-000000000002', false, 'departed'),
			('00000000-0000-0000-0000-000000000003', true, 'active'),
			('00000000-0000-0000-0000-000000000004', false, 'active');
		INSERT INTO user_active_pets SELECT id FROM pets
			WHERE id <> '00000000-0000-0000-0000-000000000004';
		INSERT INTO pet_hourly_logs VALUES
			('00000000-0000-0000-0000-000000000001', 1, '2026-09-02 17:59:59+09'),
			('00000000-0000-0000-0000-000000000001', 1, '2026-09-02 18:00:00+09'),
			('00000000-0000-0000-0000-000000000001', 1, '2026-09-03 17:59:59+09'),
			('00000000-0000-0000-0000-000000000001', 2, '2026-09-03 17:00:00+09'),
			('00000000-0000-0000-0000-000000000001', 2, '2026-09-03 18:00:00+09');
		INSERT INTO pet_hourly_logs SELECT id, 1, '2026-09-03 17:00:00+09'::timestamptz
			FROM pets WHERE id <> '00000000-0000-0000-0000-000000000001';
	`)
	if err != nil {
		t.Fatal(err)
	}
	// 同じ瞬間をUTCで渡してもJSTで渡しても結果が変わらない。
	for _, at := range []time.Time{simulatedAt, simulatedAt.UTC()} {
		visits, err = repo.FindRecentGroupVisitCountsForSimulation(context.Background(), at)
		petVisits := visits["00000000-0000-0000-0000-000000000001"]
		if err != nil || math.Abs(petVisits[1]-2) > 0.0001 || math.Abs(petVisits[2]-math.Pow(0.5, 1.0/24)) > 0.0001 {
			t.Fatalf("at=%s visits=%v err=%v", at, visits, err)
		}
	}
}

func TestRecentStatusDeltaTotalsPostgres(t *testing.T) {
	db := simulationTestDB(t)
	repo := NewPetSimulationRepository(db)
	simulatedAt := time.Date(2026, 9, 3, 18, 0, 0, 0, time.FixedZone("JST", 9*60*60))

	_, err := db.Exec(`
		ALTER TABLE pet_hourly_logs
			ADD COLUMN energy_delta_applied NUMERIC DEFAULT 0,
			ADD COLUMN curiosity_delta_applied NUMERIC DEFAULT 0,
			ADD COLUMN sociality_delta_applied NUMERIC DEFAULT 0,
			ADD COLUMN routine_delta_applied NUMERIC DEFAULT 0;
		INSERT INTO pets VALUES ('00000000-0000-0000-0000-000000000001', false, 'active');
		INSERT INTO user_active_pets VALUES ('00000000-0000-0000-0000-000000000001');
		INSERT INTO pet_hourly_logs (
			pet_id, group_master_id, simulated_at,
			energy_delta_applied, curiosity_delta_applied, sociality_delta_applied, routine_delta_applied
		) VALUES
			('00000000-0000-0000-0000-000000000001', 1, '2026-09-02 18:00:00+09', 2, 4, -2, -4),
			('00000000-0000-0000-0000-000000000001', 1, '2026-09-03 18:00:00+09', 100, 100, 100, 100);
	`)
	if err != nil {
		t.Fatal(err)
	}

	totals, err := repo.FindRecentStatusDeltaTotalsForSimulation(context.Background(), simulatedAt)
	if err != nil {
		t.Fatal(err)
	}
	got := totals["00000000-0000-0000-0000-000000000001"]
	if math.Abs(got.Energy-1) > 0.0001 || math.Abs(got.Curiosity-2) > 0.0001 ||
		math.Abs(got.Sociality+1) > 0.0001 || math.Abs(got.Routine+2) > 0.0001 {
		t.Fatalf("totals=%+v", got)
	}
}

func TestSouvenirFallbackUsesDailyVisitedGroupPostgres(t *testing.T) {
	db := simulationTestDB(t)
	_, err := db.Exec(`
		SET search_path TO pg_temp, public;
		ALTER TABLE pet_hourly_logs ADD COLUMN id UUID;
		CREATE TEMP TABLE group_masters (
			id INTEGER PRIMARY KEY,
			display_name VARCHAR(100) NOT NULL
		);
		CREATE TEMP TABLE reports (
			id UUID PRIMARY KEY,
			pet_id UUID NOT NULL,
			group_master_id INTEGER,
			created_at TIMESTAMPTZ
		);
		CREATE TEMP TABLE souvenir_masters (
			id INTEGER PRIMARY KEY,
			display_name VARCHAR(100) NOT NULL,
			group_master_id INTEGER NOT NULL UNIQUE,
			active BOOLEAN NOT NULL
		);
		CREATE TEMP TABLE pet_souvenirs (
			id UUID PRIMARY KEY,
			pet_id UUID NOT NULL,
			souvenir_master_id INTEGER NOT NULL,
			pet_hourly_log_id UUID,
			report_id UUID,
			found_at TIMESTAMPTZ NOT NULL,
			found_on DATE NOT NULL,
			source_group_master_id INTEGER,
			note TEXT,
			reported_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);

		INSERT INTO group_masters VALUES
			(10, 'レポートの群れ'),
			(20, '読書の群れ'),
			(30, '音楽の群れ'),
			(96, '家の群れ');
		INSERT INTO souvenir_masters VALUES
			(3, 'ドアノブ', 96, TRUE),
			(20, 'しおり', 20, TRUE),
			(30, 'カスタネット', 30, TRUE);

		INSERT INTO reports VALUES (
			'10000000-0000-0000-0000-000000000001',
			'00000000-0000-0000-0000-000000000001',
			10,
			'2026-10-10 13:00:00+09'
		);
		INSERT INTO pet_hourly_logs (id, pet_id, group_master_id, simulated_at) VALUES
			('20000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 96, '2026-10-09 23:00:00+09'),
			('20000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 20, '2026-10-10 14:00:00+09'),
			('20000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', 30, '2026-10-10 15:00:00+09');
		INSERT INTO pet_souvenirs VALUES (
			'30000000-0000-0000-0000-000000000001',
			'00000000-0000-0000-0000-000000000001',
			3,
			NULL,
			'10000000-0000-0000-0000-000000000001',
			'2026-10-10 13:00:00+09',
			'2026-10-10',
			96,
			'レポートといっしょに、小さなおみやげを持ち帰りました。',
			'2026-10-10 13:00:00+09',
			'2026-10-10 13:00:00+09',
			'2026-10-10 13:00:00+09'
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	migration, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", "119_select_souvenir_from_daily_visits.sql"))
	if err != nil {
		t.Fatal(err)
	}
	upSQL, _, ok := strings.Cut(string(migration), "-- +goose Down")
	if !ok {
		t.Fatal("migration 119 does not contain a Down section")
	}
	beforeFunction, functionBlock, ok := strings.Cut(upSQL, "-- +goose StatementBegin")
	if !ok {
		t.Fatal("migration 119 does not contain a StatementBegin section")
	}
	functionSQL, afterFunction, ok := strings.Cut(functionBlock, "-- +goose StatementEnd")
	if !ok {
		t.Fatal("migration 119 does not contain a StatementEnd section")
	}
	for _, statement := range []string{beforeFunction, functionSQL, afterFunction} {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`CREATE TRIGGER trigger_attach_souvenir_to_report
		AFTER INSERT ON reports FOR EACH ROW
		EXECUTE FUNCTION public.attach_souvenir_to_report()`); err != nil {
		t.Fatal(err)
	}

	// 過去の誤ったドアノブも、取得時刻より後を含む当日の直近訪問群れへ補正される。
	var masterID, sourceGroupID int
	var note string
	err = db.QueryRow(`SELECT souvenir_master_id, source_group_master_id, note
		FROM pet_souvenirs WHERE id = '30000000-0000-0000-0000-000000000001'`).
		Scan(&masterID, &sourceGroupID, &note)
	if err != nil || masterID != 30 || sourceGroupID != 30 || note != "音楽の群れで、小さなおみやげを見つけたようです。" {
		t.Fatalf("corrected master=%d source=%d note=%q err=%v", masterID, sourceGroupID, note, err)
	}

	_, err = db.Exec(`
		INSERT INTO pet_hourly_logs (id, pet_id, group_master_id, simulated_at) VALUES
			('20000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000002', 20, '2026-10-11 10:00:00+09'),
			('20000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000002', 30, '2026-10-11 12:00:00+09'),
			('20000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000003', 20, '2026-10-11 10:00:00+09'),
			('20000000-0000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000003', 30, '2026-10-11 12:00:00+09'),
			('20000000-0000-0000-0000-000000000008', '00000000-0000-0000-0000-000000000004', 20, '2026-10-11 14:00:00+09');

		-- レポートの群れにおみやげがない場合は、当日の直近訪問群れを使う。
		INSERT INTO reports VALUES (
			'10000000-0000-0000-0000-000000000002',
			'00000000-0000-0000-0000-000000000002',
			10,
			'2026-10-11 13:00:00+09'
		);

		-- レポートの群れも当日訪問済みなら、直近群れより優先する。
		INSERT INTO reports VALUES (
			'10000000-0000-0000-0000-000000000003',
			'00000000-0000-0000-0000-000000000003',
			20,
			'2026-10-11 13:00:00+09'
		);

		-- レポート作成後にしか訪問しない群れは使わず、未来の履歴を捏造しない。
		INSERT INTO reports VALUES (
			'10000000-0000-0000-0000-000000000004',
			'00000000-0000-0000-0000-000000000004',
			10,
			'2026-10-11 13:00:00+09'
		);
	`)
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.Query(`SELECT report_id, souvenir_master_id, source_group_master_id
		FROM pet_souvenirs
		WHERE report_id IN (
			'10000000-0000-0000-0000-000000000002',
			'10000000-0000-0000-0000-000000000003',
			'10000000-0000-0000-0000-000000000004'
		)
		ORDER BY report_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	type selectedSouvenir struct {
		reportID, masterID, groupID string
	}
	var selected []selectedSouvenir
	for rows.Next() {
		var souvenir selectedSouvenir
		if err := rows.Scan(&souvenir.reportID, &souvenir.masterID, &souvenir.groupID); err != nil {
			t.Fatal(err)
		}
		selected = append(selected, souvenir)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 || selected[0].masterID != "30" || selected[0].groupID != "30" ||
		selected[1].masterID != "20" || selected[1].groupID != "20" {
		t.Fatalf("selected souvenirs=%+v", selected)
	}
}
