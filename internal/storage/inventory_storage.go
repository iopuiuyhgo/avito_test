package storage

import "avito-merch-store/model"

type InventoryStorage interface {
	AddItems(userID int, item string, quantity int) error
	GetByUserID(userID int, count int) ([]model.InventoryItem, error)
}
