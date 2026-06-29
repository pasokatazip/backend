package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pasokatazip/backend/internal/controllers"
	"github.com/pasokatazip/backend/internal/infrastructure/auth"
	"github.com/pasokatazip/backend/internal/infrastructure/database"
	"github.com/pasokatazip/backend/internal/infrastructure/fincode"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/infrastructure/persistence"
	"github.com/pasokatazip/backend/internal/router"
	"github.com/pasokatazip/backend/internal/usecases"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := database.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := persistence.NewUserRepository(db)
	petRepo := persistence.NewPetRepository(db)
	notificationRepo := persistence.NewNotificationRepository(db)
	simulationRepo := persistence.NewPetSimulationRepository(db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	postRepo := persistence.NewPostRepository(db)

	reportRepo := persistence.NewReportRepository(db)

	expMin := 60
	if v := os.Getenv("JWT_EXP_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			expMin = n
		}
	}
	jwtService := auth.NewJWTTokenGenerator(jwtSecret, expMin)
	passwordHasher := auth.NewBCryptPasswordHasher()

	// User
	createUser := usecases.NewCreateUser(userRepo, jwtService, passwordHasher)
	login := usecases.NewLogin(userRepo, jwtService, jwtService, passwordHasher)
	userController := controllers.NewUserController(createUser, login)

	createPost := usecases.NewCreatePost(postRepo)
	findByPetIDPost := usecases.NewFindByPetIDPost(postRepo)
	postController := controllers.NewPostController(createPost, findByPetIDPost)

	// Pet
	createPet := usecases.NewCreatePet(petRepo)
	petController := controllers.NewPetController(createPet)

	findByTodayReport := usecases.NewFindByToDay(reportRepo)
	reportController := controllers.NewReportController(findByTodayReport)

	runHourlySimulation := usecases.NewRunHourlyPetSimulation(simulationRepo)
	simulationController := controllers.NewSimulationController(runHourlySimulation)

	// Notification
	createNotification := usecases.NewCreateNotification(notificationRepo)
	updateNotification := usecases.NewUpdateNotification(notificationRepo)
	findNotificationByUserID := usecases.NewFindNotificationByUserID(notificationRepo)
	notificationController := controllers.NewNotificationController(
		createNotification,
		updateNotification,
		findNotificationByUserID,
	)

	// fincode
	fincodeClient, err := fincode.NewClient(fincode.Config{
		BaseURL:   requiredEnv("FINCODE_API_BASE_URL"),
		SecretKey: requiredEnv("FINCODE_PRIVATE_KEY"),
	})
	if err != nil {
		log.Fatalf("failed to configure fincode client: %v", err)
	}

	ensureFincodeCustomer := usecases.NewEnsureFincodeCustomer(userRepo, fincodeClient)
	startSubscription := usecases.NewStartFincodeSubscription(
		userRepo,
		ensureFincodeCustomer,
		fincodeClient,
		"",
		30*time.Minute,
	)
	cancelSubscription := usecases.NewCancelFincodeSubscription(userRepo, fincodeClient)
	getSubscription := usecases.NewGetFincodeSubscription(userRepo)
	subscriptionController := controllers.NewSubscriptionController(
		startSubscription,
		cancelSubscription,
		getSubscription,
	)

	// fincode Webhook
	cardRegistration := usecases.NewCardRegistration(
		userRepo,
		fincodeClient,
		requiredEnv("FINCODE_PLAN_ID"),
	)
	subscRegistration := usecases.NewSubscRegistration(userRepo)
	subscCancel := usecases.NewSubscCancel(userRepo)
	fincodeController := controllers.NewWebhookController(
		cardRegistration,
		subscRegistration,
		subscCancel,
		requiredEnv("FINCODE_WEBHOOK_SIGNATURE"),
	)

	mux := router.NewRouter(
		userController,
		petController,
		postController,
		reportController,
		notificationController,
		fincodeController,
		subscriptionController,
    simulationController,
	)
	handler := middleware.CORS(os.Getenv("CORS_ALLOWED_ORIGINS"))(mux)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return value
}
