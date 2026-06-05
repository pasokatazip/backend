package presenter

import (
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/usecases"
)

type CreateUserPresenter struct{}

func NewCreateUserPresenter() *CreateUserPresenter {
	return &CreateUserPresenter{}
}

func (p *CreateUserPresenter) Output(user domain.User) usecases.CreateUserOutput {
	return usecases.CreateUserOutput{
		ID:        string(user.ID()),
		Email:     user.Email(),
		Subsc:     user.Subsc(),
		CreatedAt: user.CreatedAt(),
	}
}
