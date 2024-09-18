package utils

// Functions to generate and validate JWT tokens.

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateToken(id int, role string) (string, error) {
	hasura_claims := map[string]interface{}{
		"x-hasura-allowed-roles": []string{role},
		"x-hasura-default-role":  role,
		"x-hasura-user-id":       id,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Role":                         role,
		"exp":                          time.Now().Add(time.Hour * 3).Unix(),
		"https://hasura.io/jwt/claims": hasura_claims,
	})

	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return jwt.MapClaims{}, errors.New("invalid JWT")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return jwt.MapClaims{}, errors.New("invalid JWT claims")
	}

	return claims, nil
}
