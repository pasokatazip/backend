package onetime

import (
	"context"
	"strings"

	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type ConfirmFincodePurchase struct {
	repo           domain.UserRepository
	cardGateway    domain.FincodeCardGateway
	paymentGateway domain.FincodePaymentGateway
	amount         int
}

func NewConfirmFincodePurchase(
	repo domain.UserRepository,
	cardGateway domain.FincodeCardGateway,
	paymentGateway domain.FincodePaymentGateway,
	amount int,
) *ConfirmFincodePurchase {
	return &ConfirmFincodePurchase{
		repo: repo, cardGateway: cardGateway, paymentGateway: paymentGateway, amount: amount,
	}
}

func (u *ConfirmFincodePurchase) Execute(
	ctx context.Context,
	userID domain.UserID,
) (FincodePurchaseStatus, error) {
	if !domain.IsValidUserID(userID) || u.repo == nil || u.cardGateway == nil ||
		u.paymentGateway == nil || u.amount <= 0 {
		return FincodePurchaseStatus{}, domain.ErrValidation
	}

	user, err := u.repo.FindByID(userID)
	if err != nil {
		return FincodePurchaseStatus{}, err
	}
	if user.Subsc() {
		return purchaseStatus(user), nil
	}
	if user.FincodeCustomerID() == nil || strings.TrimSpace(*user.FincodeCustomerID()) == "" {
		return FincodePurchaseStatus{}, domain.ErrValidation
	}
	customerID := *user.FincodeCustomerID()

	cards, err := u.cardGateway.ListCards(ctx, customerID)
	if err != nil {
		return FincodePurchaseStatus{}, err
	}
	cardID := selectCard(cards)
	if cardID == "" {
		return FincodePurchaseStatus{}, domain.ErrNotFound
	}

	stableKey := usecases.FincodeIdempotencyKey("one-time:" + customerID)
	orderID := "ot_" + strings.ReplaceAll(stableKey, "-", "")[:27]
	payment, getErr := u.paymentGateway.GetPayment(ctx, orderID)
	if getErr != nil {
		payment, err = u.paymentGateway.CreatePayment(ctx, domain.FincodePaymentInput{
			ID: orderID, Amount: u.amount, IdempotencyKey: domain.NewUUIDString(),
		})
		if err != nil {
			payment, getErr = u.paymentGateway.GetPayment(ctx, orderID)
			if getErr != nil {
				return FincodePurchaseStatus{}, err
			}
		}
	}

	if !confirmPaymentSucceeded(payment.Status) {
		payment, err = u.paymentGateway.ExecutePayment(ctx, domain.FincodePaymentInput{
			ID: payment.ID, CustomerID: customerID, CardID: cardID,
			AccessID: payment.AccessID, IdempotencyKey: domain.NewUUIDString(),
		})
		if err != nil {
			return FincodePurchaseStatus{}, err
		}
	}
	if !confirmPaymentSucceeded(payment.Status) {
		return FincodePurchaseStatus{}, domain.ErrExternalService
	}
	if err := u.repo.UpdateFincodeBilling(user.ID(), payment.ID, true); err != nil {
		return FincodePurchaseStatus{}, err
	}

	return FincodePurchaseStatus{
		Purchased: true, CustomerID: user.FincodeCustomerID(), PaymentID: &payment.ID,
	}, nil
}

func selectCard(cards []domain.FincodeCard) string {
	for _, card := range cards {
		if card.DefaultFlag == "1" {
			return card.ID
		}
	}
	if len(cards) > 0 {
		return cards[0].ID
	}
	return ""
}

func purchaseStatus(user domain.User) FincodePurchaseStatus {
	return FincodePurchaseStatus{
		Purchased: user.Subsc(), CustomerID: user.FincodeCustomerID(),
		PaymentID: user.FincodeBillingID(),
	}
}

func confirmPaymentSucceeded(status string) bool {
	switch strings.ToUpper(status) {
	case "CAPTURED", "SUCCEEDED":
		return true
	default:
		return false
	}
}
