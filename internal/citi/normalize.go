package citi

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"os"
	"strings"
	"time"

	"fin-web/internal/model"
	"fin-web/internal/util"

	"github.com/google/uuid"
)

type Provider struct {
	DB *sql.DB
}

func NewCitiProvider(db *sql.DB) *Provider {
	return &Provider{
		DB: db,
	}
}

func (p *Provider) GetPrefix() string {
	return "From"
}

func (p *Provider) ParseFile(filePath string) ([]model.Transaction, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Skip the first 5 lines of the Citi CSV
	bufferedReader := bufio.NewReader(file)
	for range 5 {
		bufferedReader.ReadString('\n')
	}

	reader := csv.NewReader(bufferedReader)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var transactions []model.Transaction
	for _, r := range records[1:] {
		var amount float64
		if r[2] != "" { // Debit
			amount, _ = util.ParseAmount(r[2])
		} else if r[3] != "" { // Credit
			a, _ := util.ParseAmount(r[3])
			amount = a
		}

		normalizedName := strings.ToLower(r[1])
		normalizedCategory := strings.ToLower(r[4])

		var cc sql.NullInt32
		categories, _ := model.SearchCategories(p.DB, []string{normalizedName, normalizedCategory})
		if len(categories) > 0 {
			cc = sql.NullInt32{Valid: true, Int32: int32(categories[0].ID)}
		}

		date, _ := time.Parse("Jan 02, 2006", r[0])

		transactions = append(transactions, model.Transaction{
			ID:         uuid.NewString(),
			Name:       r[1],
			Source:     "citi",
			Account:    "citi",
			Date:       date.Format("2006-01-02"),
			Amount:     amount,
			CategoryID: cc,
			Category:   r[4],
		})
	}

	return transactions, nil
}
