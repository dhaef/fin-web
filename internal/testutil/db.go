// Package testutil provides shared helpers for tests across the project.
package testutil

import (
	"database/sql"
	_ "embed"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

//go:embed schema.sql
var schema string

// NewDB returns an isolated in-memory SQLite database with the full
// application schema applied. The connection is closed when the test finishes.
func NewDB(t *testing.T) *sql.DB {
	t.Helper()

	conn, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	_, err = conn.Exec(schema)
	require.NoError(t, err)

	return conn
}

// SeedCategory inserts a category with a single matching value, used to verify
// transaction categorization during parsing.
func SeedCategory(t *testing.T, db *sql.DB, label string, priority int, value string) int {
	t.Helper()

	res, err := db.Exec("INSERT INTO categories(label, priority) VALUES(?, ?)", label, priority)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO category_values(category_id, value) VALUES(?, ?)", id, value)
	require.NoError(t, err)

	return int(id)
}
