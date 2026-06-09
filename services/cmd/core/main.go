package main

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/sathwikshetty33/hydranet/services/database"
	grpcserver "github.com/sathwikshetty33/hydranet/services/grpc_server"
	httpserver "github.com/sathwikshetty33/hydranet/services/http_server"
	"github.com/sathwikshetty33/hydranet/services/utils"
)

func main() {
	logger := utils.NewLogger(slog.LevelInfo)

	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		logger.Warn("failed to load .env", "err", err)
	}

	db, err := database.NewDatabase(
		os.Getenv("HYDRANET_DB_HOST"),
		os.Getenv("HYDRANET_DB_USER"),
		os.Getenv("HYDRANET_DB_PASSWORD"),
		os.Getenv("HYDRANET_DB_NAME"),
		os.Getenv("HYDRANET_DB_PORT"),
	)
	if err != nil {
		logger.Error("database init failed", "err", err)
		os.Exit(1)
	}

	grpc_server := grpcserver.NewGrpcServer(db, logger)
	http_server := httpserver.NewHttpServer(db, logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	addr := envOr("HYDRANET_GRPC_ADDR", ":50051")
	
	go func() {
		if err := grpc_server.Serve(ctx, addr); err != nil {
			logger.Error("grpc server failed", "err", err)
			cancel()
		}
	}()

	httpAddr := envOr("HYDRANET_HTTP_ADDR", ":8080")
	go func() {
		logger.Info("http server listening", "addr", httpAddr)
		if err := http_server.Router.Run(httpAddr); err != nil {
			logger.Error("http server failed", "err", err)
			cancel()
		}
	}()
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
