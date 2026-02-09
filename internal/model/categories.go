package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Category struct {
	ID       int
	Category string
	Priority int
}

func SearchCategories(conn *sql.DB, queries []string) ([]Category, error) {
	if len(queries) == 0 {
		return []Category{}, errors.New("at least one query is required")
	}

	queryStr := "SELECT DISTINCT c.id, c.category, c.priority FROM categories AS c JOIN category_values AS cv ON c.id = cv.category_id WHERE"
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
			&category.Category,
			&category.Priority,
		); err != nil {
			return []Category{}, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func CreateCategory(conn *sql.DB, name string, priority int) (int, error) {
	queryStr := "INSERT INTO categories (category, priority) VALUES(?, ?) RETURNING id"
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
