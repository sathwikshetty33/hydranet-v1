package agents

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	tetragon "github.com/cilium/tetragon/api/v1/tetragon"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TetragonAgentConfig struct {
	// TetragonAddr is the gRPC target for the FineGuidanceSensors service.
	// Examples:
	//   "unix:///var/run/tetragon/tetragon.sock"  (DaemonSet co-mount)
	//   "127.0.0.1:54321"                         (port-forwarded for dev)
	TetragonAddr string

	// Patterns is the rule table evaluated against every ProcessExec event.
	// Defaults to DefaultPatterns (defined in agent.go) if left empty.
	Patterns []Pattern
}

// TetragonAgent consumes Tetragon events and hands matches to a Reporter.
// It owns no transport to the core — that lives on the injected Reporter,
// shared with any other agent kind (Falco, eBPF, log tail, etc.).
type TetragonAgent struct {
	cfg      TetragonAgentConfig
	reporter Reporter
	logger   *slog.Logger
}

func NewTetragonAgent(cfg TetragonAgentConfig, reporter Reporter, logger *slog.Logger) (*TetragonAgent, error) {
	if cfg.TetragonAddr == "" {
		return nil, errors.New("tetragon address required")
	}
	if reporter == nil {
		return nil, errors.New("reporter required")
	}
	if len(cfg.Patterns) == 0 {
		cfg.Patterns = DefaultPatterns
	}
	return &TetragonAgent{cfg: cfg, reporter: reporter, logger: logger}, nil
}

// ConsumeEvents opens a long-lived gRPC stream to Tetragon and reports any
// matching events through the Reporter until the context is cancelled or
// the stream terminates.
func (a *TetragonAgent) ConsumeEvents(ctx context.Context) error {
	conn, err := grpc.NewClient(
		a.cfg.TetragonAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("dial tetragon at %s: %w", a.cfg.TetragonAddr, err)
	}
	defer conn.Close()

	client := tetragon.NewFineGuidanceSensorsClient(conn)
	stream, err := client.GetEvents(ctx, &tetragon.GetEventsRequest{})
	if err != nil {
		return fmt.Errorf("open tetragon events stream: %w", err)
	}

	a.logger.Info("agent subscribed to tetragon", "addr", a.cfg.TetragonAddr)

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("recv tetragon event: %w", err)
		}

		details, ok := a.matchExec(resp)
		if !ok {
			continue
		}
		if reportErr := a.reporter.Report(ctx, details); reportErr != nil {
			// One bad report shouldn't terminate the stream.
			a.logger.Error("report incident failed",
				"err", reportErr,
				"pod", details.PodName,
				"msg", details.Message)
		}
	}
}

// matchExec inspects a ProcessExec event and returns IncidentDetails for the
// FIRST pattern that fires. Patterns are evaluated in their slice order, so
// put highest-severity, lowest-false-positive rules first.
// Events without an associated pod are skipped — the agent only cares about
// activity inside workloads.
func (a *TetragonAgent) matchExec(resp *tetragon.GetEventsResponse) (IncidentDetails, bool) {
	exec := resp.GetProcessExec()
	if exec == nil || exec.Process == nil || exec.Process.Pod == nil {
		return IncidentDetails{}, false
	}
	proc := exec.Process

	for _, p := range a.cfg.Patterns {
		if !p.Matches(proc.Binary, proc.Arguments) {
			continue
		}
		workload := proc.Pod.Workload
		if workload == "" {
			workload = proc.Pod.Name
		}
		// Technique is tucked into the message as a parseable prefix until
		// the IncidentReport proto gains a dedicated field.
		return IncidentDetails{
			PodName:     proc.Pod.Name,
			ServiceName: workload,
			Message:     fmt.Sprintf("[%s] %s %s", p.Technique, proc.Binary, proc.Arguments),
			Severity:    p.Severity,
		}, true
	}
	return IncidentDetails{}, false
}

// Compile-time assertion that TetragonAgent satisfies the Agent interface.
var _ Agent = (*TetragonAgent)(nil)
