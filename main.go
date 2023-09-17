package main

import (
	"context"
	"fmt"
	"github.com/perbu/omg-lol-ipd/config"
	"github.com/perbu/omg-lol-ipd/mon"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	err := realMain()
	if err != nil {
		slog.Info("realMain returned an error", "error", err)
	}
}

func realMain() error {
	configPath := getEnv("CONFIG_PATH", "config.json")
	c, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("config.Load: %w", err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	err = mon.Monitor(ctx, c)
	if err != nil {
		return fmt.Errorf("mon.Monitor: %w", err)
	}
	return nil
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
