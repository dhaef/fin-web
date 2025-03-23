package main

import (
	"fin-web/internal/controller"
	"fin-web/internal/db"
	"log"
	"os"
)

func main() {
	dbPath := os.Getenv("dbPath")
	db, err := db.NewDbConnection(dbPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	api := controller.NewController(db)

	err = api.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
