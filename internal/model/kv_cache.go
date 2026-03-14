package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type KVItem struct {
	Key       string
	Value     string
	ExpiresAt time.Time
}

var ErrKVItemNotFound = errors.New("kv item not found")

func GetKVItem(conn *sql.DB, key string) (KVItem, error) {
	var item KVItem
	var expiresAtStr string

	row := conn.QueryRow(
		"SELECT key, value, expires_at FROM kv_cache WHERE key = ?",
		key,
	)

	err := row.Scan(&item.Key, &item.Value, &expiresAtStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return KVItem{}, ErrKVItemNotFound
		}
		return KVItem{}, fmt.Errorf("query item %s: %v", key, err)
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		fmt.Println(err.Error())
		return KVItem{}, errors.New("error parsing expires_at")
	}
	item.ExpiresAt = expiresAt

	if time.Now().After(item.ExpiresAt) {
		return KVItem{}, ErrKVItemNotFound
	}

	return item, nil
}

func PutKVItem(conn *sql.DB, key string, value string, ttl time.Duration) error {
	query := `
		INSERT INTO kv_cache (key, value, expires_at) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			expires_at = excluded.expires_at;
		`
	_, err := conn.Exec(query, key, value, time.Now().Add(ttl).Format(time.RFC3339))
	if err != nil {
		return err
	}

	return nil
}
