package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/presenter"
	"github.com/pasokatazip/backend/internal/usecases"
)

type UserController struct {
	createUser *usecases.CreateUser
	login      *usecases.Login
}

func NewUserController(createUser *usecases.CreateUser, login *usecases.Login) *UserController {
	return &UserController{createUser: createUser, login: login}
}

func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, token, expiresAt, err := c.createUser.Execute(req.ToUseCaseInput())
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	pr := presenter.NewCreateUserPresenter()
	output := pr.Output(user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewCreateUserResponse(output, token, int64(time.Until(expiresAt).Seconds())))
}

func (c *UserController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var token string
	var expiresAt time.Time
	var user domain.User
	var err error

	auth := r.Header.Get("Authorization")
	tokenFromHeader := ""
	if auth != "" {
		parts := strings.Fields(auth)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			tokenFromHeader = parts[1]
		}
	}

	if req.Email != "" && req.Password != "" {
		token, expiresAt, user, err = c.login.Execute(usecases.LoginInput{Email: req.Email, Password: req.Password})
	} else if tokenFromHeader != "" {
		token, expiresAt, user, err = c.login.ExecuteToken(tokenFromHeader)
	} else {
		http.Error(w, "email and password, or Authorization header with token required", http.StatusBadRequest)
		return
	}
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}
		// Token parse or user lookup errors
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	pr := presenter.NewCreateUserPresenter()
	userOutput := pr.Output(user)

	resp := dto.LoginResponse{
		Token:     token,
		ExpiresIn: int64(time.Until(expiresAt).Seconds()),
		User:      dto.NewUserResponse(userOutput),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
