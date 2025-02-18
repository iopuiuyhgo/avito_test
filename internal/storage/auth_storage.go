package storage

type AuthStorage interface {
	AddUser(username string, hashPassword string) error
	CheckUser(username string, hashPassword string) bool
	CheckContains(username string) bool
	GetUserHash(username string) (string, error)
}
