package main

import (
	"log"
	"os"

	"fin-web/internal/controller"
	"fin-web/internal/db"
)

func main() {
	dbPath := os.Getenv("dbPath")
	transactionsDB, err := db.NewDbConnection(dbPath + "/transactions.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	netWorthDB, err := db.NewDbConnection(dbPath + "/net_worth.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	api := controller.NewController(transactionsDB, netWorthDB)

	err = api.Server.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
