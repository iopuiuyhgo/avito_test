package postgres

import (
	"avito-merch-store/model"
	"context"
	"github.com/jackc/pgx/v5"
	"time"
)

type TransactionStoragePostgres struct {
	conn *pgx.Conn
}

func CreateTransactionStoragePostgres(postgresConnect string) (*TransactionStoragePostgres, error) {
	conn, err := pgx.Connect(context.Background(), postgresConnect)
	if err != nil {
		return nil, err
	}
	return &TransactionStoragePostgres{conn}, nil
}

func (st *TransactionStoragePostgres) CreateTransaction(senderName string, receiverName string, amount int) error {
	ctx := context.Background()

	var err error
	query := `
        INSERT INTO transactions (sender_username, receiver_username, amount, created_at)
        VALUES ($1, $2, $3, $4)
    `
	_, err = st.conn.Exec(ctx, query, senderName, receiverName, amount, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (st *TransactionStoragePostgres) GetTransactionHistory(username string, count int) ([]model.Transaction, error) {
	ctx := context.Background()
	query := `
        SELECT id, sender_username, receiver_username, amount, created_at
        FROM transactions
        WHERE sender_username = $1 OR receiver_username = $1
        ORDER BY created_at DESC
    `
	rows, err := st.conn.Query(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for i := 0; rows.Next() && (i < count || count == -1); i++ {
		var t model.Transaction
		err := rows.Scan(&t.ID, &t.SenderName, &t.ReceiverName, &t.Amount, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}
