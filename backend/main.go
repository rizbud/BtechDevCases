package main

import (
	"log"
	"net/http"

	"btech-wallet/config"
	"btech-wallet/server"
)

func main() {
	port := config.LoadEnv("PORT", "3333")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"message": "Hello World!",
		}
		server.JSON(w, http.StatusOK, response)
	})

	log.Printf("Server is running on port %s", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
