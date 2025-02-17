package postgres

import (
	"avito-merch-store/internal/storage"
	"context"
	"github.com/jackc/pgx/v5"
)

type MerchStoragePostgres struct {
	conn *pgx.Conn
}

func CreateMerchStoragePostgres(postgresConnect string) (*MerchStoragePostgres, error) {
	conn, err := pgx.Connect(context.Background(), postgresConnect)
	if err != nil {
		return nil, err
	}
	return &MerchStoragePostgres{conn}, nil
}

func (st *MerchStoragePostgres) GetByName(item string) (int, error) {
	query := `
        SELECT price from merch WHERE name=$1
    `
	ans, err := st.conn.Query(context.Background(), query, item)
	defer ans.Close()
	if err != nil {
		return -1, err
	}
	if !ans.Next() {
		return -1, storage.ErrMerchNotFound
	}
	var res int
	err = ans.Scan(&res)
	if err != nil {
		return -1, err
	}
	return res, nil
}
