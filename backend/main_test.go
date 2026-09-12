package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"btech-wallet/auth"
	"btech-wallet/config"
	"btech-wallet/transaction"
	"btech-wallet/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Integration tests require a running Postgres instance (see docker-compose.yaml).
// They are skipped automatically when the database is unreachable.

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dbURL := config.LoadEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/btech_wallet_db?sslmode=disable")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := config.InitDB(ctx, dbURL)
	if err != nil || pool.Ping(ctx) != nil {
		fmt.Println("skipping integration tests: database unavailable:", err)
		os.Exit(0)
	}

	testPool = pool
	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	jwtManager := auth.NewJWTManager(config.LoadEnv("JWT_SECRET", "your_jwt_secret_key"), time.Hour)

	auth.NewAuthHandler(testPool, jwtManager).RegisterRoutes(mux)
	transaction.NewTransactionHandler(testPool).RegisterRoutes(mux)
	user.NewUserHandler(testPool).RegisterRoutes(mux)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func cleanupUserByEmail(t *testing.T, email string) {
	t.Helper()
	ctx := context.Background()

	var id string
	if err := testPool.QueryRow(ctx, "SELECT id FROM users WHERE email=$1", email).Scan(&id); err != nil {
		return
	}
	testPool.Exec(ctx, "DELETE FROM refresh_tokens WHERE user_id=$1", id)
	testPool.Exec(ctx, "DELETE FROM transactions WHERE from_user_id=$1 OR to_user_id=$1", id)
	testPool.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
}

func doJSON(t *testing.T, method, url string, body any, authToken string, out any) (int, map[string]any) {
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	raw := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if out != nil {
		b, _ := json.Marshal(raw)
		if err := json.Unmarshal(b, out); err != nil {
			t.Fatalf("failed to unmarshal response into target: %v", err)
		}
	}

	return resp.StatusCode, raw
}

func registerAndLogin(t *testing.T, srv *httptest.Server, email, password string) auth.LoginResponse {
	t.Helper()

	status, _ := doJSON(t, http.MethodPost, srv.URL+"/auth/register", auth.RegisterRequest{
		Email:           email,
		Password:        password,
		ConfirmPassword: password,
	}, "", nil)
	if status != http.StatusOK {
		t.Fatalf("register failed with status %d", status)
	}
	t.Cleanup(func() { cleanupUserByEmail(t, email) })

	var login auth.LoginResponse
	status, _ = doJSON(t, http.MethodPost, srv.URL+"/auth/login", auth.LoginRequest{
		Email:    email,
		Password: password,
	}, "", &login)
	if status != http.StatusOK {
		t.Fatalf("login failed with status %d", status)
	}

	return login
}

func uniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@example.com", prefix, uuid.NewString())
}

func TestAuthFlow(t *testing.T) {
	srv := newTestServer(t)
	email := uniqueEmail("auth")
	password := "supersecret1"

	login := registerAndLogin(t, srv, email, password)
	if login.Email != email {
		t.Errorf("login.Email = %q, want %q", login.Email, email)
	}
	if login.AuthToken == "" || login.RefreshToken == "" {
		t.Fatal("expected non-empty auth and refresh tokens")
	}

	t.Run("duplicate registration is rejected", func(t *testing.T) {
		status, _ := doJSON(t, http.MethodPost, srv.URL+"/auth/register", auth.RegisterRequest{
			Email:           email,
			Password:        password,
			ConfirmPassword: password,
		}, "", nil)
		if status != http.StatusConflict {
			t.Errorf("status = %d, want %d", status, http.StatusConflict)
		}
	})

	t.Run("login with wrong password is rejected", func(t *testing.T) {
		status, _ := doJSON(t, http.MethodPost, srv.URL+"/auth/login", auth.LoginRequest{
			Email:    email,
			Password: "wrongpassword",
		}, "", nil)
		if status != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", status, http.StatusUnauthorized)
		}
	})

	t.Run("profile requires a valid token", func(t *testing.T) {
		status, _ := doJSON(t, http.MethodGet, srv.URL+"/profile", nil, "", nil)
		if status != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", status, http.StatusUnauthorized)
		}

		var profile user.User
		status, _ = doJSON(t, http.MethodGet, srv.URL+"/profile", nil, login.AuthToken, &profile)
		if status != http.StatusOK {
			t.Fatalf("status = %d, want %d", status, http.StatusOK)
		}
		if profile.Email != email {
			t.Errorf("profile.Email = %q, want %q", profile.Email, email)
		}
	})

	t.Run("refresh token issues a new token pair", func(t *testing.T) {
		var refreshed auth.RefreshTokenResponse
		status, _ := doJSON(t, http.MethodPost, srv.URL+"/auth/refresh-token", auth.RefreshTokenRequest{
			Token: login.RefreshToken,
		}, "", &refreshed)
		if status != http.StatusOK {
			t.Fatalf("status = %d, want %d", status, http.StatusOK)
		}
		if refreshed.AuthToken == "" || refreshed.RefreshToken == "" {
			t.Fatal("expected non-empty refreshed tokens")
		}

		// old refresh token should now be revoked
		status, _ = doJSON(t, http.MethodPost, srv.URL+"/auth/refresh-token", auth.RefreshTokenRequest{
			Token: login.RefreshToken,
		}, "", nil)
		if status != http.StatusUnauthorized {
			t.Errorf("reused refresh token status = %d, want %d", status, http.StatusUnauthorized)
		}
	})
}

func TestWalletFlow(t *testing.T) {
	srv := newTestServer(t)

	senderEmail := uniqueEmail("sender")
	recipientEmail := uniqueEmail("recipient")
	sender := registerAndLogin(t, srv, senderEmail, "supersecret1")
	registerAndLogin(t, srv, recipientEmail, "supersecret1")

	t.Run("balance starts at zero", func(t *testing.T) {
		var balance transaction.BalanceResponse
		status, _ := doJSON(t, http.MethodGet, srv.URL+"/wallet/balance", nil, sender.AuthToken, &balance)
		if status != http.StatusOK {
			t.Fatalf("status = %d, want %d", status, http.StatusOK)
		}
		if balance.Balance != 0 {
			t.Errorf("balance = %v, want 0", balance.Balance)
		}
	})

	var topUpTx transaction.Transaction
	t.Run("top up increases balance", func(t *testing.T) {
		status, _ := doJSON(t, http.MethodPost, srv.URL+"/wallet/topup", transaction.TopUpRequest{Amount: 100}, sender.AuthToken, &topUpTx)
		if status != http.StatusOK {
			t.Fatalf("status = %d, want %d", status, http.StatusOK)
		}

		var balance transaction.BalanceResponse
		doJSON(t, http.MethodGet, srv.URL+"/wallet/balance", nil, sender.AuthToken, &balance)
		if balance.Balance != 100 {
			t.Errorf("balance = %v, want 100", balance.Balance)
		}
	})

	t.Run("transfer moves funds between users", func(t *testing.T) {
		status, _ := doJSON(t, http.MethodPost, srv.URL+"/wallet/transfer", transaction.TransferRequest{
			ToUserEmail: recipientEmail,
			Amount:      30,
		}, sender.AuthToken, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, want %d", status, http.StatusOK)
		}

		var senderBalance, recipientBalance transaction.BalanceResponse
		doJSON(t, http.MethodGet, srv.URL+"/wallet/balance", nil, sender.AuthToken, &senderBalance)
		if senderBalance.Balance != 70 {
			t.Errorf("sender balance = %v, want 70", senderBalance.Balance)
		}

		recipient := registerAndLoginNoCreate(t, srv, recipientEmail, "supersecret1")
		doJSON(t, http.MethodGet, srv.URL+"/wallet/balance", nil, recipient.AuthToken, &recipientBalance)
		if recipientBalance.Balance != 30 {
			t.Errorf("recipient balance = %v, want 30", recipientBalance.Balance)
		}
	})

	t.Run("transfer fails on insufficient balance", func(t *testing.T) {
		status, _ := doJSON(t, http.MethodPost, srv.URL+"/wallet/transfer", transaction.TransferRequest{
			ToUserEmail: recipientEmail,
			Amount:      100000,
		}, sender.AuthToken, nil)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", status, http.StatusBadRequest)
		}
	})

	t.Run("transfer to unknown recipient fails validation", func(t *testing.T) {
		status, _ := doJSON(t, http.MethodPost, srv.URL+"/wallet/transfer", transaction.TransferRequest{
			ToUserEmail: uniqueEmail("nonexistent"),
			Amount:      1,
		}, sender.AuthToken, nil)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", status, http.StatusBadRequest)
		}
	})

	t.Run("list and get transactions", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		url := fmt.Sprintf("%s/wallet/transactions?start_date=%s&end_date=%s", srv.URL, today, today)

		status, raw := doJSON(t, http.MethodGet, url, nil, sender.AuthToken, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, want %d", status, http.StatusOK)
		}
		data, _ := raw["data"].([]any)
		if len(data) == 0 {
			t.Fatal("expected at least one transaction in the list")
		}

		status, _ = doJSON(t, http.MethodGet, srv.URL+"/wallet/transaction/"+topUpTx.ID, nil, sender.AuthToken, nil)
		if status != http.StatusOK {
			t.Fatalf("status = %d, want %d", status, http.StatusOK)
		}

		status, _ = doJSON(t, http.MethodGet, srv.URL+"/wallet/transaction/"+uuid.NewString(), nil, sender.AuthToken, nil)
		if status != http.StatusNotFound {
			t.Errorf("unknown transaction status = %d, want %d", status, http.StatusNotFound)
		}
	})
}

// registerAndLoginNoCreate logs an already-registered user in without
// re-registering (and without adding a duplicate cleanup hook).
func registerAndLoginNoCreate(t *testing.T, srv *httptest.Server, email, password string) auth.LoginResponse {
	t.Helper()
	var login auth.LoginResponse
	status, _ := doJSON(t, http.MethodPost, srv.URL+"/auth/login", auth.LoginRequest{
		Email:    email,
		Password: password,
	}, "", &login)
	if status != http.StatusOK {
		t.Fatalf("login failed with status %d", status)
	}
	return login
}
