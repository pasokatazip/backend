package router

import (
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
)

func NewRouter(
	userController *controllers.UserController,
	petController *controllers.PetController,
	postController *controllers.PostController,
	reportController *controllers.ReportController,
	notificationController *controllers.NotificationController,
	fincodeController *controllers.FincodeController,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /webhooks/fincode", fincodeController.Handle)

	mux.HandleFunc("/users", userController.Create)
	mux.HandleFunc("/users/login", userController.Login)

	mux.Handle("/pets", middleware.Auth(http.HandlerFunc(petController.Create)))

	mux.HandleFunc("POST /posts", postController.Create)
	mux.HandleFunc("GET /posts/{pet_id}", postController.FindByPetIDPost)

	mux.HandleFunc("GET /reports/{pet_id}", reportController.FindByToday)

	mux.Handle("GET /notifications", middleware.Auth(http.HandlerFunc(notificationController.FindByUserID)))
	mux.Handle("POST /notifications", middleware.Auth(http.HandlerFunc(notificationController.Create)))
	mux.Handle("PUT /notifications", middleware.Auth(http.HandlerFunc(notificationController.Update)))
	return mux
}
