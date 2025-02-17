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

func (st *TransactionStoragePostgres) CreateTransaction(senderID int, receiverID, amount int) error {
	ctx := context.Background()

	var err error
	query := `
        INSERT INTO transactions (sender_id, receiver_id, amount, created_at)
        VALUES ($1, $2, $3, $4)
    `
	_, err = st.conn.Exec(ctx, query, senderID, receiverID, amount, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (st *TransactionStoragePostgres) GetTransactionHistory(userID int, count int) ([]model.Transaction, error) {
	ctx := context.Background()
	query := `
        SELECT id, sender_id, receiver_id, amount, created_at
        FROM transactions
        WHERE sender_id = $1 OR receiver_id = $1
        ORDER BY created_at DESC
    `
	rows, err := st.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.Transaction
	for i := 0; rows.Next() && (i < count || count == -1); i++ {
		var t model.Transaction
		err := rows.Scan(&t.ID, &t.SenderID, &t.ReceiverID, &t.Amount, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}
