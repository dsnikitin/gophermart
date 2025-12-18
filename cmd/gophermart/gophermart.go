package main

// import (
// 	"context"
// 	"errors"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/dsnikitin/gophermart/internal/config"
// 	"github.com/dsnikitin/gophermart/internal/handler"
// 	"github.com/dsnikitin/gophermart/internal/pkg/db"
// 	"github.com/dsnikitin/gophermart/internal/pkg/logger"
// 	"github.com/dsnikitin/gophermart/internal/repository"
// 	"github.com/dsnikitin/gophermart/internal/server"
// 	"github.com/dsnikitin/gophermart/internal/service"
// )

// var (
// 	startError     chan error
// 	shutdownSignal chan os.Signal
// )

// func init() {
// 	startError = make(chan error, 1)
// 	shutdownSignal = make(chan os.Signal, 1)
// 	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
// }

// func main() {
// 	cfg, err := config.New()
// 	if err != nil {
// 		logger.Log.Fatalw("Failed to init config", "error", err)
// 	}

// 	if err = logger.Setup(cfg.Log); err != nil {
// 		logger.Log.Fatalw("Failed to init logger", "error", err)
// 	}

// 	pgDB, err := db.Connect(cfg.DB)
// 	if err != nil {
// 		logger.Log.Fatalw("Failed to init db", "error", err)
// 	}
// 	defer pgDB.Close()

// 	if err := db.ApplyMigrations(cfg.DB); err != nil {
// 		logger.Log.Fatalw("Failed to apply migrations", "error", err)
// 	}
// 	r := repository.New(pgDB)
// 	s := service.New(r)
// 	h := handler.New(cfg, s)
// 	httpServer := server.New(cfg, h)

// 	go func() {
// 		logger.Log.Infow("Starting server", "address", cfg.ServerAddr)
// 		startError <- httpServer.ListenAndServe()
// 	}()

// 	gracefulShutdown(httpServer)
// }

// func gracefulShutdown(httpServer *server.Server) {
// 	select {
// 	case err := <-startError:
// 		if err != nil {
// 			if !errors.Is(err, http.ErrServerClosed) {
// 				logger.Log.Fatalw("Failed to start server", "error", err)
// 			} else {
// 				logger.Log.Info("Server was stopped")
// 			}
// 		}
// 	case <-shutdownSignal:
// 		logger.Log.Info("Received shutdown signal")
// 		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
// 		defer cancel()

// 		if err := httpServer.Shutdown(ctx); err != nil {
// 			logger.Log.Errorw("Failed to stop server gracefully", "error", err)
// 		} else {
// 			logger.Log.Info("Server stopped gracefully")
// 		}
// 	}
// }

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dsnikitin/gophermart/internal/config"
	"github.com/dsnikitin/gophermart/internal/app"
	"github.com/dsnikitin/gophermart/internal/pkg/db"
	"github.com/dsnikitin/gophermart/internal/pkg/logger"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		logger.Log.Fatalw("Failed to init config", "error", err)
	}

	if err = logger.Setup(cfg.Log); err != nil {
		logger.Log.Fatalw("Failed to setup logger", "error", err)
	}

	database, err := db.Connect(cfg.DB)
	if err != nil {
		logger.Log.Fatalw("Failed to connect to db", "error", err)
	}
	defer database.Close()

	if err := db.ApplyMigrations(cfg.DB); err != nil {
		logger.Log.Fatalw("Failed to apply migrations", "error", err)
	}

	application := app.New(cfg, database)

	go application.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.Shutdown()
}
