package main

import (
	"avito-merch-store/internal/auth"
	"avito-merch-store/internal/merchant"
	"avito-merch-store/internal/storage/postgres"
	"avito-merch-store/internal/web"
	"log"
	"net/http"
	"os"
)

func StartServer(addr string) {

}

func main() {
	err := postgres.UpMigrations(os.Getenv("POSTGRES_PATH"))
	if err != nil {
		log.Fatal(err)
	}
	ptx := os.Getenv("POSTGRES_PATH")

	err = postgres.UpMigrations(ptx)
	if err != nil {
		log.Fatal(err)
	}
	stor, err := postgres.CreateAuthStoragePostgres(os.Getenv("POSTGRES_PATH"))
	if err != nil {
		log.Fatal(err)
	}

	a, err := postgres.CreateUserStoragePostgres(ptx)
	b, err := postgres.CreateInventoryStoragePostgres(ptx)
	c, err := postgres.CreateTransactionStoragePostgres(ptx)
	d, err := postgres.CreateMerchStoragePostgres(ptx)
	if err != nil {
		log.Fatal(err)
	}
	service := web.NewService(stor, auth.CreateAuthenticator(os.Getenv("JWT_KEY")), merchant.CreateMerchant(a, b, c, d))
	log.Printf("Starting test server on :8080")
	log.Fatal(http.ListenAndServe(":8080", service))
}
