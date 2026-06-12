package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/pasokatazip/backend/internal/controllers"
	"github.com/pasokatazip/backend/internal/infrastructure/auth"
	"github.com/pasokatazip/backend/internal/infrastructure/database"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	expMin := 60
	if v := os.Getenv("JWT_EXP_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			expMin = n
		}
	}
	tokenGen := auth.NewJWTTokenGenerator(jwtSecret, expMin)

	// User
	createUser := usecases.NewCreateUser(userRepo, tokenGen)
	login := usecases.NewLogin(userRepo, tokenGen)
	userController := controllers.NewUserController(createUser, login)

	// Pet
	createPet := usecases.NewCreatePet(petRepo)
	petController := controllers.NewPetController(createPet)

	mux := router.NewRouter(userController, petController)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
