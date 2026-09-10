package presenter

import "github.com/pasokatazip/backend/internal/controllers/dto"

type UpdateUserPresenter struct{}

func NewUpdateUserPresenter() *UpdateUserPresenter {
	return &UpdateUserPresenter{}
}

func (p *UpdateUserPresenter) Success() dto.UpdateUserResponse {
	return dto.UpdateUserResponse{Message: "successful"}
}
