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
	dbPath := os.Getenv("DB_PATH")
	dirPath := os.Getenv("DIR_PATH")

	if dirPath == "" {
		log.Fatal("DIR_PATH is required")
	}

	if dirPath == "" {
		log.Fatal("dirPath is required")
	}

	DB, err := db.NewDbConnection(dbPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	bw := worker.NewBaseWorker(DB, dirPath)

	providers := []worker.Provider{
		bofa.NewBofaProvider(DB),
		citi.NewCitiProvider(DB),
		schwab.NewSchwabProvider(DB),
	}

	for _, p := range providers {
		if err := bw.Process(p); err != nil {
			log.Printf("Error processing provider %s: %v", p.GetPrefix(), err)
		}
	}
}
