package auth

type Authenticator interface {
	GenerateKey(username string) (string, error)
	ValidateKey(tokenString string) (string, error)
	HashPassword(username string) (string, error)
	CheckPassword(hashedPassword string, password string) bool
}
