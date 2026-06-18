package main

import (
	"log"

	"github.com/jaxson/FluxCore/server/api"
	"github.com/jaxson/FluxCore/server/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	router := api.NewRouter(cfg)
	addr := cfg.Server.Address

	log.Printf("starting FluxCore server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
