package main

import (
	"fmt"
	"log"
	"os"

	"fin-web/internal/citi"
	"fin-web/internal/db"
	"fin-web/internal/schwab"
)

func main() {
	dbPath := os.Getenv("dbPath")
	dirPath := os.Getenv("dirPath")

	transactionsDB, err := db.NewDbConnection(dbPath + "/transactions.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	s := schwab.NewWorker(transactionsDB, dirPath)
	err = s.Normalize()
	if err != nil {
		fmt.Println(err)
	}

	c := citi.NewWorker(transactionsDB, dirPath)
	err = c.Normalize()
	if err != nil {
		fmt.Println(err)
	}
}
