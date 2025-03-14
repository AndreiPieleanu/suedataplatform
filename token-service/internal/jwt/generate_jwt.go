package jwt

import (
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Function to generate token
var GenerateToken = func(key, role string) (string, error) {
	secret := os.Getenv("SECRET_KEY")

	// Create claims
	claims := jwt.MapClaims{
		"iss":          key,
		"role":         role,
		"expired_time": time.Now().Add(time.Hour * 24).Unix(),
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token
	signedToken, err := token.SignedString([]byte(secret))

	if err != nil {
		return "", err
	}

	return signedToken, nil
}
