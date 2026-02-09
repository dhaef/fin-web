package schwab

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

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

type Transaction struct {
	Description string     `json:"Description"`
	Date        CustomDate `json:"Date"`
	Withdrawal  string     `json:"Withdrawal"`
	Deposit     string     `json:"Deposit"`
}

type Statement struct {
	FromDate            string        `json:"FromDate"`
	ToDate              string        `json:"ToDate"`
	PostedTransactions  []Transaction `json:"PostedTransactions"`
	PendingTransactions []Transaction `json:"PendingTransactions"`
}

func (w *Worker) Normalize() error {
	filePaths, err := w.getFilePaths()
	if err != nil {
		return err
	}

	for _, filePath := range filePaths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("failed to handle file: %v\n", err)
			continue
		}

		var statement Statement
		err = json.Unmarshal(data, &statement)
		if err != nil {
			fmt.Printf("failed to unmarshal json data: %v\n", err)
			continue
		}

		transactions := []model.Transaction{}
		for _, t := range statement.PostedTransactions {
			ID := uuid.NewString()

			var amount float64
			if t.Withdrawal != "" {
				a, err := parseAmount(t.Withdrawal)
				if err != nil {
					fmt.Printf("failed to parse amount: %s, err: %v\n", t.Description, err)
				}

				amount = a
			} else if t.Deposit != "" {
				a, err := parseAmount(t.Deposit)
				if err != nil {
					fmt.Printf("failed to parse amount: %s, err: %v\n", t.Description, err)
				}

				amount = -a
			}

			var cc sql.NullString
			normalizedName := strings.ToLower(t.Description)
			categories, err := model.SearchCategories(w.DB, []string{normalizedName})
			if err != nil {
				fmt.Printf("failed to get custom category for: %s, err: %v", normalizedName, err)
			}

			if len(categories) == 0 {
				fmt.Printf("did not find any categories for: %s", normalizedName)
			} else {
				cc = sql.NullString{
					Valid:  true,
					String: categories[0].Category,
				}
			}

			transactions = append(transactions, model.Transaction{
				ID:             ID,
				Name:           t.Description,
				Source:         "schwab",
				Account:        "schwab",
				Date:           t.Date.Format("2006-01-02"),
				Amount:         amount,
				CustomCategory: cc,
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
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "Checking") {
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
