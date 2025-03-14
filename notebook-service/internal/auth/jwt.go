package auth

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	UNKNOWN = "UNKNOWN"
	ADMIN   = "ADMIN"
	DS      = "DS"
)

type key string

const CtxKey key = "username"

type JWTClaims struct {
	Username string `json:"iss"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func ValidateToken(tokenString string) (*JWTClaims, error) {
	jwtSecret := os.Getenv("SECRET_KEY")

	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}

func AuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	// Extract the token from the metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Get the token from the "authorization" metadata
	token := md["authorization"]
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	// Validate the token
	claims, err := ValidateToken(token[0][7:]) // Assume Bearer token
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Set the username in the context for later use
	ctx = context.WithValue(ctx, CtxKey, claims.Username)

	return handler(ctx, req)
}
