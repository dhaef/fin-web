package main

import (
	"log"
	"os"

	"fin-web/internal/controller"
	"fin-web/internal/db"
)

func main() {
	dbPath := os.Getenv("dbPath")

	if dbPath == "" {
		log.Fatal("dbPath is required")
	}

	DB, err := db.NewDbConnection(dbPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	api := controller.NewController(DB)

	err = api.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
