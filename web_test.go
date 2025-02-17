package main

import (
	"avito-merch-store/internal/auth"
	"avito-merch-store/internal/merchant"
	"avito-merch-store/internal/storage/postgres"
	"avito-merch-store/internal/web"
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAuthAPI(t *testing.T) {
	ptx := os.Getenv("POSTGRES_PATH")
	stor, err := postgres.CreateAuthStoragePostgres(ptx)
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
	router := web.NewService(stor, auth.CreateAuthenticator(os.Getenv("JWT_KEY")), merchant.CreateMerchant(a, b, c, d))
	err = postgres.DownMigrations(ptx)
	if err != nil {
		log.Fatal(err)
	}
	err = postgres.UpMigrations(ptx)
	if err != nil {
		log.Fatal(err)
	}

	t.Run("Successful authentication", func(t *testing.T) {
		payload := map[string]string{
			"username": "testuser",
			"password": "validpass",
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "token")
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		payload := map[string]string{
			"username": "testuser",
			"password": "wrongpass",
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestInfoAPI(t *testing.T) {
	ptx := os.Getenv("POSTGRES_PATH")

	stor, err := postgres.CreateAuthStoragePostgres(ptx)
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
	router := web.NewService(stor, auth.CreateAuthenticator(os.Getenv("JWT_KEY")), merchant.CreateMerchant(a, b, c, d))

	token := getAuthToken(t, router) // Вспомогательная функция для получения токена

	t.Run("Get info with valid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/info", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response InfoResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.IsType(t, 0, response.Coins)
		assert.IsType(t, []InventoryItem{}, response.Inventory)
	})

	t.Run("Get info without token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/info", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestSendCoinAPI(t *testing.T) {
	ptx := os.Getenv("POSTGRES_PATH")

	stor, err := postgres.CreateAuthStoragePostgres(ptx)
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
	router := web.NewService(stor, auth.CreateAuthenticator(os.Getenv("JWT_KEY")), merchant.CreateMerchant(a, b, c, d))

	token := getAuthToken(t, router)

	t.Run("Successful coin transfer", func(t *testing.T) {
		payload := SendCoinRequest{
			ToUser: "recipient",
			Amount: 50,
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Transfer to invalid user", func(t *testing.T) {
		payload := SendCoinRequest{
			ToUser: "nonexistent",
			Amount: 50,
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestBuyItemAPI(t *testing.T) {
	ptx := os.Getenv("POSTGRES_PATH")

	stor, err := postgres.CreateAuthStoragePostgres(ptx)
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
	router := web.NewService(stor, auth.CreateAuthenticator(os.Getenv("JWT_KEY")), merchant.CreateMerchant(a, b, c, d))

	token := getAuthToken(t, router)

	t.Run("Successful item purchase", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/buy/item1", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Purchase nonexistent item", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/buy/invalid_item", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Вспомогательные структуры и функции
type InfoResponse struct {
	Coins       int             `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory CoinHistory     `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []CoinTransaction `json:"received"`
	Sent     []CoinTransaction `json:"sent"`
}

type CoinTransaction struct {
	FromUser string `json:"fromUser"`
	ToUser   string `json:"toUser"`
	Amount   int    `json:"amount"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

func getAuthToken(t *testing.T, router *web.Service) string {
	payload := map[string]string{
		"username": "testuser",
		"password": "validpass",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatal("Failed to get auth token")
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Error("json cannot be unmarshal")
	}
	return response["token"]
}
