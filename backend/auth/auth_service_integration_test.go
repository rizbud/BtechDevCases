package auth

import (
	"sync"
	"testing"
	"time"

	"btech-wallet/testutil"
	"btech-wallet/user"

	"github.com/google/uuid"
)

func newTestAuthService() *AuthService {
	return &AuthService{
		AuthRepository: &AuthRepository{DB: testPool},
		UserRepository: &user.UserRepository{DB: testPool},
		JWTManager:     NewJWTManager("test-secret", time.Hour),
	}
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	s := newTestAuthService()
	email := uniqueEmail("service-register")
	t.Cleanup(func() { testutil.CleanupUserByEmail(testPool, email) })

	if err := s.register(t.Context(), email, "supersecret1"); err != nil {
		t.Fatalf("register() error = %v", err)
	}

	got, err := s.login(t.Context(), email, "supersecret1")
	if err != nil {
		t.Fatalf("login() error = %v", err)
	}
	if got.Email != email {
		t.Errorf("Email = %q, want %q", got.Email, email)
	}
	if got.AuthToken == "" || got.RefreshToken == "" {
		t.Fatal("expected non-empty auth and refresh tokens")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	s := newTestAuthService()
	email := uniqueEmail("service-login-wrong")
	t.Cleanup(func() { testutil.CleanupUserByEmail(testPool, email) })

	if err := s.register(t.Context(), email, "supersecret1"); err != nil {
		t.Fatalf("register() error = %v", err)
	}

	_, err := s.login(t.Context(), email, "wrongpassword")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if err.Error() != "Email or password is incorrect" {
		t.Errorf("error = %q, want %q", err.Error(), "Email or password is incorrect")
	}
}

func TestAuthService_Login_UnknownEmail(t *testing.T) {
	s := newTestAuthService()

	_, err := s.login(t.Context(), uniqueEmail("nobody"), "supersecret1")
	if err == nil {
		t.Fatal("expected error for unknown email, got nil")
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	s := newTestAuthService()
	email := uniqueEmail("service-refresh")
	t.Cleanup(func() { testutil.CleanupUserByEmail(testPool, email) })

	if err := s.register(t.Context(), email, "supersecret1"); err != nil {
		t.Fatalf("register() error = %v", err)
	}
	login, err := s.login(t.Context(), email, "supersecret1")
	if err != nil {
		t.Fatalf("login() error = %v", err)
	}

	refreshed, err := s.refreshToken(t.Context(), login.RefreshToken)
	if err != nil {
		t.Fatalf("refreshToken() error = %v", err)
	}
	if refreshed.AuthToken == "" || refreshed.RefreshToken == "" {
		t.Fatal("expected non-empty refreshed tokens")
	}

	// the old refresh token must now be revoked
	if _, err := s.refreshToken(t.Context(), login.RefreshToken); err == nil {
		t.Fatal("expected error when reusing a revoked refresh token, got nil")
	}
}

func TestAuthService_RefreshToken_ConcurrentReuse(t *testing.T) {
	s := newTestAuthService()
	email := uniqueEmail("service-refresh-race")
	t.Cleanup(func() { testutil.CleanupUserByEmail(testPool, email) })

	if err := s.register(t.Context(), email, "supersecret1"); err != nil {
		t.Fatalf("register() error = %v", err)
	}
	login, err := s.login(t.Context(), email, "supersecret1")
	if err != nil {
		t.Fatalf("login() error = %v", err)
	}

	var wg sync.WaitGroup
	results := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, results[0] = s.refreshToken(t.Context(), login.RefreshToken)
	}()
	go func() {
		defer wg.Done()
		_, results[1] = s.refreshToken(t.Context(), login.RefreshToken)
	}()
	wg.Wait()

	successes := 0
	for _, err := range results {
		if err == nil {
			successes++
		}
	}

	if successes != 1 {
		t.Errorf("successful concurrent refreshes using the same token = %d, want exactly 1 (refreshToken does not serialize the read-then-revoke against a concurrent call)", successes)
	}
}

func TestAuthService_RefreshToken_UnknownToken(t *testing.T) {
	s := newTestAuthService()

	_, err := s.refreshToken(t.Context(), uuid.NewString())
	if err == nil {
		t.Fatal("expected error for unknown refresh token, got nil")
	}
	if err.Error() != "Refresh token not found or expired" {
		t.Errorf("error = %q, want %q", err.Error(), "Refresh token not found or expired")
	}
}
