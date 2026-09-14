package dto

import "time"

type SouvenirPraiseFlagResponse struct {
	HasPraised bool       `json:"hasPraised"`
	ReportDate string     `json:"reportDate"`
	PraisedAt  *time.Time `json:"praisedAt"`
}
