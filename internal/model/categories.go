package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Category struct {
	ID       int
	Priority int
	Label    string
	Values   []CategoryValue
}

type CategoryValue struct {
	ID         sql.NullInt64
	CategoryID sql.NullInt64
	Value      sql.NullString
}

func GetCategories(conn *sql.DB) ([]Category, error) {
	rows, err := conn.Query(
		"SELECT c.id, c.label, c.priority FROM categories as c ORDER BY priority",
	)
	if err != nil {
		return []Category{}, err
	}
	defer rows.Close()

	categories := []Category{}

	for rows.Next() {
		category := Category{}
		if err := rows.Scan(
			&category.ID,
			&category.Label,
			&category.Priority,
		); err != nil {
			return []Category{}, err
		}

		categories = append(categories, category)

	}

	return categories, nil
}

func GetCategory(conn *sql.DB, ID string) (Category, error) {
	rows, err := conn.Query(
		"SELECT c.id, c.label, c.priority, cv.id as category_value_id, cv.value, cv.category_id FROM categories as c LEFT JOIN category_values as cv on c.id = cv.category_id WHERE c.id = ?",
		ID,
	)
	if err != nil {
		return Category{}, err
	}
	defer rows.Close()

	category := Category{}
	categoryValues := []CategoryValue{}

	for rows.Next() {
		categoryValue := CategoryValue{}
		if err := rows.Scan(
			&category.ID,
			&category.Label,
			&category.Priority,
			&categoryValue.ID,
			&categoryValue.Value,
			&categoryValue.CategoryID,
		); err != nil {
			return Category{}, err
		}

		if categoryValue.ID.Valid {
			categoryValues = append(categoryValues, categoryValue)
		}
	}

	category.Values = categoryValues

	return category, nil
}

func SearchCategories(conn *sql.DB, queries []string) ([]Category, error) {
	if len(queries) == 0 {
		return []Category{}, errors.New("at least one query is required")
	}

	queryStr := "SELECT DISTINCT c.id, c.priority, c.label FROM categories AS c JOIN category_values AS cv ON c.id = cv.category_id WHERE"
	args := []any{}
	filter := "? LIKE '%' || cv.value || '%'"
	filters := []string{}

	for _, v := range queries {
		args = append(args, v)
		filters = append(filters, filter)
	}

	queryStr = fmt.Sprintf("%s %s order by priority nulls last", queryStr, strings.Join(filters, " OR "))

	rows, err := conn.Query(
		queryStr,
		args...,
	)
	if err != nil {
		return []Category{}, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		category := Category{}
		if err := rows.Scan(
			&category.ID,
			&category.Priority,
			&category.Label,
		); err != nil {
			return []Category{}, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func CreateCategory(conn *sql.DB, name string, priority int) (int, error) {
	queryStr := "INSERT INTO categories (label, priority) VALUES(?, ?) RETURNING id"
	args := []any{
		name,
		priority,
	}

	var lastInsertID int
	err := conn.QueryRow(
		queryStr,
		args...,
	).Scan(&lastInsertID)
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

type UpdateCategoryParams struct {
	Label    *string
	Priority *int
}

func UpdateCategory(conn *sql.DB, ID string, params UpdateCategoryParams) error {
	queryStr := "UPDATE categories SET"
	updates := []string{}
	args := []any{}

	if params.Label != nil {
		updates = append(updates, " label = ?")
		args = append(args, *params.Label)
	}

	if params.Priority != nil {
		updates = append(updates, " priority = ?")
		args = append(args, *params.Priority)
	}

	if len(updates) == 0 {
		return nil
	}

	queryStr += strings.Join(updates, ",")
	queryStr += " WHERE id = ?"
	args = append(args, ID)

	_, err := conn.Exec(
		queryStr,
		args...,
	)
	if err != nil {
		return err
	}

	return nil
}

func DeleteCategory(conn *sql.DB, ID string) error {
	queryStr := "DELETE FROM categories WHERE id = ?"

	_, err := conn.Exec(
		queryStr,
		ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func CreateCategoryValue(conn *sql.DB, categoryID int, value string) (int, error) {
	queryStr := "INSERT INTO category_values (category_id, value) VALUES(?, ?) RETURNING id"
	args := []any{
		categoryID,
		value,
	}

	var lastInsertID int
	err := conn.QueryRow(
		queryStr,
		args...,
	).Scan(&lastInsertID)
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

type UpdateCategoryValueParams struct {
	Value *string
}

func UpdateCategoryValue(conn *sql.DB, ID string, params UpdateCategoryValueParams) error {
	queryStr := "UPDATE category_values SET"
	updates := []string{}
	args := []any{}

	if params.Value != nil {
		updates = append(updates, " value = ?")
		args = append(args, *params.Value)
	}

	if len(updates) == 0 {
		return nil
	}

	queryStr += strings.Join(updates, ",")
	queryStr += " WHERE id = ?"
	args = append(args, ID)

	_, err := conn.Exec(
		queryStr,
		args...,
	)
	if err != nil {
		return err
	}

	return nil
}

func DeleteCategoryValue(conn *sql.DB, ID int) error {
	queryStr := "DELETE FROM category_values WHERE id = ?"

	_, err := conn.Exec(
		queryStr,
		ID,
	)
	if err != nil {
		return err
	}

	return nil
}
