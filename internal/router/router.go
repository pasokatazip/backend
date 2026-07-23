package router

import (
	"net/http"

	"github.com/pasokatazip/backend/internal/controllers"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
)

func NewRouter(
	userController *controllers.UserController,
	petController *controllers.PetController,
	activePetEvolutionHistoryController *controllers.ActivePetEvolutionHistoryController,
	currentPetEvolutionStatusController *controllers.CurrentPetEvolutionStatusController,
	petGrowthRecordController *controllers.PetGrowthRecordController,
	postController *controllers.PostController,
	reportController *controllers.ReportController,
	notificationController *controllers.NotificationController,
	fincodeController *controllers.FincodeController,
	subscriptionController *controllers.SubscriptionController,
	simulationController *controllers.SimulationController,
) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /webhooks/fincode", fincodeController.Handle)
	mux.Handle("POST /subscriptions/checkout", middleware.Auth(http.HandlerFunc(subscriptionController.Start)))
	mux.Handle("GET /subscriptions", middleware.Auth(http.HandlerFunc(subscriptionController.Get)))
	mux.Handle("DELETE /subscriptions", middleware.Auth(http.HandlerFunc(subscriptionController.Cancel)))

	mux.HandleFunc("POST /users", userController.Create)
	mux.HandleFunc("POST /users/login", userController.Login)
	mux.HandleFunc("PUT /users/email", userController.UpdateEmail)
	mux.HandleFunc("PUT /users/password", userController.UpdatePassword)

	mux.Handle("POST /pets", middleware.Auth(http.HandlerFunc(petController.Create)))
	mux.Handle("GET /pets/me", middleware.Auth(http.HandlerFunc(petController.Current)))
	mux.Handle("PATCH /pets/departure", middleware.Auth(http.HandlerFunc(petController.UpdateDepartureStatus)))
	mux.Handle("GET /subsc/history_pet", middleware.Auth(http.HandlerFunc(petController.History)))
	mux.Handle("GET /subsc/allPets", middleware.Premium(http.HandlerFunc(petController.All)))
	mux.Handle("PUT /subsc/pet/{pet_id}", middleware.Premium(http.HandlerFunc(petController.UpdateProfile)))
	mux.Handle("GET /pets/evolutions", middleware.Auth(http.HandlerFunc(activePetEvolutionHistoryController.Find)))
	mux.Handle("GET /pets/evolution-status", middleware.Auth(http.HandlerFunc(currentPetEvolutionStatusController.Find)))
	mux.Handle("GET /subsc/pets/{pet_id}/evolutions", middleware.Premium(http.HandlerFunc(petGrowthRecordController.FindByPetID)))
	mux.Handle("GET /subsc/growth_records/{pet_id}", middleware.Premium(http.HandlerFunc(petGrowthRecordController.FindByPetID)))

	mux.HandleFunc("POST /posts", postController.Create)
	mux.HandleFunc("GET /posts/{pet_id}", postController.FindByPetIDPost)

	mux.HandleFunc("GET /reports/{pet_id}", reportController.FindByDate)
	mux.Handle("GET /subsc/reports/{pet_id}", middleware.Premium(http.HandlerFunc(reportController.FindAllByPetID)))

	mux.HandleFunc("POST /simulations/hourly", simulationController.RunHourly)

	mux.Handle("GET /notifications", middleware.Auth(http.HandlerFunc(notificationController.FindByUserID)))
	mux.Handle("POST /notifications", middleware.Auth(http.HandlerFunc(notificationController.Create)))
	mux.Handle("PUT /notifications", middleware.Auth(http.HandlerFunc(notificationController.Update)))
	return mux
}
