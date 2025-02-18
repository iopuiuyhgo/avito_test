package main

import (
	"avito-merch-store/internal/auth"
	"avito-merch-store/internal/storage/postgres"
	"avito-merch-store/model"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

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
	Received []ReceivedItem `json:"received"`
	Sent     []SentItem     `json:"sent"`
}

type ReceivedItem struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type SentItem struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}

func getAuthToken(t *testing.T, serverURL string, usr string, pass string) string {
	authReq := AuthRequest{
		Username: usr,
		Password: pass,
	}
	body, err := json.Marshal(authReq)
	if err != nil {
		t.Fatalf("Ошибка маршалинга запроса аутентификации: %v", err)
	}

	res, err := http.Post(serverURL+"/api/auth", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Ошибка отправки запроса аутентификации: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			t.Errorf(err.Error())
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200 от /api/auth, получен %d", res.StatusCode)
	}

	var authResp AuthResponse
	err = json.NewDecoder(res.Body).Decode(&authResp)
	if err != nil {
		t.Fatalf("Ошибка декодирования ответа аутентификации: %v", err)
	}
	if authResp.Token == "" {
		t.Fatalf("Получен пустой токен")
	}
	return authResp.Token
}

func TestAPI(t *testing.T) {
	i := 0

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
	err := postgres.DownMigrations(ptx, "/internal/storage/postgres/migrations")
	err = postgres.UpMigrations(ptx, "/internal/storage/postgres/migrations")

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
	go runServer(ptx, "8080", au, stor, a, b, c, d, &wg1)
	wg1.Wait()

	URL := "http://127.0.0.1:8080"

	t.Run("Auth_Success", func(t *testing.T) {
		token := getAuthToken(t, URL, "auth_testuser", "password")
		if token == "" {
			t.Error("Токен не должен быть пустым")
		}
		log.Printf("log%d\n", i)
		i++
	})

	t.Run("Auth_InvalidRequest", func(t *testing.T) {
		res, err := http.Post(URL+"/api/auth", "application/json", bytes.NewReader([]byte(`{"invalid": "data"}`)))
		if err != nil {
			t.Fatalf("Ошибка отправки запроса /api/auth с неверными данными: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusBadRequest && res.StatusCode != http.StatusUnauthorized {
			t.Errorf("Ожидался статус 400 или 401 для неверного запроса, получен %d", res.StatusCode)
		}
	})

	user := getAuthToken(t, URL, "testuser", "password")
	authHeader := "Bearer " + user
	_ = getAuthToken(t, URL, "anotherUser", "password")

	t.Run("Info_NoAuth", func(t *testing.T) {
		req, err := http.NewRequest("GET", URL+"/api/info", nil)
		if err != nil {
			t.Fatalf("Ошибка создания запроса /api/info: %v", err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Ошибка выполнения запроса /api/info: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("Ожидался статус 401 для отсутствия токена, получен %d", res.StatusCode)
		}
	})

	t.Run("Info_Success", func(t *testing.T) {
		req, err := http.NewRequest("GET", URL+"/api/info", nil)
		if err != nil {
			t.Fatalf("Ошибка создания запроса /api/info: %v", err)
		}
		req.Header.Set("Authorization", authHeader)
		client := &http.Client{Timeout: 5 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Ошибка выполнения запроса /api/info: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Ожидался статус 200, получен %d", res.StatusCode)
		}

		var infoResp InfoResponse
		err = json.NewDecoder(res.Body).Decode(&infoResp)
		if err != nil {
			t.Fatalf("Ошибка декодирования ответа /api/info: %v", err)
		}
		if infoResp.Coins < 0 {
			t.Errorf("Неверное количество монет: %d", infoResp.Coins)
		}
	})

	t.Run("SendCoin_NoAuth", func(t *testing.T) {
		reqBody := SendCoinRequest{
			ToUser: "anotherUser",
			Amount: 10,
		}
		body, _ := json.Marshal(reqBody)
		req, err := http.NewRequest("POST", URL+"/api/sendCoin", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Ошибка создания запроса /api/sendCoin: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Ошибка выполнения запроса /api/sendCoin: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("Ожидался статус 401 для отсутствия токена, получен %d", res.StatusCode)
		}
	})

	t.Run("SendCoin_InvalidBody", func(t *testing.T) {
		req, err := http.NewRequest("POST", URL+"/api/sendCoin", bytes.NewReader([]byte(`{"toUser": 123}`)))
		if err != nil {
			t.Fatalf("Ошибка создания запроса /api/sendCoin: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Ошибка выполнения запроса /api/sendCoin: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Ожидался статус 400 для неверного тела запроса, получен %d", res.StatusCode)
		}
	})

	t.Run("SendCoin_Success", func(t *testing.T) {
		reqBody := SendCoinRequest{
			ToUser: "anotherUser",
			Amount: 10,
		}
		body, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Ошибка маршалинга запроса /api/sendCoin: %v", err)
		}
		req, err := http.NewRequest("POST", URL+"/api/sendCoin", bytes.NewReader(body))
		log.Println(req.URL)
		if err != nil {
			t.Fatalf("Ошибка создания запроса /api/sendCoin: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		client := &http.Client{Timeout: 5 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Ошибка выполнения запроса /api/sendCoin: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusOK {
			t.Errorf("Ожидался статус 200 при успешной отправке монет, получен %d", res.StatusCode)
		}
	})

	t.Run("Buy_NoAuth", func(t *testing.T) {
		req, err := http.NewRequest("GET", URL+"/api/buy/sword", nil)
		if err != nil {
			t.Fatalf("Ошибка создания запроса /api/buy: %v", err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Ошибка выполнения запроса /api/buy: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("Ожидался статус 401 для отсутствия токена, получен %d", res.StatusCode)
		}
	})

	t.Run("Buy_Success", func(t *testing.T) {
		req, err := http.NewRequest("GET", URL+"/api/buy/socks", nil)
		if err != nil {
			t.Fatalf("Ошибка создания запроса /api/buy: %v", err)
		}
		req.Header.Set("Authorization", authHeader)
		client := &http.Client{Timeout: 5 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Ошибка выполнения запроса /api/buy: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusOK {
			t.Errorf("Ожидался статус 200 при успешной покупке, получен %d", res.StatusCode)
		}
	})

	t.Run("Buy_InvalidItem", func(t *testing.T) {
		req, err := http.NewRequest("GET", URL+"/api/buy/incorrect", nil)
		if err != nil {
			t.Fatalf("Ошибка создания запроса /api/buy с некорректным item: %v", err)
		}
		req.Header.Set("Authorization", authHeader)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Ошибка выполнения запроса /api/buy с некорректным item: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Println(err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("Ожидался статус 400 для некорректного item, получен %d", res.StatusCode)
		}
	})
}
