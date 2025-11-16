package main

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	prHandlers "github.com/qwerty268/pull_request_service/internal/rest_api/pullrequests"
	teamsHandlers "github.com/qwerty268/pull_request_service/internal/rest_api/teams"
	userHandlers "github.com/qwerty268/pull_request_service/internal/rest_api/users"
	prUsecase "github.com/qwerty268/pull_request_service/internal/usecases/pullrequests"
	prStorage "github.com/qwerty268/pull_request_service/internal/usecases/pullrequests/storage"
	teamUsecase "github.com/qwerty268/pull_request_service/internal/usecases/teams"
	teamStorage "github.com/qwerty268/pull_request_service/internal/usecases/teams/storage"
	userUsecase "github.com/qwerty268/pull_request_service/internal/usecases/users"
	userStorage "github.com/qwerty268/pull_request_service/internal/usecases/users/storage"
	"github.com/qwerty268/pull_request_service/internal/utils"
)

func main() {
	dsn := "host=localhost port=5432 user=lev-demchenko dbname=postgres sslmode=disable"

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// Проверим соединение
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	prStorage := prStorage.NewStorage(db)
	teamStorage := teamStorage.NewStorage(db)
	userStorage := userStorage.NewStorage(db)

	prUsecase := prUsecase.NewUsecase(prStorage, teamStorage, userStorage)
	teamUsecase := teamUsecase.NewUsecase(teamStorage)
	userUsecase := userUsecase.NewUsecase(userStorage, prStorage)

	prHandlers := prHandlers.NewHandlers(prUsecase)
	teamsHandlers := teamsHandlers.NewHandlers(teamUsecase)
	userHandlers := userHandlers.NewUserHandlers(userUsecase)

	e := echo.New()
	e.Validator = utils.NewHTTPRequestValidator()

	prHandlers.RegisterHandlers(e)
	teamsHandlers.RegisterHandlers(e)
	userHandlers.RegisterHandlers(e)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
