package tui

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"fin-web/internal/model"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type TransactionsTable struct {
	table table.Model
}

func NewTransactionsTable(dbConn *sql.DB) TransactionsTable {
	transactions, err := model.QueryTransactions(dbConn, model.QueryTransactionsFilters{
		OrderBy:        "date",
		OrderDirection: "DESC",
		Limit:          1,
	})
	if err != nil {
		log.Fatal("error fetching default dates: " + err.Error())
	}

	date, err := time.Parse("2006-01-02", transactions[0].Date)
	if err != nil {
		fmt.Println(err.Error())
	}

	firstDayOfThisMonth, endOfThisMonth := getStartAndEndOfMonth(date)

	startDate := firstDayOfThisMonth.Format("2006-01-02")
	endDate := endOfThisMonth.Format("2006-01-02")

	transactions, err = model.QueryTransactions(dbConn, model.QueryTransactionsFilters{
		OrderBy:        "amount",
		OrderDirection: "desc",
		StartDate:      startDate,
		EndDate:        endDate,
		// Categories:          categories,
		CategoriesToExclude: []string{"34"},
	})
	if err != nil {
		log.Fatal("error fetching transactions: " + err.Error())
	}

	columns := []table.Column{
		{Title: "Name", Width: 50},
		{Title: "Amount", Width: 10},
		{Title: "Date", Width: 10},
		{Title: "Category", Width: 10},
		{Title: "Account", Width: 10},
	}

	rows := make([]table.Row, len(transactions))

	for idx, t := range transactions {
		rows[idx] = table.Row{
			t.Name,
			strconv.FormatFloat(t.Amount, 'f', 2, 64),
			t.Date,
			t.CustomCategory.String,
			t.Account,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(7),
		table.WithWidth(100),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return TransactionsTable{table: t}
}

func (t TransactionsTable) Init() tea.Cmd {
	return nil
}

func (t TransactionsTable) View() tea.View {
	return tea.NewView(baseStyle.Render(t.table.View()) + "\n  " + t.table.HelpView() + "\n")
}

func (t TransactionsTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if t.table.Focused() {
				t.table.Blur()
			}
		case "t":
			t.table.Focus()
		case "q", "ctrl+c":
			return t, tea.Quit
		case "enter":

			return t, tea.Batch(
				tea.Printf("Let's go to %s!", t.table.SelectedRow()[1]),
			)
		}
	}

	t.table, cmd = t.table.Update(msg)
	return t, cmd
}

func getStartAndEndOfMonth(date time.Time) (time.Time, time.Time) {
	year, month, _ := date.Date()
	firstDayOfThisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	endOfThisMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)

	return firstDayOfThisMonth, endOfThisMonth
}
