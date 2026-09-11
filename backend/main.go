package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"btech-wallet/auth"
	"btech-wallet/config"
	"btech-wallet/server"
	"btech-wallet/transaction"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	port := config.LoadEnv("PORT", "3333")
	dbURL := config.LoadEnv("DATABASE_URL", "postgres://user:password@localhost:5432/mydb")
	JWTSecret := config.LoadEnv("JWT_SECRET", "your_jwt_secret_key")

	ctx := context.Background()
	mux := http.NewServeMux()

	pool, err := config.InitDB(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping the database: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"message": "Hello World!",
		}
		server.JSON(w, http.StatusOK, response)
	})

	jwtManager := auth.NewJWTManager(JWTSecret, 3600*time.Second) // 3600 seconds = 1 hour

	authHandler := auth.NewAuthHandler(pool, jwtManager)
	transactionHandler := transaction.NewTransactionHandler(pool)

	authHandler.RegisterRoutes(mux)
	transactionHandler.RegisterRoutes(mux)

	log.Printf("Server is running on port %s", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
