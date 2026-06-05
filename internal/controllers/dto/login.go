package dto

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string             `json:"token"`
	ExpiresIn int64              `json:"expires_in"`
	User      CreateUserResponse `json:"user"`
}
