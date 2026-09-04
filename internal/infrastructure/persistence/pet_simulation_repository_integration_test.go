//go:build integration

package persistence

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pasokatazip/backend/internal/domain"
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
	visits, err := repo.FindRecentGroupVisitCountsForSimulation(simulatedAt)
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
	want := domain.PetGroupVisitCounts{
		"00000000-0000-0000-0000-000000000001": {1: 2, 2: 1},
	}
	// 同じ瞬間をUTCで渡してもJSTで渡しても結果が変わらない。
	for _, at := range []time.Time{simulatedAt, simulatedAt.UTC()} {
		visits, err = repo.FindRecentGroupVisitCountsForSimulation(at)
		if err != nil || !reflect.DeepEqual(visits, want) {
			t.Fatalf("at=%s visits=%v want=%v err=%v", at, visits, want, err)
		}
	}
}
