package router

import (
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers"
)

func NewRouter(userController *controllers.UserController) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/users", userController.Create)
	return mux
}
