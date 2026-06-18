package main

import (
	"context"
	"log"
	"time"

	"github.com/jaxson/FluxCore/server/api"
	"github.com/jaxson/FluxCore/server/config"
	"github.com/jaxson/FluxCore/server/db"
	"github.com/jaxson/FluxCore/server/service"
)

const databasePingTimeout = 5 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	router := api.NewRouter(cfg)
	addr := cfg.Server.Address
	conn, err := db.Open(cfg.Database)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() {
		if err := db.Close(conn); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), databasePingTimeout)
	defer cancel()
	if err := db.Ping(ctx, conn); err != nil {
		log.Fatalf("ping database: %v", err)
	}
	if err := service.Migrate(conn); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	log.Printf("starting FluxCore server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
