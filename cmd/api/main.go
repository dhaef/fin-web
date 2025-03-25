package main

import (
	"fin-web/internal/controller"
	"fin-web/internal/db"
	"log"
	"os"
)

func main() {
	dbPath := os.Getenv("dbPath")
	transactionsDb, err := db.NewDbConnection(dbPath + "/transactions.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	netWorthDb, err := db.NewDbConnection(dbPath + "/net_worth.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	api := controller.NewController(transactionsDb, netWorthDb)

	err = api.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
