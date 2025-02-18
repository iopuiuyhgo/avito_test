package storage

import "avito-merch-store/model"

type TransactionStorage interface {
	CreateTransaction(senderUsername string, receiverUsername string, amount int) error
	GetTransactionHistory(username string, count int) ([]model.Transaction, error)
}
