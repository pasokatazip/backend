package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/usecases"
)

type ReportPresenter struct{}

func NewReportPresenter() *ReportPresenter {
	return &ReportPresenter{}
}

func (p *ReportPresenter) Daily(output usecases.FindByDateReportOutput) dto.ReportsResponse {
	return dto.NewReportsResponse(output)
}

func (p *ReportPresenter) Subscription(output usecases.SubscriptionReportsOutput) dto.SubscriptionReportsResponse {
	return dto.NewSubscriptionReportsResponse(output)
}
