package main

import (
	"encoding/json"
	"log"
	"os"

	"fin-web/internal/db"
	"fin-web/internal/model"
)

func main() {
	dbPath := os.Getenv("dbPath")
	transactionsDB, err := db.NewDbConnection(dbPath + "/transactions.db")
	if err != nil {
		log.Fatal(err.Error())
	}

	fileBytes, err := os.ReadFile("categories.json")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	var result map[string]struct {
		Values   []string `json:"values"`
		Priority int      `json:"priority"`
	}

	err = json.Unmarshal(fileBytes, &result)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	for category, data := range result {
		categoryID := 0
		for i, v := range data.Values {
			if i == 0 {
				id, err := model.CreateCategory(transactionsDB, category, data.Priority)
				if err != nil {
					log.Fatalf("Error creating category: %v", err)
				}

				categoryID = id
			}

			_, err := model.CreateCategoryValue(transactionsDB, categoryID, v)
			if err != nil {
				log.Fatalf("Error creating category value: %v", err)
			}
		}
	}
}
