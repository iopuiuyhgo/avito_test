package main

import (
	"avito-merch-store/internal/auth"
	"avito-merch-store/internal/merchant"
	"avito-merch-store/internal/storage"
	"avito-merch-store/internal/storage/postgres"
	"avito-merch-store/internal/web"
	"avito-merch-store/model"
	"log"
	"net/http"
	"os"
	"sync"
)

func StartServer(addr string) {

}

func runServer(ptx string,
	port string,
	au auth.Authenticator,
	stor storage.AuthStorage,
	users storage.UserStorage,
	inventory storage.InventoryStorage,
	transaction storage.TransactionStorage,
	merch storage.MerchStorage,
	wg1 *sync.WaitGroup) {
	// /internal/storage/postgres/migrations

	service := web.NewService(stor, au,
		merchant.CreateMerchant(users, inventory, transaction, merch))
	wg1.Done()
	log.Fatal(http.ListenAndServe(":"+port, service))
}

func main() {
	ptx := os.Getenv("POSTGRES_PATH")

	items := []model.Item{
		{"t-shirt", 80},
		{"cup", 20},
		{"book", 50},
		{"pen", 10},
		{"powerbank", 200},
		{"hoody", 300},
		{"umbrella", 200},
		{"socks", 10},
		{"wallet", 50},
		{"pink-hoody", 500},
	}

	err := postgres.DownMigrations(ptx, "/migrations")
	err = postgres.UpMigrations(ptx, "/migrations")

	stor, err := postgres.CreateAuthStoragePostgres(ptx)
	if err != nil {
		log.Fatal(err)
	}
	au := auth.CreateAuthenticator(os.Getenv("JWT_KEY"))

	a, err := postgres.CreateUserStoragePostgres(ptx)
	b, err := postgres.CreateInventoryStoragePostgres(ptx)
	c, err := postgres.CreateTransactionStoragePostgres(ptx)
	d, err := postgres.CreateMerchStoragePostgres(ptx, items)
	if err != nil {
		log.Fatal(err)
	}
	var wg1 sync.WaitGroup
	wg1.Add(1)
	log.Printf("Starting test server on :8080")
	runServer(ptx, "8080", au, stor, a, b, c, d, &wg1)
}
