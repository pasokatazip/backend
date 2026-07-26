package fincode

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pasokatazip/backend/internal/domain"
)

type listCardsResponse struct {
	List []struct {
		ID          string `json:"id"`
		CustomerID  string `json:"customer_id"`
		DefaultFlag string `json:"default_flag"`
	} `json:"list"`
}

func (c *Client) ListCards(ctx context.Context, customerID string) ([]domain.FincodeCard, error) {
	if customerID == "" {
		return nil, domain.ErrValidation
	}

	var response listCardsResponse
	if err := c.doJSON(
		ctx,
		http.MethodGet,
		"/v1/customers/"+url.PathEscape(customerID)+"/cards",
		"",
		nil,
		&response,
	); err != nil {
		return nil, fmt.Errorf("list fincode cards: %w", err)
	}

	cards := make([]domain.FincodeCard, 0, len(response.List))
	for _, card := range response.List {
		if card.ID == "" {
			continue
		}
		cards = append(cards, domain.FincodeCard{
			ID: card.ID, CustomerID: card.CustomerID, DefaultFlag: card.DefaultFlag,
		})
	}
	return cards, nil
}
