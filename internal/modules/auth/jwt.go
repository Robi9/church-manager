package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenExpiration = 24 * time.Hour

func generateToken(userID int64, secret string) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"iat":     now.Unix(),
		"exp":     now.Add(tokenExpiration).Unix(),
	})

	return token.SignedString([]byte(secret))
}
