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
	// Environment variables
	dsn := requiredEnv("DATABASE_URL")
	jwtSecret := requiredEnv("JWT_SECRET")
	expMin := envIntOrDefault("JWT_EXP_MIN", 2880)
	corsAllowedOrigins := envOrDefault("CORS_ALLOWED_ORIGINS", "")
	fincodeBaseURL := requiredEnv("FINCODE_API_BASE_URL")
	fincodePrivateKey := requiredEnv("FINCODE_PRIVATE_KEY")
	fincodePaymentAmount := envPositiveIntOrDefault("FINCODE_PAYMENT_AMOUNT", defaultFincodePaymentAmount)
	billingMode := envOrDefault("FINCODE_BILLING_MODE", "one_time")
	fincodePurchaseSuccessURL := envOrDefault("FINCODE_PURCHASE_SUCCESS_URL", defaultFincodePurchaseSuccessURL)
	var fincodePlanID, webhookSignature string
	switch billingMode {
	case "one_time":
	case "subscription":
		fincodePlanID = requiredEnv("FINCODE_PLAN_ID")
		webhookSignature = requiredEnv("FINCODE_WEBHOOK_SIGNATURE")
	default:
		log.Fatalf("FINCODE_BILLING_MODE must be one_time or subscription, got %q", billingMode)
	}

	// Database
	db, err := database.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Repositories
	userRepo := persistence.NewUserRepository(db)
	petRepo := persistence.NewPetRepository(db)
	petExperienceRepo := persistence.NewPetExperienceRepository(db)
	petExperienceEventRepo := persistence.NewPetExperienceEventRepository(db)
	evolutionStageRepo := persistence.NewEvolutionStageRepository(db)
	petEvolutionRepo := persistence.NewPetEvolutionRepository(db)
	notificationRepo := persistence.NewNotificationRepository(db)
	simulationRepo := persistence.NewPetSimulationRepository(db)
	departureRepo := persistence.NewPetDepartureRepository(db)
	postRepo := persistence.NewPostRepository(db)
	reportRepo := persistence.NewReportRepository(db)
	groupMasterRepo := persistence.NewGroupMasterRepository(db)
	evolutionRuleRepo := persistence.NewEvolutionRuleRepository(db)
	petSouvenirRepo := persistence.NewPetSouvenirRepository(db)
	souvenirPraiseFlagRepo := persistence.NewSouvenirPraiseFlagRepository(db)

	// Infrastructure services: authentication and fincode
	jwtService := auth.NewJWTTokenGenerator(jwtSecret, expMin)
	passwordHasher := auth.NewBCryptPasswordHasher()

	fincodeClient, err := fincode.NewClient(fincode.Config{
		BaseURL:   fincodeBaseURL,
		SecretKey: fincodePrivateKey,
	})
	if err != nil {
		log.Fatalf("failed to configure fincode client: %v", err)
	}

	// Usecases: User
	createUser := usecases.NewCreateUser(userRepo, jwtService, passwordHasher)
	runPetDepartureCheck := usecases.NewRunPetDepartureCheck(departureRepo)
	login := usecases.NewLogin(userRepo, petRepo, jwtService, jwtService, passwordHasher, runPetDepartureCheck)
	updateUserEmail := usecases.NewUpdateUserEmail(userRepo, passwordHasher)
	updateUserPassword := usecases.NewUpdateUserPassword(userRepo, passwordHasher)

	// Usecases: Post
	createPost := usecases.NewCreatePost(postRepo, petRepo)
	findByPetIDPost := usecases.NewFindByPetIDPost(postRepo, petRepo)

	// Usecases: Pet
	createPet := usecases.NewCreatePet(petRepo)
	findMyActivePet := usecases.NewFindMyActivePet(
		petRepo,
		groupMasterRepo,
		departureRepo,
	)
	findAllPets := usecases.NewFindAllPets(petRepo)
	findHistoryPets := usecases.NewFindHistoryPets(petRepo)
	updatePetProfile := usecases.NewUpdatePetProfile(petRepo)
	updatePetDepartureStatus := usecases.NewUpdatePetDepartureStatus(departureRepo)

	// Usecases: Pet evolution and growth
	findActivePetEvolutionHistory := usecases.NewFindActivePetEvolutionHistory(petRepo, evolutionStageRepo, petEvolutionRepo)
	findCurrentPetEvolutionStatus := usecases.NewFindCurrentPetEvolutionStatus(
		petRepo,
		petExperienceRepo,
		evolutionStageRepo,
		evolutionRuleRepo,
		petEvolutionRepo,
	)
	findPetGrowthRecord := usecases.NewFindPetGrowthRecord(petRepo, evolutionStageRepo, petExperienceRepo, petExperienceEventRepo, petEvolutionRepo)

	// Usecases: Pet souvenir
	findLatestPetSouvenir := usecases.NewFindLatestPetSouvenir(petSouvenirRepo)
	findLatestHistoricalPetSouvenir := usecases.NewFindLatestHistoricalPetSouvenir(petSouvenirRepo)
	markSouvenirPraised := usecases.NewMarkSouvenirPraised(souvenirPraiseFlagRepo, reportRepo)

	// Usecases: Report
	findByDateReport := usecases.NewFindByDate(reportRepo, souvenirPraiseFlagRepo, petRepo)
	findAllReportsByPetID := usecases.NewFindAllReportsByPetID(reportRepo, petRepo)
	findSubscriptionReports := usecases.NewFindSubscriptionReports(
		reportRepo,
		petRepo,
		evolutionStageRepo,
		souvenirPraiseFlagRepo,
	)

	// Usecases: Simulation
	runHourlySimulation := usecases.NewRunHourlyPetSimulation(simulationRepo)

	// Usecases: Notification
	createNotification := usecases.NewCreateNotification(notificationRepo)
	updateNotification := usecases.NewUpdateNotification(notificationRepo)
	findNotificationByUserID := usecases.NewFindNotificationByUserID(notificationRepo)

	// Controllers
	userController := controllers.NewUserController(createUser, login, updateUserEmail, updateUserPassword)
	postController := controllers.NewPostController(createPost, findByPetIDPost)
	petController := controllers.NewPetController(
		createPet,
		findMyActivePet,
		findAllPets,
		findHistoryPets,
		updatePetProfile,
		updatePetDepartureStatus,
	)
	activePetEvolutionHistoryController := controllers.NewActivePetEvolutionHistoryController(findActivePetEvolutionHistory)
	currentPetEvolutionStatusController := controllers.NewCurrentPetEvolutionStatusController(findCurrentPetEvolutionStatus)
	latestPetSouvenirController := controllers.NewLatestPetSouvenirController(
		findLatestPetSouvenir,
		findLatestHistoricalPetSouvenir,
	)
	souvenirPraiseFlagController := controllers.NewSouvenirPraiseFlagController(
		markSouvenirPraised,
	)
	petGrowthRecordController := controllers.NewPetGrowthRecordController(findPetGrowthRecord)
	reportController := controllers.NewReportController(findByDateReport, findAllReportsByPetID, findSubscriptionReports)
	simulationController := controllers.NewSimulationController(runHourlySimulation)
	notificationController := controllers.NewNotificationController(
		createNotification,
		updateNotification,
		findNotificationByUserID,
	)

	// Usecases and controllers: fincode billing mode
	ensureFincodeCustomer := usecases.NewEnsureFincodeCustomer(userRepo, fincodeClient)
	var (
		subscriptionController *controllers.SubscriptionController
		purchaseController     *controllers.PurchaseController
		cardRegistration       controllers.HandleCardRegistUsecase
		subscRegistration      controllers.HandleSubscriptionRegistUsecase
		subscCancel            controllers.HandleSubscriptionCancelUsecase
	)

	switch billingMode {
	case "one_time":
		startPurchase := onetime.NewStartFincodePurchase(
			userRepo, ensureFincodeCustomer, fincodeClient, "",
			fincodePurchaseSuccessURL,
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
			userRepo, fincodeClient, fincodePlanID,
		)
		subscRegistration = subsc.NewSubscRegistration(userRepo)
		subscCancel = subsc.NewSubscCancel(userRepo)
	}

	// Controller: fincode webhook
	fincodeController := controllers.NewWebhookController(
		cardRegistration,
		subscRegistration,
		subscCancel,
		webhookSignature,
	)

	// Router and middleware
	mux := router.NewRouter(
		userController,
		petController,
		activePetEvolutionHistoryController,
		currentPetEvolutionStatusController,
		latestPetSouvenirController,
		souvenirPraiseFlagController,
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

	handler := middleware.CORS(corsAllowedOrigins)(mux)

	// HTTP server
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

// envIntOrDefault preserves the fallback for unset or invalid integer values.
func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
