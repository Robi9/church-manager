package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Robi9/church-manager/internal/modules/auth"
	"github.com/gin-gonic/gin"
)

type sessionStub struct {
	digest  string
	revoked bool
	err     error
}

func (s *sessionStub) ValidateSession(_ context.Context, digest string) (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	if s.revoked || digest != s.digest {
		return 0, sql.ErrNoRows
	}
	return 9, nil
}
func TestSessionAuthenticationAndRevocation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	token, secret := strings.Repeat("a", 43), "test-secret"
	store := &sessionStub{digest: auth.SessionDigest(token, secret)}
	router := gin.New()
	router.GET("/private", AuthMiddleware(secret, store), func(c *gin.Context) {
		if c.GetInt64("user_id") != 9 {
			t.Error("wrong authenticated user")
		}
		c.Status(200)
	})
	check := func(token string, status int) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/private", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != status {
			t.Fatalf("status=%d want %d", res.Code, status)
		}
	}
	check("", 401)
	check("old.jwt.token", 401)
	check(strings.Repeat("b", 43), 401)
	check(token, 200)
	store.revoked = true
	check(token, 401)
	store.err = errors.New("database unavailable")
	check(token, 503)
}
