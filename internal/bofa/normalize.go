package bofa

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"fin-web/internal/model"

	"github.com/google/uuid"
)

type Provider struct {
	DB *sql.DB
}

func NewBofaProvider(db *sql.DB) *Provider {
	return &Provider{
		DB: db,
	}
}

func (p *Provider) GetPrefix() string {
	return "bofa"
}

func (p *Provider) ParseFile(filePath string) ([]model.Transaction, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}

	var transactions []model.Transaction
	for _, r := range records[1:] {
		// Category logic (optimized: you should eventually cache these)
		var cc sql.NullInt32
		normalizedName := strings.ToLower(r[2])
		categories, _ := model.SearchCategories(p.DB, []string{normalizedName}) //

		if len(categories) > 0 {
			cc = sql.NullInt32{Valid: true, Int32: int32(categories[0].ID)}
		}

		date, err := time.Parse("01/02/2006", r[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %q: %w", r[0], err)
		}
		amount, err := parseAmount(r[4])
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount %q: %w", r[4], err)
		}

		transactions = append(transactions, model.Transaction{
			ID:         uuid.NewString(),
			Name:       r[2],
			Source:     "bank_of_america",
			Account:    "bank_of_america",
			Date:       date.Format("2006-01-02"),
			Amount:     amount,
			CategoryID: cc,
		})
	}
	return transactions, nil
}

func parseAmount(amount string) (float64, error) {
	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return 0, err
	}

	return val * -1, nil
}
