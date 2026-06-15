package worker

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"strings"

	"fin-web/internal/model"
)

type Provider interface {
	GetPrefix() string
	ParseFile(filePath string) ([]model.Transaction, error)
}

type BaseWorker struct {
	DB      *sql.DB
	DirPath string
}

func NewBaseWorker(db *sql.DB, dp string) *BaseWorker {
	return &BaseWorker{
		DB:      db,
		DirPath: dp,
	}
}

func (bw *BaseWorker) Process(p Provider) error {
	entries, err := os.ReadDir(bw.DirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), p.GetPrefix()) {
			continue
		}

		filePath := path.Join(bw.DirPath, entry.Name())

		transactions, err := p.ParseFile(filePath)
		if err != nil {
			fmt.Printf("failed to parse %s: %v\n", entry.Name(), err)
			continue
		}

		var insertedIDs []string
		var failed bool

		for _, t := range transactions {
			if err := model.CreateTransaction(bw.DB, t); err != nil {
				fmt.Printf("failed to create transaction %s: %v\n", t.Name, err)
				failed = true
				break
			}
			insertedIDs = append(insertedIDs, t.ID)
		}

		if failed {
			for _, id := range insertedIDs {
				if err := model.DeleteTransaction(bw.DB, id); err != nil {
					fmt.Printf("failed to rollback transaction ID %s: %v\n", id, err)
				}
			}
			continue
		}

		err = os.Remove(filePath)
		if err != nil {
			fmt.Println("Error deleting file:", err)
		}

	}
	return nil
}
