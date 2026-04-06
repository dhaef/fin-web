package main

import (
	"log"
	"os"

	"fin-web/internal/controller"
	"fin-web/internal/db"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	tiingoToken := os.Getenv("TIINGO_TOKEN")

	if tiingoToken == "" {
		log.Fatal("TIINGO_TOKEN is required")
	}

	if tiingoToken == "" {
		log.Fatal("tiingoToken is required")
	}

	DB, err := db.NewDbConnection(dbPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	api := controller.NewController(DB, tiingoToken)

	err = api.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
