package storage

import "fmt"

var ErrMerchNotFound = fmt.Errorf("error: cannot found item")

type MerchStorage interface {
	GetByName(item string) (int, error)
}
