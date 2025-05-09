// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	//"time"

	"github.com/gorilla/mux"

	"tetris-be/api"
	_ "tetris-be/handler"
	"tetris-be/util"
)

func main() {
	// Load config
	config, err := util.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Router setup
	r := mux.NewRouter()

	// REST API endpoints
	api.RegisterRoutes(r)

	// Middleware
	//r.Use(loggingMiddleware)

	// Server setup
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%d", config.Server.Port),
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
	}

	log.Printf("Server listening on port %d", config.Server.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
