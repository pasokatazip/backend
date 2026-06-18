package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers/dto"
	"github.com/pasokatazip/backend/internal/domain"
	"github.com/pasokatazip/backend/internal/presenter"
	"github.com/pasokatazip/backend/internal/usecases"
)

type PostController struct {
	createPost  *usecases.CreatePost
	findByPetId *usecases.FindByPetIDPost
}

func NewPostController(createPost *usecases.CreatePost, findByPetIDPost *usecases.FindByPetIDPost) *PostController {
	return &PostController{createPost: createPost, findByPetId: findByPetIDPost}
}

func (c *PostController) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	post, err := c.createPost.Execute(req.ToUseCaseInput())
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	pr := presenter.NewCreatePostPresenter()
	output := pr.Output(post)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.NewCreatePostResponse(output))
}

func (c *PostController) FindByPetIDPost(w http.ResponseWriter, r *http.Request) {
	petID := r.PathValue("pet_id")

	if petID == "" {
		http.Error(w, "missing pet_id", http.StatusBadRequest)
		return
	}

	outputs, err := c.findByPetId.Execute(usecases.FindByPetIDPostInput{PetID: domain.PetID(petID)})
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to fetch posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(outputs)
}
