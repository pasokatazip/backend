package fincode

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/timeutil"
)

const fincodeDateTimeLayout = "2006/01/02 15:04:05"

type createCardSessionRequest struct {
	Expire                 string `json:"expire"`
	ShopServiceName        string `json:"shop_service_name,omitempty"`
	GuideMailSendFlag      string `json:"guide_mail_send_flag"`
	CompletionMailSendFlag string `json:"completion_mail_send_flag"`
	CustomerID             string `json:"customer_id"`
	SuccessURL             string `json:"success_url,omitempty"`
	CancelURL              string `json:"cancel_url,omitempty"`
}

type createCardSessionResponse struct {
	ID      string `json:"id"`
	LinkURL string `json:"link_url"`
	Expire  string `json:"expire"`
}

func (c *Client) CreateCardSession(
	ctx context.Context,
	input domain.FincodeCardSessionInput,
) (domain.FincodeCardSession, error) {
	if input.CustomerID == "" || input.ExpiresAt.IsZero() {
		return domain.FincodeCardSession{}, domain.ErrValidation
	}

	var response createCardSessionResponse
	err := c.doJSON(
		ctx,
		"POST",
		"/v1/card_sessions",
		"",
		createCardSessionRequest{
			Expire:                 input.ExpiresAt.In(timeutil.LocationJST()).Format(fincodeDateTimeLayout),
			ShopServiceName:        input.ShopServiceName,
			GuideMailSendFlag:      "0",
			CompletionMailSendFlag: "0",
			CustomerID:             input.CustomerID,
			SuccessURL:             input.SuccessURL,
			CancelURL:              input.CancelURL,
		},
		&response,
	)
	if err != nil {
		return domain.FincodeCardSession{}, fmt.Errorf("create fincode card session: %w", err)
	}
	if response.ID == "" || !validHTTPURL(response.LinkURL) {
		return domain.FincodeCardSession{}, fmt.Errorf("%w: create fincode card session: invalid response", domain.ErrExternalService)
	}

	expiresAt := input.ExpiresAt
	if parsed, err := time.ParseInLocation(fincodeDateTimeLayout, response.Expire, timeutil.LocationJST()); err == nil {
		expiresAt = parsed
	}

	return domain.FincodeCardSession{
		ID:        response.ID,
		LinkURL:   response.LinkURL,
		ExpiresAt: expiresAt,
	}, nil
}

func validHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http")
}
