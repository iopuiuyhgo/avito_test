package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

type AuthStoragePostgres struct {
	conn *pgx.Conn
}

func CreateAuthStoragePostgres(postgresConnect string) (*AuthStoragePostgres, error) {
	conn, err := pgx.Connect(context.Background(), postgresConnect)
	if err != nil {
		return nil, err
	}
	return &AuthStoragePostgres{conn}, nil
}

func (auth *AuthStoragePostgres) AddUser(username, password string) error {
	query := `
        INSERT INTO auth (username, password_hash)
        VALUES ($1, $2)
    `
	_, err := auth.conn.Exec(context.Background(), query, username, password)
	if err != nil {
		return err
	}
	return nil
}

func (auth *AuthStoragePostgres) CheckUser(username, password string) bool {
	query := `
        SELECT * FROM auth WHERE username = $1 AND password_hash = $2
    `
	ans, err := auth.conn.Query(context.Background(), query, username, password)
	defer ans.Close()
	if err != nil {
		return false
	}
	return ans.Next()
}

func (auth *AuthStoragePostgres) CheckContains(username string) bool {
	query := `
        SELECT * FROM auth WHERE username = $1
    `
	ans, err := auth.conn.Query(context.Background(), query, username)
	defer ans.Close()
	if err != nil {
		return false
	}
	return ans.Next()
}

// Exported method to satisfy the interface
func (auth *AuthStoragePostgres) GetUserHash(username string) (string, error) {
	query := `
        SELECT password_hash FROM auth WHERE username = $1
    `
	ans, err := auth.conn.Query(context.Background(), query, username)
	defer ans.Close()
	if err != nil {
		return "", err
	}
	if !ans.Next() {
		return "", errors.New("user does not exist")
	}
	var r string
	err = ans.Scan(&r)
	if err != nil {
		return "", err
	}
	return r, nil
}
