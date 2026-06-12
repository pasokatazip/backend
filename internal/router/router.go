package router

import (
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers"
)

func NewRouter(
	userController *controllers.UserController,
	petController *controllers.PetController,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/users", userController.Create)
	mux.HandleFunc("/users/login", userController.Login)
	mux.HandleFunc("/pets", petController.Create)
	return mux
}