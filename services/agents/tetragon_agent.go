package agents

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	tetragon "github.com/cilium/tetragon/api/v1/tetragon"
	hydranetv1 "github.com/sathwikshetty33/hydranet/services/proto/hydranet/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type TetragonAgentConfig struct {
	// TetragonAddr is the gRPC target for the FineGuidanceSensors service.
	// Examples:
	//   "unix:///var/run/tetragon/tetragon.sock"  (DaemonSet co-mount)
	//   "127.0.0.1:54321"                         (port-forwarded for dev)
	TetragonAddr string

	// CoreAddr is the gRPC target for the HydraNet core IncidentService.
	//   "hydranet-core.hydranet.svc:50051"  /  "localhost:50051"
	CoreAddr string

	// AgentID identifies this agent to the core (agents.id PK).
	AgentID string

	// AuthToken is the raw bearer token; the core hashes and compares.
	AuthToken string

	// SuspiciousPatterns are substrings matched against the exec'd
	// binary + arguments. Defaults to a sensible set if left empty.
	SuspiciousPatterns []string
}

type TetragonAgent struct {
	cfg            TetragonAgentConfig
	logger         *slog.Logger
	coreConn       *grpc.ClientConn
	incidentClient hydranetv1.IncidentServiceClient
}

func NewTetragonAgent(cfg TetragonAgentConfig, logger *slog.Logger) (*TetragonAgent, error) {
	if cfg.TetragonAddr == "" {
		return nil, errors.New("tetragon address required")
	}
	if cfg.CoreAddr == "" {
		return nil, errors.New("core address required")
	}
	if cfg.AgentID == "" {
		return nil, errors.New("agent id required")
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

	conn, err := grpc.NewClient(
		cfg.CoreAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial core %s: %w", cfg.CoreAddr, err)
	}

	return &TetragonAgent{
		cfg:            cfg,
		logger:         logger,
		coreConn:       conn,
		incidentClient: hydranetv1.NewIncidentServiceClient(conn),
	}, nil
}

// Close releases the gRPC connection to the core.
func (a *TetragonAgent) Close() error {
	if a.coreConn == nil {
		return nil
	}
	return a.coreConn.Close()
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

// ReportIncident sends a single IncidentReport to the core's IncidentService.
// Auth and identity travel as gRPC metadata so the wire payload stays focused
// on the incident itself.
func (a *TetragonAgent) ReportIncident(ctx context.Context, incident IncidentDetails) error {
	md := metadata.Pairs(
		"authorization", "Bearer "+a.cfg.AuthToken,
		"x-agent-id", a.cfg.AgentID,
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	ack, err := a.incidentClient.ReportIncident(ctx, &hydranetv1.IncidentReport{
		PodName:     incident.PodName,
		ServiceName: incident.ServiceName,
		Message:     incident.Message,
		Severity:    incident.Severity,
	})
	if err != nil {
		return fmt.Errorf("report incident to core: %w", err)
	}
	a.logger.Debug("incident reported",
		"pod", incident.PodName,
		"incident_id", ack.GetIncidentId())
	return nil
}

// Compile-time assertion that TetragonAgent satisfies the Agent interface.
var _ Agent = (*TetragonAgent)(nil)
