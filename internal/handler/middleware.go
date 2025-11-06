package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is used for safe value storage in context.
type contextKey string

// UserIDKey is the key for storing user ID in request context.
const UserIDKey contextKey = "userID"

// JWTMiddleware creates middleware for JWT token validation in Authorization header.
// Extracts user ID from token and adds it to request context.
// Requires header format: "Bearer <token>".
func JWTMiddleware(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing auth token", http.StatusUnauthorized)
				return
			}

			// Extract token from header (format: "Bearer <token>")
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "invalid auth token format", http.StatusUnauthorized)
				return
			}

			// Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Check signing method (must be HMAC)
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, http.ErrAbortHandler
				}
				return jwtSecret, nil
			})

			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			// Extract user ID from claims and add to context
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				userID, ok := claims["sub"].(string)
				if !ok {
					http.Error(w, "invalid token claims", http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), UserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "invalid token", http.StatusUnauthorized)
			}
		})
	}
}
