package citi

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"fin-web/internal/model"

	"github.com/google/uuid"
)

type Worker struct {
	DB          *sql.DB
	DirPath     string
	CategoryMap map[string][]string
}

func NewWorker(db *sql.DB, dp string, cm map[string][]string) *Worker {
	return &Worker{
		DB:          db,
		DirPath:     dp,
		CategoryMap: cm,
	}
}

func (w *Worker) Normalize() error {
	filePaths, err := w.getFilePaths()
	if err != nil {
		return err
	}

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Error:", err)
			break
		}
		defer file.Close()

		// skip the first 5 lines of the csv
		bufferedReader := bufio.NewReader(file)
		for range 5 {
			bufferedReader.ReadString('\n') // Discard 5 lines
		}

		reader := csv.NewReader(bufferedReader)
		records, err := reader.ReadAll()
		if err != nil {
			fmt.Println("Error:", err)
			break
		}

		transactions := []model.Transaction{}
		for _, r := range records[1:] {
			ID := uuid.NewString()

			var amount float64
			if r[2] != "" {
				a, err := parseAmount(r[2])
				if err != nil {
					fmt.Printf("failed to parse amount: %s, err: %v\n", r[1], err)
				}

				amount = a
			} else if r[3] != "" {
				a, err := parseAmount(r[3])
				if err != nil {
					fmt.Printf("failed to parse amount: %s, err: %v\n", r[1], err)
				}

				amount = -a
			}

			var cc sql.NullString
			normalizedName := strings.ToLower(r[1])
			normalizedCategory := strings.ToLower(r[4])
		out:
			for cat, vals := range w.CategoryMap {
				for _, val := range vals {
					if strings.Contains(normalizedName, val) || strings.Contains(normalizedCategory, val) {
						cc = sql.NullString{
							Valid:  true,
							String: cat,
						}
						break out
					}
				}
			}

			t, err := time.Parse("Jan 02, 2006", r[0])
			if err != nil {
				fmt.Println("Error:", err)
			}

			transactions = append(transactions, model.Transaction{
				ID:             ID,
				Name:           r[1],
				Source:         "citi",
				Account:        "citi",
				Date:           t.Format("2006-01-02"),
				Amount:         amount,
				CustomCategory: cc,
				Category:       r[4],
			},
			)
		}

		for _, t := range transactions {
			err = model.CreateTransaction(w.DB, t)
			if err != nil {
				fmt.Printf("failed to create transaction %s, err: %v\n", t.Name, err)
				continue
			}
		}

		err = os.Remove(filePath)
		if err != nil {
			fmt.Println("Error deleting file:", err)
		}
	}

	return nil
}

func (w *Worker) getFilePaths() ([]string, error) {
	entries, err := os.ReadDir(w.DirPath)
	if err != nil {
		return []string{}, err
	}

	filePaths := []string{}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "From") {
			filePaths = append(filePaths, path.Join(w.DirPath, entry.Name()))
		}
	}

	return filePaths, nil
}

func parseAmount(amount string) (float64, error) {
	r := strings.NewReplacer("$", "", ",", "")
	cleanInput := r.Replace(amount)

	val, err := strconv.ParseFloat(cleanInput, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}
