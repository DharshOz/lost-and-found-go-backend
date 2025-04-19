package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateJWT(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	// Parse the token and validate its signature and claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure that the token is signed using the HMAC method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil {
		log.Println("Error parsing token:", err)
		return nil, err
	}

	// Check token claims for validity
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Debugging: Log the claims for validation
		log.Println("Token Claims:", claims)

		// Ensure that "userId" is present in the claims
		if userID, ok := claims["userId"].(string); ok {
			// Successfully extracted userId
			log.Println("Extracted User ID:", userID)
		} else {
			// Handle missing or invalid userId claim
			log.Println("Error: Missing or invalid userId claim")
		}
	} else {
		return nil, fmt.Errorf("invalid token claims or invalid token")
	}

	return token, nil
}
