package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

type ReportRepository struct {
	DB *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{DB: db}
}

func (r *ReportRepository) FindByToday(
	petID domain.PetID,
) ([]domain.Report, error) {
	now := timeutil.NowJST()

	start := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0, 0, 0, 0,
		timeutil.LocationJST(),
	)

	end := start.Add(24 * time.Hour)

	query := `
		SELECT
			id,
			pet_id,
			hour_slot,
			gossip,
			group_master_id,
			behavior_type,
			behavior_label
		FROM reports
		WHERE pet_id = $1
			AND created_at >= $2
			AND created_at < $3
		ORDER BY hour_slot
	`

	rows, err := r.DB.Query(query, petID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []domain.Report

	for rows.Next() {
		var (
			id            string
			petIDStr      string
			hourSlot      int
			gossip        string
			groupMasterID int
			behaviorType  string
			behaviorLabel string
		)

		if err := rows.Scan(
			&id,
			&petIDStr,
			&hourSlot,
			&gossip,
			&groupMasterID,
			&behaviorType,
			&behaviorLabel,
		); err != nil {
			return nil, err
		}

		report, err := domain.NewReport(
			domain.ReportID(id),
			domain.PetID(petIDStr),
			hourSlot,
			&gossip,
			domain.GroupMasterID(groupMasterID),
			behaviorType,
			behaviorLabel,
		)
		if err != nil {
			return nil, err
		}

		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}
