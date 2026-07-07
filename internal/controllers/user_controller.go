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
	createUser     *usecases.CreateUser
	login          *usecases.Login
	updateEmail    *usecases.UpdateUserEmail
	updatePassword *usecases.UpdateUserPassword
}

func NewUserController(
	createUser *usecases.CreateUser,
	login *usecases.Login,
	updateEmail *usecases.UpdateUserEmail,
	updatePassword *usecases.UpdateUserPassword,
) *UserController {
	return &UserController{
		createUser:     createUser,
		login:          login,
		updateEmail:    updateEmail,
		updatePassword: updatePassword,
	}
}

// Create ユーザーを新規登録します。
// @Summary ユーザー登録
// @Description メールアドレスとパスワードを使用して新規ユーザーを作成し、認証トークンを返します。
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "ユーザー情報"
// @Success 201 {object} dto.CreateUserResponse "登録成功"
// @Failure 400 {string} string "リクエスト不正"
// @Failure 405 {string} string "許可されていないメソッド"
// @Failure 500 {string} string "サーバーエラー"
// @Router /users [post]
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

// Login ユーザーを認証します。
// @Summary ログイン
// @Description メールアドレスとパスワード、または Authorization ヘッダーのトークンを使用して認証します。
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest false "ログイン情報（トークン認証時は省略可能）"
// @Param Authorization header string false "Bearer トークン"
// @Success 200 {object} dto.LoginResponse "ログイン成功"
// @Failure 400 {string} string "認証情報不足またはリクエスト不正"
// @Failure 401 {string} string "認証失敗"
// @Failure 405 {string} string "許可されていないメソッド"
// @Router /users/login [post]
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

// UpdateEmail changes a user's email address after password verification.
// @Summary メールアドレス変更
// @Tags users
// @Accept json
// @Param request body dto.UpdateUserEmailRequest true "変更内容"
// @Produce json
// @Success 200 {object} dto.UpdateUserResponse
// @Failure 400 {string} string "入力不備"
// @Failure 401 {string} string "認証失敗"
// @Failure 409 {string} string "メールアドレス重複"
// @Router /users/email [put]
func (c *UserController) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateUserEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := c.updateEmail.Execute(req.ToUseCaseInput()); err != nil {
		writeUpdateUserError(w, err)
		return
	}
	writeUpdateUserSuccess(w)
}

// UpdatePassword changes a user's password after current-password verification.
// @Summary パスワード変更
// @Tags users
// @Accept json
// @Param request body dto.UpdateUserPasswordRequest true "変更内容"
// @Produce json
// @Success 200 {object} dto.UpdateUserResponse
// @Failure 400 {string} string "入力不備"
// @Failure 401 {string} string "認証失敗"
// @Router /users/password [put]
func (c *UserController) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateUserPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := c.updatePassword.Execute(req.ToUseCaseInput()); err != nil {
		writeUpdateUserError(w, err)
		return
	}
	writeUpdateUserSuccess(w)
}

func writeUpdateUserSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.UpdateUserResponse{Message: "successful"})
}

func writeUpdateUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrUnauthorized):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, domain.ErrAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "failed to update user", http.StatusInternalServerError)
	}
}
