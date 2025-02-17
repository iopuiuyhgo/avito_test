package postgres

import (
	"avito-merch-store/model"
	"context"
	"github.com/jackc/pgx/v5"
)

type InventoryStoragePostgres struct {
	conn *pgx.Conn
}

func CreateInventoryStoragePostgres(postgresConnect string) (*InventoryStoragePostgres, error) {
	conn, err := pgx.Connect(context.Background(), postgresConnect)
	if err != nil {
		return nil, err
	}
	return &InventoryStoragePostgres{conn}, nil
}

func (st *InventoryStoragePostgres) AddItems(userID int, item string, quantity int) error {
	ctx := context.Background()

	query := `
        INSERT INTO inventory (user_id, item_name, quantity)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, item_name)
        DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity
    `

	_, err := st.conn.Exec(ctx, query, userID, item, quantity)
	if err != nil {
		return err
	}

	return nil
}

func (st *InventoryStoragePostgres) GetByUserID(userID int, count int) ([]model.InventoryItem, error) {
	query := `
        SELECT * FROM inventory WHERE user_id = $1
    `
	response, err := st.conn.Query(context.Background(), query, userID)
	defer response.Close()
	if err != nil {
		return nil, err
	}
	var res []model.InventoryItem
	for i := 0; response.Next() && i < count; i++ {
		var item model.InventoryItem
		err := response.Scan(&item.UserID, &item.ItemName, &item.Quantity)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}

	return res, nil
}
