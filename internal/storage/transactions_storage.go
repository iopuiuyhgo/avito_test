package storage

import "avito-merch-store/model"

type TransactionStorage interface {
	CreateTransaction(senderID, receiverID, amount int) error
	GetTransactionHistory(userID int, count int) ([]model.Transaction, error)
}
