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
}

func NewPetController(createPet *usecases.CreatePet, findHistoryPets *usecases.FindHistoryPets) *PetController {
	return &PetController{createPet: createPet, findHistoryPets: findHistoryPets}
}

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
