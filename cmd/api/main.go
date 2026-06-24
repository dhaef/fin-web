package main

import (
	"fmt"
	"log"
	"os"

	"fin-web/internal/controller"
	"fin-web/internal/db"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	tiingoToken := os.Getenv("TIINGO_TOKEN")
	port := os.Getenv("PORT")

	if dbPath == "" {
		log.Fatal("DB_PATH is required")
	}

	if tiingoToken == "" {
		log.Fatal("TIINGO_TOKEN is required")
	}

	if port == "" {
		port = "3000"
	}

	fmt.Println("hello")
	DB, err := db.NewDbConnection(dbPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	api := controller.NewController(DB, tiingoToken, port)

	err = api.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
