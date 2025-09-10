package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	store "tetris-be/internal/store/room"
	"time"

	"net/http"
)

func run(ctx context.Context) error {

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM|syscall.SIGINT)
	defer cancel()

	serverHandler := NewServerHandler(logger, nil, nil)
	cfg := LoadConfig()
	//v1 := http.NewServeMux()
	//v1.Handle("/v1/", http.StripPrefix("/v1", mux))
	httpServer := http.Server{
		Addr:    net.JoinHostPort(cfg.host, strconv.Itoa(cfg.port)),
		Handler: serverHandler,
	}
	logger.Info(fmt.Sprintf("Starting httpServer on port %d:...", cfg.port))
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			logger.Error(fmt.Sprintf("error listening and serving: %s\n", err))
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	//graceful shutdown
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("error when shutting down: ", "error", err.Error())
		}
	}()
	wg.Wait()
	return nil
}

type Config struct {
	host string
	port int
}

func NewServerHandler(logger *slog.Logger, cfg *Config, roomStore store.RoomsStore) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger, cfg, roomStore)

	return mux
}
