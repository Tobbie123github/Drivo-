package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ResetPasswordClaims struct {
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

type Claims struct {
	jwt.RegisteredClaims

	Role string `json:"role"`
}

func CreateToken(jwtSecret string, userID string, role string) (string, error) {

	now := time.Now().UTC()
	exp := now.Add(7 * 24 * time.Hour)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		Role: role,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokStr, err := tok.SignedString([]byte(jwtSecret))

	if err != nil {
		return "", fmt.Errorf("Error generating token: %s", err)
	}

	return tokStr, err
}

// verify Token

func VerifyToken(jwtSecret string, tokenString string) (Claims, error) {

	var claims Claims

	parsedTkn, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(jwtSecret), nil
	},

		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil {
		return Claims{}, fmt.Errorf("parsed token failed: %v", err)
	}

	if !parsedTkn.Valid {
		return Claims{}, fmt.Errorf("invalid token")
	}

	if claims.Subject == "" {
		return Claims{}, fmt.Errorf("Token missing subject")
	}

	return claims, nil

}

func GenerateResetOTPToken(jwtSecret string, userID string) (string, error) {

	now := time.Now().UTC()
	exp := now.Add(15 * time.Minute)

	claims := ResetPasswordClaims{
		Purpose: "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtSecret))

}

func VerifyResetOTPToken(jwtSecret string, tokenString string) (ResetPasswordClaims, error) {

	var claims ResetPasswordClaims

	parsedTkn, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !parsedTkn.Valid {
		return ResetPasswordClaims{}, fmt.Errorf("invalid or expired token")

	}

	if claims.Purpose != "password_reset" {
		return ResetPasswordClaims{}, fmt.Errorf("invalid token purpose")
	}

	return claims, nil

}
