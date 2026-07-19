package persistence

import (
	"database/sql"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
)

type ExtractedNounRepository struct {
	DB *sql.DB
}

func NewExtractedNounRepository(db *sql.DB) *ExtractedNounRepository {
	return &ExtractedNounRepository{DB: db}
}

func (r *ExtractedNounRepository) Create(extractedNoun domain.ExtractedNoun) (domain.ExtractedNoun, error) {
	query := `
		INSERT INTO extracted_nouns (
			post_id,
			noun_text,
			normalized_noun,
			noun_embedding,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			post_id,
			noun_text,
			normalized_noun,
			noun_embedding,
			created_at
	`

	row := r.DB.QueryRow(
		query,
		extractedNoun.PostID(),
		extractedNoun.NounText(),
		extractedNoun.NormalizedNoun(),
		extractedNoun.NounEmbedding(),
		extractedNoun.CreatedAt(),
	)

	return scanExtractedNoun(row)
}

func (r *ExtractedNounRepository) FindByPostID(postID domain.PostID) ([]domain.ExtractedNoun, error) {
	query := `
		SELECT
			id,
			post_id,
			noun_text,
			normalized_noun,
			noun_embedding,
			created_at
		FROM extracted_nouns
		WHERE post_id = $1
		ORDER BY id
	`

	rows, err := r.DB.Query(query, postID)
	if err != nil {
		return nil, mapPersistenceError(err)
	}
	defer rows.Close()

	extractedNouns := make([]domain.ExtractedNoun, 0)
	for rows.Next() {
		extractedNoun, err := scanExtractedNoun(rows)
		if err != nil {
			return nil, mapPersistenceError(err)
		}
		extractedNouns = append(extractedNouns, extractedNoun)
	}

	if err := rows.Err(); err != nil {
		return nil, mapPersistenceError(err)
	}

	return extractedNouns, nil
}

type extractedNounScanner interface {
	Scan(dest ...any) error
}

func scanExtractedNoun(scanner extractedNounScanner) (domain.ExtractedNoun, error) {
	var (
		id             int
		postID         string
		nounText       string
		normalizedNoun string
		nounEmbedding  sql.NullString
		createdAt      time.Time
	)

	if err := scanner.Scan(
		&id,
		&postID,
		&nounText,
		&normalizedNoun,
		&nounEmbedding,
		&createdAt,
	); err != nil {
		return domain.ExtractedNoun{}, mapPersistenceError(err)
	}

	var nounEmbeddingValue *string
	if nounEmbedding.Valid {
		nounEmbeddingValue = &nounEmbedding.String
	}

	return domain.NewExtractedNoun(
		domain.ExtractedNounID(id),
		domain.PostID(postID),
		nounText,
		normalizedNoun,
		nounEmbeddingValue,
		createdAt,
	), nil
}
