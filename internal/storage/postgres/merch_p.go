package postgres

import (
	"avito-merch-store/internal/storage"
	"avito-merch-store/model"
	"context"
	"github.com/jackc/pgx/v5"
)

type MerchStoragePostgres struct {
	conn *pgx.Conn
}

func CreateMerchStoragePostgres(postgresConnect string, items []model.Item) (*MerchStoragePostgres, error) {
	conn, err := pgx.Connect(context.Background(), postgresConnect)
	if err != nil {
		return nil, err
	}
	query := `
        INSERT INTO merch (name, price)
        VALUES ($1, $2)
        ON CONFLICT (name) DO NOTHING
    `

	for _, item := range items {
		_, err := conn.Exec(context.Background(), query, item.Name, item.Price)

		if err != nil {
			err1 := conn.Close(context.Background())
			if err1 != nil {
				return nil, err1
			}
			return nil, err
		}
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
