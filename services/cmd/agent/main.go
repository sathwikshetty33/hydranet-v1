package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sathwikshetty33/hydranet/services/agents"
	"github.com/sathwikshetty33/hydranet/services/utils"
)

func main() {
	logger := utils.NewLogger(slog.LevelInfo)

	cfg, err := configFromEnv()
	if err != nil {
		logger.Error("invalid agent config", "err", err)
		os.Exit(1)
	}

	agent, err := agents.NewTetragonAgent(cfg, logger)
	if err != nil {
		logger.Error("agent init failed", "err", err)
		os.Exit(1)
	}
	defer agent.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	runWithBackoff(ctx, logger, agent.ConsumeEvents)
}

func configFromEnv() (agents.TetragonAgentConfig, error) {
	cfg := agents.TetragonAgentConfig{
		TetragonAddr: envOr("HYDRANET_TETRAGON_ADDR", "unix:///var/run/tetragon/tetragon.sock"),
		CoreAddr:     envOr("HYDRANET_CORE_ADDR", "localhost:50051"),
		AgentID:      os.Getenv("HYDRANET_AGENT_ID"),
		AuthToken:    os.Getenv("HYDRANET_AGENT_TOKEN"),
	}
	if cfg.AgentID == "" {
		return cfg, errors.New("HYDRANET_AGENT_ID is required")
	}
	if cfg.AuthToken == "" {
		return cfg, errors.New("HYDRANET_AGENT_TOKEN is required")
	}
	return cfg, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// runWithBackoff calls run in a loop, re-dialing Tetragon with exponential
// backoff whenever the stream returns. Stops when the context is cancelled.
func runWithBackoff(ctx context.Context, logger *slog.Logger, run func(context.Context) error) {
	const (
		baseDelay = time.Second
		maxDelay  = 30 * time.Second
	)
	delay := baseDelay
	for {
		err := run(ctx)
		if ctx.Err() != nil {
			logger.Info("agent stopped", "reason", ctx.Err())
			return
		}
		logger.Error("agent stream ended; reconnecting", "err", err, "after", delay)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}
