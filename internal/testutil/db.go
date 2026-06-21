// Package testutil provides shared helpers for tests across the project.
package testutil

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// categorySchema mirrors the production categories/category_values tables and
// the priority-uniqueness triggers. It is kept here so tests stay hermetic and
// don't depend on a checked-in database file.
const categorySchema = `
CREATE TABLE categories(
	id integer primary key autoincrement,
	priority INTEGER not null,
	label text,
	is_ignored BOOLEAN DEFAULT 0,
	type TEXT CHECK(type IS NULL OR type IN ('income', 'fixed', 'fun', 'neutral'))
);

CREATE TABLE category_values(
	id integer primary key autoincrement,
	category_id integer not null,
	value text not null
);

CREATE TRIGGER validate_insert_categories
BEFORE INSERT ON categories
FOR EACH ROW
WHEN EXISTS (SELECT 1 FROM categories WHERE priority = NEW.priority)
BEGIN
	SELECT RAISE(ABORT, 'Error: This value already exists in the table.');
END;

CREATE TRIGGER validate_update_category_priority
BEFORE UPDATE OF priority ON categories
FOR EACH ROW
WHEN EXISTS (SELECT 1 FROM categories
			 WHERE priority = NEW.priority
			 AND id != OLD.id)
BEGIN
	SELECT RAISE(ABORT, 'Error: This value already exists in another row.');
END;
`

// NewCategoryDB returns an isolated in-memory SQLite database with the
// categories schema applied. The connection is closed when the test finishes.
func NewCategoryDB(t *testing.T) *sql.DB {
	t.Helper()

	conn, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	_, err = conn.Exec(categorySchema)
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
