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
			r.id,
			r.pet_id,
			r.hour_slot,
			COALESCE(r.gossip, ''),
			r.group_master_id,
			gm.display_name,
			r.behavior_type,
			r.behavior_label,
			r.created_at
		FROM reports r
		INNER JOIN group_masters gm ON gm.id = r.group_master_id
		WHERE r.pet_id = $1
			AND r.created_at >= $2
			AND r.created_at < $3
		ORDER BY r.hour_slot
	`

	rows, err := r.DB.Query(query, petID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReports(rows)
}

func (r *ReportRepository) FindAllByPetID(petID domain.PetID) ([]domain.Report, error) {
	query := `
		SELECT
			r.id,
			r.pet_id,
			r.hour_slot,
			COALESCE(r.gossip, ''),
			r.group_master_id,
			gm.display_name,
			r.behavior_type,
			r.behavior_label,
			r.created_at
		FROM reports r
		INNER JOIN group_masters gm ON gm.id = r.group_master_id
		WHERE r.pet_id = $1
		ORDER BY r.created_at DESC
	`

	rows, err := r.DB.Query(query, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanReports(rows)
}

func (r *ReportRepository) scanReports(rows *sql.Rows) ([]domain.Report, error) {
	reports := make([]domain.Report, 0)

	for rows.Next() {
		var (
			id            string
			petIDStr      string
			hourSlot      int
			gossip        string
			groupMasterID int
			groupName     string
			behaviorType  string
			behaviorLabel string
			createdAt     time.Time
		)

		if err := rows.Scan(
			&id,
			&petIDStr,
			&hourSlot,
			&gossip,
			&groupMasterID,
			&groupName,
			&behaviorType,
			&behaviorLabel,
			&createdAt,
		); err != nil {
			return nil, err
		}

		report, err := domain.NewPersistedReport(
			domain.ReportID(id),
			domain.PetID(petIDStr),
			hourSlot,
			&gossip,
			domain.GroupMasterID(groupMasterID),
			behaviorType,
			behaviorLabel,
			groupName,
			createdAt,
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
