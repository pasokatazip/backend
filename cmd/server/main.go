package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/pasokatazip/backend/internal/controllers"
	"github.com/pasokatazip/backend/internal/infrastructure/auth"
	"github.com/pasokatazip/backend/internal/infrastructure/database"
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

	// Notification
	createNotification := usecases.NewCreateNotification(notificationRepo)
	updateNotification := usecases.NewUpdateNotification(notificationRepo)
	findNotificationByUserID := usecases.NewFindNotificationByUserID(notificationRepo)
	notificationController := controllers.NewNotificationController(
		createNotification,
		updateNotification,
		findNotificationByUserID,
	)

	mux := router.NewRouter(userController, petController, postController, reportController, notificationController)
	handler := middleware.CORS(os.Getenv("CORS_ALLOWED_ORIGINS"))(mux)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
