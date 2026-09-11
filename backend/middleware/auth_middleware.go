package middleware

import (
	"context"
	"net/http"
	"strings"

	"btech-wallet/config"
	"btech-wallet/server"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	JWTSecret := config.LoadEnv("JWT_SECRET", "your_jwt_secret_key")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			server.ErrorResponseJSON(w, http.StatusUnauthorized, "Missing Authorization header", nil)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			server.ErrorResponseJSON(w, http.StatusUnauthorized, "Invalid Authorization header format", nil)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return []byte(JWTSecret), nil
		})

		if err != nil || !token.Valid {
			server.ErrorResponseJSON(w, http.StatusUnauthorized, "Invalid token", nil)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			server.ErrorResponseJSON(w, http.StatusUnauthorized, "Invalid token", nil)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			server.ErrorResponseJSON(w, http.StatusUnauthorized, "Invalid token", nil)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
