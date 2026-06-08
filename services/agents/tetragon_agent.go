package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

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

	// CoreURL is the HydraNet core base URL, no trailing slash, no /api/v1.
	//   "https://hydranet.example"  /  "http://localhost:8080"
	CoreURL string

	// AgentID identifies this agent to the core (agents.id PK as a string).
	AgentID string

	// AuthToken is the raw bearer token; the core hashes and compares.
	AuthToken string

	// SuspiciousPatterns are substrings matched against the exec'd
	// binary + arguments. Defaults to a sensible set if left empty.
	SuspiciousPatterns []string

	// HTTPClient is used for reporting. Defaults to a 10s-timeout client.
	HTTPClient *http.Client
}

type TetragonAgent struct {
	cfg    TetragonAgentConfig
	logger *slog.Logger
}

func NewTetragonAgent(cfg TetragonAgentConfig, logger *slog.Logger) *TetragonAgent {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	if len(cfg.SuspiciousPatterns) == 0 {
		cfg.SuspiciousPatterns = []string{
			"nc", "ncat",
			"wget", "curl",
			"/etc/shadow", "/etc/passwd",
			"/proc/self/status",
			"capsh",
		}
	}
	return &TetragonAgent{cfg: cfg, logger: logger}
}

// ConsumeEvents opens a long-lived gRPC stream to Tetragon and reports any
// matching events as incidents until the context is cancelled or the stream
// terminates.
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
		if reportErr := a.ReportIncident(ctx, details); reportErr != nil {
			// One bad report shouldn't terminate the stream.
			a.logger.Error("report incident failed",
				"err", reportErr,
				"pod", details.PodName,
				"msg", details.Message)
		}
	}
}

// matchExec inspects a ProcessExec event and returns IncidentDetails when
// the binary or its arguments contain any configured suspicious pattern.
// Events without an associated pod are skipped — the agent only cares
// about activity inside workloads.
func (a *TetragonAgent) matchExec(resp *tetragon.GetEventsResponse) (IncidentDetails, bool) {
	exec := resp.GetProcessExec()
	if exec == nil || exec.Process == nil || exec.Process.Pod == nil {
		return IncidentDetails{}, false
	}
	proc := exec.Process

	haystack := proc.Binary + " " + proc.Arguments
	for _, needle := range a.cfg.SuspiciousPatterns {
		if needle == "" || !strings.Contains(haystack, needle) {
			continue
		}
		workload := proc.Pod.Workload
		if workload == "" {
			workload = proc.Pod.Name
		}
		return IncidentDetails{
			PodName:     proc.Pod.Name,
			ServiceName: workload,
			Message:     fmt.Sprintf("Tetragon: %s %s", proc.Binary, proc.Arguments),
			Severity:    "high",
		}, true
	}
	return IncidentDetails{}, false
}

// ReportIncident POSTs an incident to the core. Auth is a bearer token; the
// agent id is part of the URL so the core knows who is reporting.
func (a *TetragonAgent) ReportIncident(ctx context.Context, incident IncidentDetails) error {
	if a.cfg.CoreURL == "" {
		return fmt.Errorf("core url not configured")
	}
	if a.cfg.AgentID == "" {
		return fmt.Errorf("agent id not configured")
	}

	payload, err := json.Marshal(incident)
	if err != nil {
		return fmt.Errorf("marshal incident: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/agents/%s/incidents",
		strings.TrimRight(a.cfg.CoreURL, "/"), a.cfg.AgentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if a.cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+a.cfg.AuthToken)
	}

	res, err := a.cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("post incident: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("core rejected incident (%d): %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	a.logger.Debug("incident reported", "pod", incident.PodName, "status", res.StatusCode)
	return nil
}

// Compile-time assertion that TetragonAgent satisfies the Agent interface.
var _ Agent = (*TetragonAgent)(nil)
