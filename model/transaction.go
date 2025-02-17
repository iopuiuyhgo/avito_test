package model

import "time"

type Transaction struct {
	ID         int
	SenderID   int
	ReceiverID int
	Amount     int
	CreatedAt  time.Time
}
