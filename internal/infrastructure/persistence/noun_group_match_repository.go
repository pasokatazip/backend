package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type NounGroupMatchRepository struct {
	DB *sql.DB
}

func NewNounGroupMatchRepository(db *sql.DB) *NounGroupMatchRepository {
	return &NounGroupMatchRepository{DB: db}
}

func (r *NounGroupMatchRepository) Create(nounGroupMatch domain.NounGroupMatch) (domain.NounGroupMatch, error) {
	query := `
		INSERT INTO noun_group_matches (
			extracted_noun_id,
			group_master_id,
			keyword_score,
			vector_score,
			keyword_weight,
			match_score,
			match_reason,
			selected,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING
			id,
			extracted_noun_id,
			group_master_id,
			keyword_score,
			vector_score,
			keyword_weight,
			match_score,
			match_reason,
			selected,
			created_at
	`

	row := r.DB.QueryRow(
		query,
		nounGroupMatch.ExtractedNounID(),
		nounGroupMatch.GroupMasterID(),
		nounGroupMatch.KeywordScore(),
		nounGroupMatch.VectorScore(),
		nounGroupMatch.KeywordWeight(),
		nounGroupMatch.MatchScore(),
		nounGroupMatch.MatchReason(),
		nounGroupMatch.Selected(),
		nounGroupMatch.CreatedAt(),
	)

	return scanNounGroupMatch(row)
}

func (r *NounGroupMatchRepository) FindByExtractedNounID(extractedNounID domain.ExtractedNounID) ([]domain.NounGroupMatch, error) {
	query := `
		SELECT
			id,
			extracted_noun_id,
			group_master_id,
			keyword_score,
			vector_score,
			keyword_weight,
			match_score,
			match_reason,
			selected,
			created_at
		FROM noun_group_matches
		WHERE extracted_noun_id = $1
		ORDER BY match_score DESC, id
	`

	rows, err := r.DB.Query(query, int(extractedNounID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]domain.NounGroupMatch, 0)
	for rows.Next() {
		match, err := scanNounGroupMatch(rows)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

func (r *NounGroupMatchRepository) FindSelectedByExtractedNounID(extractedNounID domain.ExtractedNounID) (domain.NounGroupMatch, error) {
	query := `
		SELECT
			id,
			extracted_noun_id,
			group_master_id,
			keyword_score,
			vector_score,
			keyword_weight,
			match_score,
			match_reason,
			selected,
			created_at
		FROM noun_group_matches
		WHERE extracted_noun_id = $1
			AND selected = TRUE
		ORDER BY match_score DESC, id
		LIMIT 1
	`

	row := r.DB.QueryRow(query, int(extractedNounID))
	return scanNounGroupMatch(row)
}

type nounGroupMatchScanner interface {
	Scan(dest ...any) error
}

func scanNounGroupMatch(scanner nounGroupMatchScanner) (domain.NounGroupMatch, error) {
	var (
		id              int
		extractedNounID int
		groupMasterID   int
		keywordScore    float64
		vectorScore     float64
		keywordWeight   float64
		matchScore      float64
		matchReason     sql.NullString
		selected        bool
		createdAt       time.Time
	)

	if err := scanner.Scan(
		&id,
		&extractedNounID,
		&groupMasterID,
		&keywordScore,
		&vectorScore,
		&keywordWeight,
		&matchScore,
		&matchReason,
		&selected,
		&createdAt,
	); err != nil {
		return domain.NounGroupMatch{}, err
	}

	var matchReasonValue *string
	if matchReason.Valid {
		matchReasonValue = &matchReason.String
	}

	return domain.NewNounGroupMatch(
		domain.NounGroupMatchID(id),
		domain.ExtractedNounID(extractedNounID),
		domain.GroupMasterID(groupMasterID),
		keywordScore,
		vectorScore,
		keywordWeight,
		matchScore,
		matchReasonValue,
		selected,
		createdAt,
	), nil
}
