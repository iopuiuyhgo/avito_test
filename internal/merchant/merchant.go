package merchant

import (
	"avito-merch-store/internal/storage"
	"avito-merch-store/model"
	"fmt"
	"log"
)

type Merchant struct {
	users       storage.UserStorage
	inventory   storage.InventoryStorage
	transaction storage.TransactionStorage
	merch       storage.MerchStorage
}

func CreateMerchant(users storage.UserStorage, inventory storage.InventoryStorage,
	transaction storage.TransactionStorage, merch storage.MerchStorage) Merchant {
	return Merchant{users, inventory, transaction, merch}
}

type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Item      `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}

type Item struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []Transaction `json:"received"`
	Sent     []Transaction `json:"sent"`
}

type Transaction struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
	Amount   int    `json:"amount"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

var ErrNotEnoughCoins = fmt.Errorf("not enough coins")

func (m *Merchant) AddUser(username string) error {
	return m.users.Create(username, 1000)

}

func (m *Merchant) GetInfoByUsername(username string) (*InfoResponse, error) {
	user, err := m.users.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	inventory, err := m.inventory.GetByUserID(user.ID, -1)
	if err != nil {
		return nil, err
	}

	var result []Item
	for _, item := range inventory {
		result = append(result, Item{
			Type:     item.ItemName,
			Quantity: item.Quantity,
		})
	}
	return &InfoResponse{Coins: user.Coins, Inventory: result}, nil
}

func (m *Merchant) GetTransactions(username string) ([]model.Transaction, error) {
	user, err := m.users.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	history, err := m.transaction.GetTransactionHistory(user.ID, -1)
	if err != nil {
		return nil, err
	}
	return history, nil

}
func (m *Merchant) Buy(username string, item string) error {
	user, err := m.users.GetByUsername(username)
	if err != nil {
		return err
	}
	price, err := m.merch.GetByName(item)
	if err != nil {
		return err
	}
	if user.Coins < price {
		return ErrNotEnoughCoins
	}
	err = m.users.UpdateCoins(user.ID, user.Coins-price)
	if err != nil {
		return err
	}
	err = m.inventory.AddItems(user.ID, item, 1)
	if err != nil {
		return err
	}
	return nil
}

func (m *Merchant) SendCoin(username string, receiver string, count int) error {
	user, err := m.users.GetByUsername(username)
	if err != nil {
		return err
	}
	if user.Coins < count {
		return ErrNotEnoughCoins
	}
	user2, err := m.users.GetByUsername(receiver)
	if err != nil {
		return err
	}
	log.Println(user2)
	err = m.users.UpdateCoins(user2.ID, user2.Coins+count)
	if err != nil {
		return err
	}
	err = m.users.UpdateCoins(user.ID, user.Coins-count)
	if err != nil {
		return err
	}
	err = m.transaction.CreateTransaction(user.ID, user2.ID, count)
	if err != nil {
		return err
	}
	return nil
}
