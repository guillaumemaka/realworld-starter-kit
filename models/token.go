package models

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var tokenSecret = []byte("All-your-base")

// Token represents the JWT token
type Token string

// TokenClaims is a custom claims struct for JWT
type TokenClaims struct {
	User *User
	jwt.StandardClaims
}

const jwtExpiryDuration = time.Hour * 24 * 7 // ~7 days

// NewToken generates a new JWT token, using User
func NewToken(u *User) (string, error) {
	claims := TokenClaims{
		u,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(jwtExpiryDuration).Unix(),
			Issuer:    "golang-gorilla-conduit",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(tokenSecret)
	if err != nil {
		return "", err
	}
	return ss, nil
}

// ValidateToken validates the JWT and returns the claims
func ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return tokenSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("Token not valid")
	}
	return claims, nil
}
