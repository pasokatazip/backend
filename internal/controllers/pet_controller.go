package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/presenter"
	"github.com/pasokatazip/backend/internal/usecases"
)

type PetController struct {
	createPet       *usecases.CreatePet
	findHistoryPets *usecases.FindHistoryPets
	updateProfile   *usecases.UpdatePetProfile
}

func NewPetController(
	createPet *usecases.CreatePet,
	findHistoryPets *usecases.FindHistoryPets,
	updateProfile *usecases.UpdatePetProfile,
) *PetController {
	return &PetController{
		createPet:       createPet,
		findHistoryPets: findHistoryPets,
		updateProfile:   updateProfile,
	}
}

// Create ペットを新規登録します。
// @Summary ペット登録
// @Description 認証中のユーザーに紐づくペットを作成します。
// @Tags pets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreatePetRequest true "ペット情報"
// @Success 201 {object} dto.CreatePetResponse "登録成功"
// @Failure 400 {string} string "リクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 405 {string} string "許可されていないメソッド"
// @Failure 500 {string} string "サーバーエラー"
// @Router /pets [post]
func (c *PetController) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreatePetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	userID := domain.UserID(userIDString)
	if !domain.IsValidUserID(userID) {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	pet, err := c.createPet.Execute(req.ToUseCaseInput(userID))

	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to create pet", http.StatusInternalServerError)
		return
	}

	pr := presenter.NewPetPresenter()
	output := pr.Output(pet)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewCreatePetResponse(output))
}

func (c *PetController) History(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	userID := domain.UserID(userIDString)
	if !domain.IsValidUserID(userID) {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	pets, err := c.findHistoryPets.Execute(usecases.FindHistoryPetsInput{UserID: userID})
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to fetch history pets", http.StatusInternalServerError)
		return
	}

	pr := presenter.NewPetPresenter()
	outputs := make([]usecases.PetOutput, 0, len(pets))
	for _, pet := range pets {
		outputs = append(outputs, pr.Output(pet))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewPetListResponse(outputs))
}

// UpdateProfile ペットの名前とカラーを変更します。
// @Summary ペットの名前・カラー変更
// @Description サブスクリプション契約中のユーザーが、自分のペットの名前とカラーを変更します。
// @Tags pets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param pet_id path string true "ペットID"
// @Param request body dto.UpdatePetProfileRequest true "変更内容"
// @Success 200 {object} dto.CreatePetResponse "変更成功"
// @Failure 400 {string} string "リクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 403 {string} string "サブスクリプションが必要"
// @Failure 404 {string} string "ペットが見つからない"
// @Failure 405 {string} string "許可されていないメソッド"
// @Failure 500 {string} string "サーバーエラー"
// @Router /subsc/pet/{pet_id} [put]
func (c *PetController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", http.MethodPut)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.UpdatePetProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	pet, err := c.updateProfile.Execute(req.ToUseCaseInput(
		domain.PetID(r.PathValue("pet_id")),
		domain.UserID(userIDString),
	))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, domain.ErrNotFound):
			http.Error(w, "pet not found", http.StatusNotFound)
		default:
			http.Error(w, "failed to update pet", http.StatusInternalServerError)
		}
		return
	}

	output := presenter.NewPetPresenter().Output(pet)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewUpdatePetProfileResponse(output))
}
