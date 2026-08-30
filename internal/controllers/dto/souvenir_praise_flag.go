package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

type SouvenirPraiseFlagResponse struct {
	HasPraised bool       `json:"hasPraised"`
	ReportDate string     `json:"reportDate"`
	PraisedAt  *time.Time `json:"praisedAt"`
}

func NewSouvenirPraiseFlagResponse(
	output usecases.SouvenirPraiseFlagOutput,
) SouvenirPraiseFlagResponse {
	return SouvenirPraiseFlagResponse{
		HasPraised: output.HasPraised,
		ReportDate: output.ReportDate.Format("2006-01-02"),
		PraisedAt:  output.PraisedAt,
	}
}
