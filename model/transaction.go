package model

import "time"

type Transaction struct {
	ID           int
	SenderName   string
	ReceiverName string
	Amount       int
	CreatedAt    time.Time
}
