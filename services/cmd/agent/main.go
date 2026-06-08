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

	tetragonCfg, reporterCfg, err := configFromEnv()
	if err != nil {
		logger.Error("invalid agent config", "err", err)
		os.Exit(1)
	}

	reporter, err := agents.NewCoreReporter(reporterCfg)
	if err != nil {
		logger.Error("reporter init failed", "err", err)
		os.Exit(1)
	}
	defer reporter.Close()

	agent, err := agents.NewTetragonAgent(tetragonCfg, reporter, logger)
	if err != nil {
		logger.Error("agent init failed", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	runWithBackoff(ctx, logger, agent.ConsumeEvents)
}

func configFromEnv() (agents.TetragonAgentConfig, agents.CoreReporterConfig, error) {
	tetragonCfg := agents.TetragonAgentConfig{
		TetragonAddr: envOr("HYDRANET_TETRAGON_ADDR", "unix:///var/run/tetragon/tetragon.sock"),
	}
	reporterCfg := agents.CoreReporterConfig{
		CoreAddr:  envOr("HYDRANET_CORE_ADDR", "localhost:50051"),
		AgentID:   os.Getenv("HYDRANET_AGENT_ID"),
		AuthToken: os.Getenv("HYDRANET_AGENT_TOKEN"),
	}
	if reporterCfg.AgentID == "" {
		return tetragonCfg, reporterCfg, errors.New("HYDRANET_AGENT_ID is required")
	}
	if reporterCfg.AuthToken == "" {
		return tetragonCfg, reporterCfg, errors.New("HYDRANET_AGENT_TOKEN is required")
	}
	return tetragonCfg, reporterCfg, nil
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
