package usecases

import "github.com/pasokatazip/backend/internal/domain"

type FincodeSubscriptionStatus struct {
	Active         bool
	CustomerID     *string
	SubscriptionID *string
}

type GetFincodeSubscription struct {
	repo domain.UserRepository
}

func NewGetFincodeSubscription(repo domain.UserRepository) *GetFincodeSubscription {
	return &GetFincodeSubscription{repo: repo}
}

func (u *GetFincodeSubscription) Execute(userID domain.UserID) (FincodeSubscriptionStatus, error) {
	if !domain.IsValidUserID(userID) {
		return FincodeSubscriptionStatus{}, domain.ErrValidation
	}

	user, err := u.repo.FindByID(userID)
	if err != nil {
		return FincodeSubscriptionStatus{}, err
	}

	return FincodeSubscriptionStatus{
		Active:         user.Subsc(),
		CustomerID:     user.FincodeCustomerID(),
		SubscriptionID: user.FincodeSubscriptionID(),
	}, nil
}
