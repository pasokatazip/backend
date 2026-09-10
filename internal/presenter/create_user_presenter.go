package presenter

import (
	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type CreateUserPresenter struct{}

func NewCreateUserPresenter() *CreateUserPresenter {
	return &CreateUserPresenter{}
}

func (p *CreateUserPresenter) CreateResponse(user domain.User, token string, expiresIn int64) dto.CreateUserResponse {
	return dto.NewCreateUserResponse(p.Output(user), token, expiresIn)
}

func (p *CreateUserPresenter) LoginResponse(user domain.User, token string, expiresIn int64) dto.LoginResponse {
	return dto.LoginResponse{
		Token:     token,
		ExpiresIn: expiresIn,
		User:      dto.NewUserResponse(p.Output(user)),
	}
}

func (p *CreateUserPresenter) Output(user domain.User) usecases.CreateUserOutput {
	return usecases.CreateUserOutput{
		ID:                    string(user.ID()),
		Email:                 user.Email(),
		Subsc:                 user.Subsc(),
		FincodeCustomerID:     user.FincodeCustomerID(),
		FincodeSubscriptionID: user.FincodeSubscriptionID(),
		CreatedAt:             user.CreatedAt(),
	}
}
