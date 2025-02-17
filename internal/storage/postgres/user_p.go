package postgres

import (
	"avito-merch-store/internal/storage"
	"avito-merch-store/model"
	"context"
	"errors"
	_ "github.com/golang-migrate/migrate/v4/database/pgx" // Драйвер для database/sql
	_ "github.com/golang-migrate/migrate/v4/source/file"  // Источник миграций из файлов
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // Адаптер pgx для database/sql
	"log"
)

type UserStoragePostgres struct {
	conn *pgx.Conn
}

func CreateUserStoragePostgres(postgresConnect string) (*UserStoragePostgres, error) {
	conn, err := pgx.Connect(context.Background(), postgresConnect)
	if err != nil {
		return nil, err
	}
	return &UserStoragePostgres{conn}, nil
}

func (st *UserStoragePostgres) Create(username string, coins int) error {
	query := `
        INSERT INTO users (username, coins)
        VALUES ($1, $2)
    `

	_, err := st.conn.Exec(context.Background(), query, username, coins)
	if err != nil {
		return err
	}

	return nil
}

func (st *UserStoragePostgres) GetByUsername(username string) (*model.User, error) {
	query := `
        SELECT id, username, coins 
        FROM users 
        WHERE username = $1
    `

	res := model.User{}
	err := st.conn.QueryRow(context.Background(), query, username).Scan(&res.ID, &res.Username, &res.Coins)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storage.ErrUserNotFound
	}

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (st *UserStoragePostgres) UpdateCoins(userID int, newCoins int) error {
	query := `
        UPDATE users 
        SET coins = $1 
        WHERE id = $2
    `

	result, err := st.conn.Exec(context.Background(), query, newCoins, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	log.Println(rowsAffected)
	return nil
}
