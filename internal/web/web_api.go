package web

import (
	"avito-merch-store/internal/auth"
	"avito-merch-store/internal/merchant"
	"avito-merch-store/internal/storage"
	"avito-merch-store/model"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strings"
)

type Service struct {
	router  *mux.Router
	storage storage.AuthStorage
	auth    auth.Authenticator
	merch   merchant.Merchant
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"` // Логин получателя
	Amount int    `json:"amount"` // Количество монет
}

func NewService(storage storage.AuthStorage, auth auth.Authenticator, merchant merchant.Merchant) *Service {
	s := &Service{
		router:  mux.NewRouter(),
		storage: storage,
		auth:    auth,
		merch:   merchant,
	}

	s.configureRouter()
	return s
}

func (s *Service) configureRouter() {
	s.router.HandleFunc("/api/info", s.AuthMiddleware(s.GetInfoHandler)).Methods("GET")
	s.router.HandleFunc("/api/transactions", s.AuthMiddleware(s.GetTransactionsHandler)).Methods("GET")
	s.router.HandleFunc("/api/sendCoin", s.AuthMiddleware(s.SendCoinHandler)).Methods("POST")
	s.router.HandleFunc("/api/buy/{item}", s.AuthMiddleware(s.BuyItemHandler)).Methods("GET")
	s.router.HandleFunc("/api/auth", s.AuthHandler).Methods("POST")
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Service) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		key, err := s.auth.ValidateKey(tokenString)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
		}
		ctx := context.WithValue(r.Context(), "name", key)
		next(w, r.WithContext(ctx))
	}
}

func (s *Service) GetInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	info, err := s.merch.GetInfoByUsername(ctx.Value("name").(string))
	if errors.Is(err, storage.ErrUserNotFound) {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, info)
}

func (s *Service) GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := ctx.Value("name").(string)
	transactions, err := s.merch.GetTransactions(username)
	if errors.Is(err, storage.ErrUserNotFound) {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	respondWithJSON(w, http.StatusOK, transactions)
}

func (s *Service) SendCoinHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body, err := io.ReadAll(r.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
	}(r.Body)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error: cannot read request body")
		return
	}

	var requestData SendCoinRequest
	if err := json.Unmarshal(body, &requestData); err != nil {
		respondWithError(w, http.StatusBadRequest, "error: invalid JSON format")
		return
	}
	err = s.merch.SendCoin(ctx.Value("name").(string), requestData.ToUser, requestData.Amount)
	if errors.Is(err, storage.ErrUserNotFound) {
		respondWithError(w, http.StatusNotFound, err.Error())
	}
	if err != nil {
		respondWithJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, nil)
}

func (s *Service) BuyItemHandler(w http.ResponseWriter, r *http.Request) {
	item := mux.Vars(r)["item"]
	ctx := r.Context()

	err := s.merch.Buy(ctx.Value("name").(string), item)
	if errors.Is(err, merchant.ErrNotEnoughCoins) {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJSON(w, http.StatusOK, nil)
}

func (s *Service) AuthHandler(w http.ResponseWriter, r *http.Request) {
	var req model.AuthRequestWeb
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	if !s.storage.CheckContains(req.Username) {
		hash, err := s.auth.HashPassword(req.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		err = s.storage.AddUser(req.Username, hash)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		hash, err := s.storage.GetUserHash(req.Password)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if !s.auth.CheckPassword(hash, req.Password) {
			respondWithError(w, http.StatusUnauthorized, "password is not correct")
			return
		}
	}

	key, err := s.auth.GenerateKey(req.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = s.merch.AddUser(req.Username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, model.AuthResponseWeb{
		Token: key,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		err := json.NewEncoder(w).Encode(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, model.ErrorResponseWeb{Errors: message})
}
