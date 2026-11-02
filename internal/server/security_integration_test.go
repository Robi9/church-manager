package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Robi9/church-manager/internal/config"
	"github.com/Robi9/church-manager/internal/modules/auth"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Run against a migrated, disposable local database only. Never production.
func TestSecurityIntegration(t *testing.T) {
	connection := os.Getenv("TEST_DATABASE_URL")
	if connection == "" {
		t.Skip("TEST_DATABASE_URL not configured")
	}
	parsed, err := url.Parse(connection)
	if err != nil || parsed.Hostname() != "127.0.0.1" || !strings.HasSuffix(parsed.Path, "_security_test") {
		t.Fatal("security integration tests require a local *_security_test database")
	}
	db, err := sql.Open("postgres", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := auth.NewRepository(db)
	secret := strings.Repeat("test-secret-", 4)
	service := auth.NewService(repo, secret)
	email := fmt.Sprintf("security-%d@example.org", time.Now().UnixNano())
	user, err := service.Register(email, "test-password-only-123")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
	gin.SetMode(gin.TestMode)
	router := SetupRouter(db, &config.Config{JWTSecret: secret, AllowedOrigins: []string{"https://church.example.org"}})
	request := func(method, path, body, token string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		return res
	}
	login := func() string {
		t.Helper()
		payload, _ := json.Marshal(map[string]string{"email": email, "password": "test-password-only-123"})
		res := request("POST", "/api/auth/login", string(payload), "")
		if res.Code != 200 {
			t.Fatalf("login status=%d body=%s", res.Code, res.Body.String())
		}
		var result struct {
			Data struct {
				Token string `json:"token"`
			} `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result.Data.Token
	}
	assertStatus := func(token string, expected int) {
		t.Helper()
		res := request("GET", "/api/auth/me", "", token)
		if res.Code != expected {
			t.Fatalf("session status=%d want %d", res.Code, expected)
		}
	}
	for _, path := range []string{"/api/members", "/api/dashboard/stats"} {
		if res := request("GET", path, "", ""); res.Code != 401 {
			t.Fatalf("unprotected %s", path)
		}
	}
	if res := request("POST", "/api/auth/register", "{}", ""); res.Code != 404 {
		t.Fatal("registration exposed")
	}
	first, second := login(), login()
	assertStatus(first, 200)
	assertStatus(second, 200)
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE token_hash = $1`, first).Scan(&count); err != nil || count != 0 {
		t.Fatal("raw session token stored in database")
	}
	if res := request("POST", "/api/auth/logout-all", "", first); res.Code != 200 {
		t.Fatal("logout failed")
	}
	assertStatus(first, 401)
	assertStatus(second, 401)
	third := login()
	if _, err := db.Exec(`UPDATE users SET disabled_at = NOW() WHERE id = $1`, user.ID); err != nil {
		t.Fatal(err)
	}
	assertStatus(third, 401)
	if _, err := db.Exec(`UPDATE users SET disabled_at = NULL, updated_at = NOW() WHERE id = $1`, user.ID); err != nil {
		t.Fatal(err)
	}
	assertStatus(third, 401) // Updating credentials must not revive the old session.
	fourth := login()
	assertStatus(fourth, 200)
	if _, err := db.Exec(`UPDATE sessions SET expires_at = NOW() - INTERVAL '1 second' WHERE user_id = $1`, user.ID); err != nil {
		t.Fatal(err)
	}
	assertStatus(fourth, 401)
	fifth := login()
	assertStatus(fifth, 200)
	if res := request("POST", "/api/auth/logout", "", fifth); res.Code != 200 {
		t.Fatal("single-session logout failed")
	}
	assertStatus(fifth, 401)
}
