package controllers

import (
	"errors"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type customerClaims struct {
	CustomerID uuid.UUID
	CustomerNo uint32
	FirstName  string
	LastName   string
}

func extractCustomerClaims(c *gin.Context) (*customerClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.New("missing token")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "changeme-use-env-in-production"
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	if tt, ok := claims["token_type"].(string); !ok || tt != "customer_access" {
		return nil, errors.New("customer access token required")
	}

	customerIdStr, ok := claims["customer_id"].(string)
	if !ok {
		return nil, errors.New("customer_id not in claims")
	}
	customerId, err := uuid.Parse(customerIdStr)
	if err != nil {
		return nil, errors.New("invalid customer_id")
	}

	customerNoFloat, ok := claims["customer_no"].(float64)
	if !ok {
		return nil, errors.New("customer_no not in claims")
	}

	firstName, _ := claims["first_name"].(string)
	lastName, _ := claims["last_name"].(string)

	return &customerClaims{
		CustomerID: customerId,
		CustomerNo: uint32(customerNoFloat),
		FirstName:  firstName,
		LastName:   lastName,
	}, nil
}
