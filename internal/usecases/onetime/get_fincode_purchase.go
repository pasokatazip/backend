package onetime

import "github.com/pasokatazip/backend/internal/domain"

type FincodePurchaseStatus struct {
	Purchased  bool
	CustomerID *string
	PaymentID  *string
}

type GetFincodePurchase struct {
	repo domain.UserRepository
}

func NewGetFincodePurchase(repo domain.UserRepository) *GetFincodePurchase {
	return &GetFincodePurchase{repo: repo}
}

func (u *GetFincodePurchase) Execute(userID domain.UserID) (FincodePurchaseStatus, error) {
	if !domain.IsValidUserID(userID) {
		return FincodePurchaseStatus{}, domain.ErrValidation
	}
	user, err := u.repo.FindByID(userID)
	if err != nil {
		return FincodePurchaseStatus{}, err
	}
	return FincodePurchaseStatus{
		Purchased: user.Subsc(), CustomerID: user.FincodeCustomerID(),
		PaymentID: user.FincodeBillingID(),
	}, nil
}
