package schwab

import (
	"database/sql"
	"encoding/json"
	"os"
	"strings"

	"fin-web/internal/model"
	"fin-web/internal/util"

	"github.com/google/uuid"
)

type Provider struct {
	DB *sql.DB
}

func NewSchwabProvider(db *sql.DB) *Provider {
	return &Provider{
		DB: db,
	}
}

func (p *Provider) GetPrefix() string {
	return "Checking"
}

type statementSchema struct {
	PostedTransactions []struct {
		Description string     `json:"Description"`
		Date        CustomDate `json:"Date"`
		Withdrawal  string     `json:"Withdrawal"`
		Deposit     string     `json:"Deposit"`
	} `json:"PostedTransactions"`
}

func (p *Provider) ParseFile(filePath string) ([]model.Transaction, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var statement statementSchema
	if err := json.Unmarshal(data, &statement); err != nil {
		return nil, err
	}

	var transactions []model.Transaction
	for _, t := range statement.PostedTransactions {
		var amount float64
		if t.Withdrawal != "" {
			amount, _ = util.ParseAmount(t.Withdrawal)
		} else if t.Deposit != "" {
			a, _ := util.ParseAmount(t.Deposit)
			amount = -a
		}

		normalizedName := strings.ToLower(t.Description)
		var cc sql.NullInt32
		categories, _ := model.SearchCategories(p.DB, []string{normalizedName})
		if len(categories) > 0 {
			cc = sql.NullInt32{Valid: true, Int32: int32(categories[0].ID)}
		}

		transactions = append(transactions, model.Transaction{
			ID:         uuid.NewString(),
			Name:       t.Description,
			Source:     "schwab",
			Account:    "schwab",
			Date:       t.Date.Format("2006-01-02"),
			Amount:     amount,
			CategoryID: cc,
		})
	}

	return transactions, nil
}
