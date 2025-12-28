package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dsnikitin/gophermart/internal/app"
	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/pkg/db"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		logger.Log.Fatalw("Failed to init config", "error", err.Error())
	}

	if err = logger.Setup(cfg.Log); err != nil {
		logger.Log.Fatalw("Failed to setup logger", "error", err.Error())
	}

	database, err := db.Connect(cfg.DB)
	if err != nil {
		logger.Log.Fatalw("Failed to connect to db", "error", err.Error())
	}
	defer database.Close()

	if err := db.ApplyMigrations(cfg.DB); err != nil {
		logger.Log.Fatalw("Failed to apply migrations", "error", err.Error())
	}

	application := app.New(cfg, database)

	go application.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.Shutdown()
}
