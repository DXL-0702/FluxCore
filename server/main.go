package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jaxson/FluxCore/server/api"
	"github.com/jaxson/FluxCore/server/config"
	"github.com/jaxson/FluxCore/server/db"
	"github.com/jaxson/FluxCore/server/service"
)

const databasePingTimeout = 5 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal error: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	router := api.NewRouter(cfg)
	addr := cfg.Server.Address
	conn, err := db.Open(cfg.Database)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if err := db.Close(conn); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), databasePingTimeout)
	defer cancel()
	if err := db.Ping(ctx, conn); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	if err := service.Migrate(conn); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	log.Printf("starting FluxCore server on %s", addr)
	if err := router.Run(addr); err != nil {
		return fmt.Errorf("server stopped: %w", err)
	}

	return nil
}
