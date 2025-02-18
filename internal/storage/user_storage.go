package storage

import (
	"avito-merch-store/model"
	"fmt"
)

var ErrUserNotFound = fmt.Errorf("user not found")

type UserStorage interface {
	Create(username string, coins int) error
	GetByUsername(username string) (*model.User, error)
	UpdateCoins(userID int, newCoins int) error
}
