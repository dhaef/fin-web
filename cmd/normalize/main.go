package main

import (
	"encoding/json"
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

	categories, err := loadCategories()
	if err != nil {
		log.Fatal(err)
	}

	s := schwab.NewWorker(transactionsDB, dirPath, categories)
	err = s.Normalize()
	if err != nil {
		fmt.Println(err)
	}

	c := citi.NewWorker(transactionsDB, dirPath, categories)
	err = c.Normalize()
	if err != nil {
		fmt.Println(err)
	}
}

func loadCategories() (map[string][]string, error) {
	data, err := os.ReadFile("categories.json")
	if err != nil {
		return nil, err
	}

	var result map[string][]string

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
