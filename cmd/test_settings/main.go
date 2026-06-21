package main

import (
	"context"
	"fmt"
	"log"

	"github.com/claudio-nehemia/interior_backend/internal/config"
	"github.com/claudio-nehemia/interior_backend/pkg/cache"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	cacheStore := cache.NewRedis(cfg.RedisAddr(), cfg.RedisPassword)
	ctx := context.Background()

	// Ping
	if err := cacheStore.Ping(ctx); err != nil {
		log.Fatal("Redis ping failed: ", err)
	}

	// Delete key
	key := "setting:all"
	err = cacheStore.Del(ctx, key)
	if err != nil {
		log.Fatal("Failed to delete settings cache key: ", err)
	}

	fmt.Println("Successfully cleared settings cache key:", key)
}
