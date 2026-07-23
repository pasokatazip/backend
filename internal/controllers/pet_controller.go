package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/presenter"
	"github.com/pasokatazip/backend/internal/usecases"
)

type PetController struct {
	createPet       *usecases.CreatePet
	findMyActivePet *usecases.FindMyActivePet
	findAllPets     *usecases.FindAllPets
	findHistoryPets *usecases.FindHistoryPets
	updateProfile   *usecases.UpdatePetProfile
	updateDeparture *usecases.UpdatePetDepartureStatus
}

func NewPetController(
	createPet *usecases.CreatePet,
	findMyActivePet *usecases.FindMyActivePet,
	findAllPets *usecases.FindAllPets,
	findHistoryPets *usecases.FindHistoryPets,
	updateProfile *usecases.UpdatePetProfile,
	updateDeparture *usecases.UpdatePetDepartureStatus,
) *PetController {
	return &PetController{
		createPet:       createPet,
		findMyActivePet: findMyActivePet,
		findAllPets:     findAllPets,
		findHistoryPets: findHistoryPets,
		updateProfile:   updateProfile,
		updateDeparture: updateDeparture,
	}
}

// UpdateDepartureStatus updates the authenticated user's active pet to
// eligible or departed after validating the configured departure rules.
// @Summary ペットの旅立ち状態を更新
// @Description 自分のアクティブペットを eligible または departed に更新します。年齢、進化段階、旅立ち予定日を満たさない更新は拒否されます。
// @Tags pets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdatePetDepartureStatusRequest true "旅立ち状態"
// @Success 200 {object} usecases.UpdatePetDepartureStatusOutput "更新成功"
// @Failure 400 {string} string "条件またはリクエスト不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "アクティブペットが見つからない"
// @Failure 500 {string} string "サーバーエラー"
// @Router /pets/departure [patch]
func (c *PetController) UpdateDepartureStatus(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdatePetDepartureStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	output, err := c.updateDeparture.Execute(req.ToUseCaseInput(domain.UserID(userIDString)))
	if err != nil {
		writeDomainError(w, err, "failed to update pet departure status")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

// Current returns the authenticated user's active pet, current group, and departure readiness.
// @Summary 現在のペットと所属群れ取得
// @Description 認証中のユーザーのアクティブペット、現在の群れ、旅立ち状態を取得します。departure.can_depart が true の場合は旅立ち画面を表示できます。
// @Tags pets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.CurrentPetResponse "取得成功"
// @Failure 401 {string} string "認証が必要"
// @Failure 404 {string} string "アクティブペットが見つからない"
// @Failure 405 {string} string "許可されていないメソッド"
// @Failure 500 {string} string "サーバーエラー"
// @Router /pets/me [get]
func (c *PetController) Current(w http.ResponseWriter, r *http.Request) {
	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	output, err := c.findMyActivePet.Execute(usecases.FindMyActivePetInput{
		UserID: domain.UserID(userIDString),
	})
	if err != nil {
		writeDomainError(w, err, "failed to fetch current pet")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewCurrentPetResponse(output))
}

// All returns every pet owned by the authenticated premium user.
// @Summary ユーザーに紐づく全ペット取得
// @Description プレミアムユーザーのuser_idに紐づく全てのペットを取得します。
// @Tags pets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.AllPetListResponse "取得成功"
// @Failure 401 {string} string "認証が必要"
// @Failure 403 {string} string "サブスクリプションが必要"
// @Failure 405 {string} string "許可されていないメソッド"
// @Failure 500 {string} string "サーバーエラー"
// @Router /subsc/allPets [get]
func (c *PetController) All(w http.ResponseWriter, r *http.Request) {
	userIDString, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, domain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	userID := domain.UserID(userIDString)
	pets, err := c.findAllPets.Execute(usecases.FindAllPetsInput{UserID: userID})
	if err != nil {
		writeDomainError(w, err, "failed to fetch pets")
		return
	}

	pr := presenter.NewPetPresenter()
	outputs := make([]usecases.PetOutput, 0, len(pets))
	for _, pet := range pets {
		outputs = append(outputs, pr.Output(pet))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewAllPetListResponse(outputs))
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
		writeDomainError(w, err, "failed to create pet")
		return
	}

	pr := presenter.NewPetPresenter()
	output := pr.Output(pet)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewCreatePetResponse(output))
}

// History returns deleted pets owned by the authenticated user.
// @Summary ペット履歴取得
// @Description 認証中のユーザーに紐づく削除済みペットの一覧を取得します。
// @Tags pets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.PetListResponse "取得成功"
// @Failure 400 {string} string "ユーザーID不正"
// @Failure 401 {string} string "認証が必要"
// @Failure 405 {string} string "許可されていないメソッド"
// @Failure 500 {string} string "サーバーエラー"
// @Router /subsc/history_pet [get]
func (c *PetController) History(w http.ResponseWriter, r *http.Request) {
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
		writeDomainError(w, err, "failed to fetch history pets")
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
		writeDomainError(w, err, "failed to update pet")
		return
	}

	output := presenter.NewPetPresenter().Output(pet)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.NewUpdatePetProfileResponse(output))
}
