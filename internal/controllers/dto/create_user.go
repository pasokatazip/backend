package dto

import (
	"time"

	"github.com/pasokatazip/backend/internal/usecases"
)

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r CreateUserRequest) ToUseCaseInput() usecases.CreateUserInput {
	return usecases.CreateUserInput{
		Email:    r.Email,
		Password: r.Password,
	}
}

type CreateUserResponse struct {
	Token                 string    `json:"token"`
	ExpiresIn             int64     `json:"expires_in"`
	ID                    string    `json:"id"`
	Email                 string    `json:"email"`
	Subsc                 bool      `json:"subsc"`
	FincodeCustomerID     *string   `json:"fincode_customer_id,omitempty"`
	FincodeSubscriptionID *string   `json:"fincode_subscription_id,omitempty"`
	CreatedAt             time.Time `json:"createdAt"`
}

type UserResponse struct {
	ID                    string    `json:"id"`
	Email                 string    `json:"email"`
	Subsc                 bool      `json:"subsc"`
	FincodeCustomerID     *string   `json:"fincode_customer_id,omitempty"`
	FincodeSubscriptionID *string   `json:"fincode_subscription_id,omitempty"`
	CreatedAt             time.Time `json:"createdAt"`
}

func NewCreateUserResponse(output usecases.CreateUserOutput, token string, expiresIn int64) CreateUserResponse {
	return CreateUserResponse{
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

func NewUserResponse(output usecases.CreateUserOutput) UserResponse {
	return UserResponse{
		ID:                    output.ID,
		Email:                 output.Email,
		Subsc:                 output.Subsc,
		FincodeCustomerID:     output.FincodeCustomerID,
		FincodeSubscriptionID: output.FincodeSubscriptionID,
		CreatedAt:             output.CreatedAt,
	}
}
