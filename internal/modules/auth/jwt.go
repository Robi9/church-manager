package auth

import "github.com/golang-jwt/jwt/v5"

func generateToken(userID int64) (string, error) {
	secret := []byte("secret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
	})

	return token.SignedString(secret)
}
