package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/pasokatazip/backend/docs"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/pasokatazip/backend/internal/controllers"
	"github.com/pasokatazip/backend/internal/infrastructure/auth"
	"github.com/pasokatazip/backend/internal/infrastructure/database"
	"github.com/pasokatazip/backend/internal/infrastructure/fincode"
	"github.com/pasokatazip/backend/internal/infrastructure/middleware"
	"github.com/pasokatazip/backend/internal/infrastructure/persistence"
	"github.com/pasokatazip/backend/internal/router"
	"github.com/pasokatazip/backend/internal/usecases"
	"github.com/pasokatazip/backend/internal/usecases/onetime"
	"github.com/pasokatazip/backend/internal/usecases/subsc"
)

const (
	defaultFincodePaymentAmount      = 999
	defaultFincodePurchaseSuccessURL = "http://localhost:3000/Subscription"
)

// @title PETYO-YO API
// @version 1.0
// @description PETYO-YO backend API documentation
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 「Bearer {JWT}」の形式で入力してください
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
	petExperienceRepo := persistence.NewPetExperienceRepository(db)
	petExperienceEventRepo := persistence.NewPetExperienceEventRepository(db)
	evolutionStageRepo := persistence.NewEvolutionStageRepository(db)
	petEvolutionRepo := persistence.NewPetEvolutionRepository(db)
	notificationRepo := persistence.NewNotificationRepository(db)
	simulationRepo := persistence.NewPetSimulationRepository(db)
	departureRepo := persistence.NewPetDepartureRepository(db)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	postRepo := persistence.NewPostRepository(db)

	reportRepo := persistence.NewReportRepository(db)

	expMin := 2880
	if v := os.Getenv("JWT_EXP_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			expMin = n
		}
	}
	jwtService := auth.NewJWTTokenGenerator(jwtSecret, expMin)
	passwordHasher := auth.NewBCryptPasswordHasher()

	// User
	createUser := usecases.NewCreateUser(userRepo, jwtService, passwordHasher)
	runPetDepartureCheck := usecases.NewRunPetDepartureCheck(departureRepo)
	login := usecases.NewLogin(userRepo, petRepo, jwtService, jwtService, passwordHasher, runPetDepartureCheck)
	updateUserEmail := usecases.NewUpdateUserEmail(userRepo, passwordHasher)
	updateUserPassword := usecases.NewUpdateUserPassword(userRepo, passwordHasher)
	userController := controllers.NewUserController(createUser, login, updateUserEmail, updateUserPassword)

	createPost := usecases.NewCreatePost(postRepo)
	findByPetIDPost := usecases.NewFindByPetIDPost(postRepo)
	postController := controllers.NewPostController(createPost, findByPetIDPost)

	// Pet
	createPet := usecases.NewCreatePet(petRepo)
	findMyActivePet := usecases.NewFindMyActivePet(
		petRepo,
		persistence.NewGroupMasterRepository(db),
		departureRepo,
	)
	findAllPets := usecases.NewFindAllPets(petRepo)
	findHistoryPets := usecases.NewFindHistoryPets(petRepo)
	updatePetProfile := usecases.NewUpdatePetProfile(petRepo)
	updatePetDepartureStatus := usecases.NewUpdatePetDepartureStatus(departureRepo)
	petController := controllers.NewPetController(
		createPet,
		findMyActivePet,
		findAllPets,
		findHistoryPets,
		updatePetProfile,
		updatePetDepartureStatus,
	)
	findActivePetEvolutionHistory := usecases.NewFindActivePetEvolutionHistory(petRepo, evolutionStageRepo, petEvolutionRepo)
	activePetEvolutionHistoryController := controllers.NewActivePetEvolutionHistoryController(findActivePetEvolutionHistory)
	findCurrentPetEvolutionStatus := usecases.NewFindCurrentPetEvolutionStatus(
		petRepo,
		petExperienceRepo,
		evolutionStageRepo,
		persistence.NewEvolutionRuleRepository(db),
		petEvolutionRepo,
	)
	currentPetEvolutionStatusController := controllers.NewCurrentPetEvolutionStatusController(findCurrentPetEvolutionStatus)
	findPetGrowthRecord := usecases.NewFindPetGrowthRecord(petRepo, evolutionStageRepo, petExperienceRepo, petExperienceEventRepo, petEvolutionRepo)
	petGrowthRecordController := controllers.NewPetGrowthRecordController(findPetGrowthRecord)

	findByDateReport := usecases.NewFindByDate(reportRepo)
	findAllReportsByPetID := usecases.NewFindAllReportsByPetID(reportRepo)
	findSubscriptionReports := usecases.NewFindSubscriptionReports(reportRepo, petRepo)
	reportController := controllers.NewReportController(findByDateReport, findAllReportsByPetID, findSubscriptionReports)

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
	fincodePaymentAmount := envPositiveIntOrDefault(
		"FINCODE_PAYMENT_AMOUNT", defaultFincodePaymentAmount,
	)
	var (
		subscriptionController *controllers.SubscriptionController
		purchaseController     *controllers.PurchaseController
		cardRegistration       controllers.HandleCardRegistUsecase
		subscRegistration      controllers.HandleSubscriptionRegistUsecase
		subscCancel            controllers.HandleSubscriptionCancelUsecase
		webhookSignature       string
	)

	switch billingMode := envOrDefault("FINCODE_BILLING_MODE", "one_time"); billingMode {
	case "one_time":
		startPurchase := onetime.NewStartFincodePurchase(
			userRepo, ensureFincodeCustomer, fincodeClient, "",
			envOrDefault("FINCODE_PURCHASE_SUCCESS_URL", defaultFincodePurchaseSuccessURL),
			30*time.Minute,
		)
		confirmPurchase := onetime.NewConfirmFincodePurchase(
			userRepo, fincodeClient, fincodeClient, fincodePaymentAmount,
		)
		purchaseController = controllers.NewPurchaseController(
			startPurchase, confirmPurchase,
		)
	case "subscription":
		startSubscription := subsc.NewStartFincodeSubscription(
			userRepo, ensureFincodeCustomer, fincodeClient, "", 30*time.Minute,
		)
		cancelSubscription := subsc.NewCancelFincodeSubscription(userRepo, fincodeClient)
		getSubscription := subsc.NewGetFincodeSubscription(userRepo)
		subscriptionController = controllers.NewSubscriptionController(
			startSubscription, cancelSubscription, getSubscription,
		)
		cardRegistration = subsc.NewCardRegistration(
			userRepo, fincodeClient, requiredEnv("FINCODE_PLAN_ID"),
		)
		subscRegistration = subsc.NewSubscRegistration(userRepo)
		subscCancel = subsc.NewSubscCancel(userRepo)
		webhookSignature = requiredEnv("FINCODE_WEBHOOK_SIGNATURE")
	default:
		log.Fatalf("FINCODE_BILLING_MODE must be one_time or subscription, got %q", billingMode)
	}

	// fincode Webhook
	fincodeController := controllers.NewWebhookController(
		cardRegistration,
		subscRegistration,
		subscCancel,
		webhookSignature,
	)

	mux := router.NewRouter(
		userController,
		petController,
		activePetEvolutionHistoryController,
		currentPetEvolutionStatusController,
		petGrowthRecordController,
		postController,
		reportController,
		notificationController,
		fincodeController,
		subscriptionController,
		purchaseController,
		simulationController,
	)

	mux.Handle("/docs/", httpSwagger.WrapHandler)

	handler := middleware.CORS(os.Getenv("CORS_ALLOWED_ORIGINS"))(mux)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return value
}

func envPositiveIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		log.Fatalf("%s must be a positive integer, got %q", key, value)
	}
	return parsed
}
