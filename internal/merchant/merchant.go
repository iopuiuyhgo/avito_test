package merchant

import (
	"avito-merch-store/internal/storage"
	"avito-merch-store/model"
	"fmt"
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
	Received []TransactionFrom `json:"received"`
	Sent     []TransactionTo   `json:"sent"`
}

type Transaction struct {
	FromUser string
	ToUser   string
	Amount   int
}
type TransactionFrom struct {
	FromUser string `json:"fromUser,omitempty"`
	Amount   int    `json:"amount"`
}
type TransactionTo struct {
	ToUser string `json:"toUser,omitempty"`
	Amount int    `json:"amount"`
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
var ErrIncorrectCount = fmt.Errorf("you can't send less than one coin")

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

	trans, err := m.GetTransactions(username)
	if err != nil {
		return nil, err
	}

	var transRes CoinHistory
	for _, item := range trans {
		tr := Transaction{FromUser: item.SenderName, ToUser: item.ReceiverName, Amount: item.Amount}
		if item.SenderName == username {
			transRes.Sent = append(transRes.Sent, TransactionTo{tr.ToUser, tr.Amount})
		} else {
			transRes.Received = append(transRes.Received, TransactionFrom{tr.FromUser, tr.Amount})
		}
	}

	return &InfoResponse{Coins: user.Coins, Inventory: result, CoinHistory: transRes}, nil
}

func (m *Merchant) GetTransactions(username string) ([]model.Transaction, error) {
	history, err := m.transaction.GetTransactionHistory(username, -1)
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
	if count < 1 {
		return ErrIncorrectCount
	}
	user2, err := m.users.GetByUsername(receiver)
	if err != nil {
		return err
	}
	err = m.users.UpdateCoins(user2.ID, user2.Coins+count)
	if err != nil {
		return err
	}
	err = m.users.UpdateCoins(user.ID, user.Coins-count)
	if err != nil {
		return err
	}
	err = m.transaction.CreateTransaction(user.Username, user2.Username, count)
	if err != nil {
		return err
	}
	return nil
}
