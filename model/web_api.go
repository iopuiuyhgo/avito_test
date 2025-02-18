package model

type AuthRequestWeb struct {
	Username string
	Password string
}

type AuthResponseWeb struct {
	Token string `json:"token"`
}

type ErrorResponseWeb struct {
	Errors string `json:"errors"`
}

type InfoResponseWeb struct {
	Coins       int
	Inventory   []InventoryWeb
	CoinHistory CoinHistoryWeb
}

type InventoryWeb struct {
	Type     string
	Quantity int
}

type CoinHistoryWeb struct {
	Received []Transaction
	Sent     []Transaction
}

type TransactionWeb struct {
	User   string
	Amount int
}

type SendCoinRequestWeb struct {
	ToUser string
	Amount int
}
