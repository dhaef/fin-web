package main

import (
	"log"
	"os"

	"fin-web/internal/db"
	"fin-web/internal/tui"

	tea "charm.land/bubbletea/v2"
)

func main() {
	dbPath := os.Getenv("dbPath")

	if dbPath == "" {
		log.Fatal("dbPath is required")
	}

	dbConn, err := db.NewDbConnection(dbPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	app := tea.NewProgram(
		tui.NewApp(
			tui.NewRoute("hello", tui.NewText("Hello, world!")),
			tui.NewRoute("transactions", tui.NewTransactionsTable(dbConn)),
		),
	)
	if _, err := app.Run(); err != nil {
		panic(err)
	}
}
