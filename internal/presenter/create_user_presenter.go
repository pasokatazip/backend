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
	output := p.Output(user)
	return dto.CreateUserResponse{
		Token:                 token,
		ExpiresIn:             expiresIn,
		ID:                    output.ID,
		Email:                 output.Email,
		Subsc:                 output.Subsc,
		FincodeCustomerID:     output.FincodeCustomerID,
		FincodeSubscriptionID: output.FincodeSubscriptionID,
		CreatedAt:             output.CreatedAt,
	}
}

func (p *CreateUserPresenter) LoginResponse(user domain.User, token string, expiresIn int64) dto.LoginResponse {
	output := p.Output(user)
	return dto.LoginResponse{
		Token:     token,
		ExpiresIn: expiresIn,
		User: dto.UserResponse{
			ID:                    output.ID,
			Email:                 output.Email,
			Subsc:                 output.Subsc,
			FincodeCustomerID:     output.FincodeCustomerID,
			FincodeSubscriptionID: output.FincodeSubscriptionID,
			CreatedAt:             output.CreatedAt,
		},
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
