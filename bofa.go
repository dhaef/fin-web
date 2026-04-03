package bofa

import (
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
	DB      *sql.DB
	DirPath string
}

func NewWorker(db *sql.DB, dp string) *Worker {
	return &Worker{
		DB:      db,
		DirPath: dp,
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

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			fmt.Println("Error:", err)
			break
		}

		transactions := []model.Transaction{}
		for _, r := range records[1:] {
			ID := uuid.NewString()

			var amount float64
			if r[4] != "" {
				a, err := parseAmount(r[4])
				if err != nil {
					fmt.Printf("failed to parse amount: %s, err: %v\n", r[1], err)
				}

				amount = a
			}

			var cc sql.NullInt32
			normalizedName := strings.ToLower(r[2])

			categories, err := model.SearchCategories(
				w.DB,
				[]string{normalizedName},
			)
			if err != nil {
				fmt.Printf("failed to get custom category for: %s, err: %v", normalizedName, err)
			}

			if len(categories) == 0 {
				fmt.Printf("did not find any categories for: %s\n", normalizedName)
			} else {
				cc = sql.NullInt32{
					Valid: true,
					Int32: int32(categories[0].ID),
				}
			}

			t, err := time.Parse("01/02/2006", r[0])
			if err != nil {
				fmt.Println("Error:", err)
			}

			transactions = append(transactions, model.Transaction{
				ID:         ID,
				Name:       r[2],
				Source:     "bank_of_america",
				Account:    "bank_of_america",
				Date:       t.Format("2006-01-02"),
				Amount:     amount,
				CategoryID: cc,
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
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "bofa") {
			filePaths = append(filePaths, path.Join(w.DirPath, entry.Name()))
		}
	}

	return filePaths, nil
}

func parseAmount(amount string) (float64, error) {
	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return 0, err
	}

	return val * -1, nil
}
