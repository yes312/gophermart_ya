package jwtpackage

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	TokenExp time.Duration
	Secret   string
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

var ErrInvalidToken = errors.New("invalid token")

func NewToken(tokenExp time.Duration, secret string) *Token {
	return &Token{
		TokenExp: tokenExp,
		Secret:   secret,
	}
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (tok *Token) BuildJWTString(UserID string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tok.TokenExp)),
		},

		UserID: UserID,
	})

	tokenString, err := token.SignedString([]byte(tok.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (tok *Token) GetUserID(tokenString string) (string, error) {

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(tok.Secret), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	return claims.UserID, err
}
