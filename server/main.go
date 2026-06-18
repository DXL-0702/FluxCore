package main

import (
	"log"
	"os"

	"github.com/jaxson/FluxCore/server/api"
)

func main() {
	router := api.NewRouter()
	addr := serverAddr()

	log.Printf("starting FluxCore server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func serverAddr() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return ":" + port
}
