package main

import (
	"log"
	"os"

	"fin-web/internal/bofa"
	"fin-web/internal/citi"
	"fin-web/internal/db"
	"fin-web/internal/schwab"
	"fin-web/internal/worker"
)

func main() {
	dbPath := os.Getenv("dbPath")
	dirPath := os.Getenv("dirPath")

	if dbPath == "" {
		log.Fatal("dbPath is required")
	}

	if dirPath == "" {
		log.Fatal("dirPath is required")
	}

	transactionsDB, err := db.NewDbConnection(dbPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	bw := worker.NewBaseWorker(transactionsDB, dirPath)

	providers := []worker.Provider{
		bofa.NewBofaProvider(transactionsDB),
		citi.NewCitiProvider(transactionsDB),
		schwab.NewSchwabProvider(transactionsDB),
	}

	for _, p := range providers {
		if err := bw.Process(p); err != nil {
			log.Printf("Error processing provider %s: %v", p.GetPrefix(), err)
		}
	}
}
