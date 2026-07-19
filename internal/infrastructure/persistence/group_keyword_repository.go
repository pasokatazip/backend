package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type GroupKeywordRepository struct {
	DB *sql.DB
}

func NewGroupKeywordRepository(db *sql.DB) *GroupKeywordRepository {
	return &GroupKeywordRepository{DB: db}
}

func (r *GroupKeywordRepository) FindActive() ([]domain.GroupKeyword, error) {
	query := `
		SELECT
			id,
			group_master_id,
			keyword,
			normalized_keyword,
			weight,
			match_type,
			active,
			created_at,
			updated_at
		FROM group_keywords
		WHERE active = TRUE
		ORDER BY id
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	return scanGroupKeywords(rows)
}

func (r *GroupKeywordRepository) FindActiveByGroupMasterID(groupMasterID domain.GroupMasterID) ([]domain.GroupKeyword, error) {
	query := `
		SELECT
			id,
			group_master_id,
			keyword,
			normalized_keyword,
			weight,
			match_type,
			active,
			created_at,
			updated_at
		FROM group_keywords
		WHERE group_master_id = $1
			AND active = TRUE
		ORDER BY id
	`

	rows, err := r.DB.Query(query, int(groupMasterID))
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	return scanGroupKeywords(rows)
}

func (r *GroupKeywordRepository) FindByNormalizedKeyword(normalizedKeyword string) ([]domain.GroupKeyword, error) {
	query := `
		SELECT
			id,
			group_master_id,
			keyword,
			normalized_keyword,
			weight,
			match_type,
			active,
			created_at,
			updated_at
		FROM group_keywords
		WHERE normalized_keyword = $1
			AND active = TRUE
		ORDER BY weight DESC, id
	`

	rows, err := r.DB.Query(query, normalizedKeyword)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	return scanGroupKeywords(rows)
}

func (r *GroupKeywordRepository) FindCandidatesByNormalizedNoun(normalizedNoun string) ([]domain.GroupKeyword, error) {
	query := `
		SELECT
			id,
			group_master_id,
			keyword,
			normalized_keyword,
			weight,
			match_type,
			active,
			created_at,
			updated_at
		FROM group_keywords
		WHERE active = TRUE
			AND (
				normalized_keyword = $1
				OR (
					match_type IN ('partial', 'exact_or_partial')
					AND (
						$1 LIKE '%' || normalized_keyword || '%'
						OR normalized_keyword LIKE '%' || $1 || '%'
					)
				)
			)
		ORDER BY weight DESC, id
	`

	rows, err := r.DB.Query(query, normalizedNoun)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	return scanGroupKeywords(rows)
}

type groupKeywordScanner interface {
	Scan(dest ...any) error
}

func scanGroupKeyword(scanner groupKeywordScanner) (domain.GroupKeyword, error) {
	var (
		id                int
		groupMasterID     int
		keyword           string
		normalizedKeyword string
		weight            float64
		matchType         string
		active            bool
		createdAt         time.Time
		updatedAt         time.Time
	)

	if err := scanner.Scan(
		&id,
		&groupMasterID,
		&keyword,
		&normalizedKeyword,
		&weight,
		&matchType,
		&active,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.GroupKeyword{}, mapPersistenceError(err)
	}

	return domain.NewGroupKeyword(
		domain.GroupKeywordID(id),
		domain.GroupMasterID(groupMasterID),
		keyword,
		normalizedKeyword,
		weight,
		matchType,
		active,
		createdAt,
		updatedAt,
	), nil
}

func scanGroupKeywords(rows *sql.Rows) ([]domain.GroupKeyword, error) {
	keywords := make([]domain.GroupKeyword, 0)
	for rows.Next() {
		keyword, err := scanGroupKeyword(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		keywords = append(keywords, keyword)
	}

	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return keywords, nil
}
