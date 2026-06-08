package agents

import (
	"context"
	"errors"
	"fmt"
	"io"

	hydranetv1 "github.com/sathwikshetty33/hydranet/services/proto/hydranet/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Reporter is the sink-side contract every agent uses to forward incidents
// to the core. The transport is whatever the implementation chooses — gRPC,
// stdout, an in-memory queue for tests — so agents stay decoupled from how
// "the core" is reached.
type Reporter interface {
	Report(ctx context.Context, incident IncidentDetails) error
	io.Closer
}

// CoreReporter is the production Reporter: it speaks the IncidentService
// gRPC to the HydraNet core. Auth and identity travel in gRPC metadata so
// the wire payload stays focused on the incident itself.
type CoreReporter struct {
	conn    *grpc.ClientConn
	client  hydranetv1.IncidentServiceClient
	agentID string
	token   string
}

type CoreReporterConfig struct {
	// CoreAddr is the gRPC target of the HydraNet core IncidentService.
	//   "hydranet-core.hydranet.svc:50051"  /  "localhost:50051"
	CoreAddr string

	// AgentID identifies this agent to the core.
	AgentID string

	// AuthToken is the raw bearer token; the core hashes and compares.
	AuthToken string
}

func NewCoreReporter(cfg CoreReporterConfig) (*CoreReporter, error) {
	if cfg.CoreAddr == "" {
		return nil, errors.New("core address required")
	}
	if cfg.AgentID == "" {
		return nil, errors.New("agent id required")
	}
	conn, err := grpc.NewClient(
		cfg.CoreAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial core %s: %w", cfg.CoreAddr, err)
	}
	return &CoreReporter{
		conn:    conn,
		client:  hydranetv1.NewIncidentServiceClient(conn),
		agentID: cfg.AgentID,
		token:   cfg.AuthToken,
	}, nil
}

func (r *CoreReporter) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *CoreReporter) Report(ctx context.Context, incident IncidentDetails) error {
	md := metadata.Pairs(
		"authorization", "Bearer "+r.token,
		"x-agent-id", r.agentID,
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	if _, err := r.client.ReportIncident(ctx, &hydranetv1.IncidentReport{
		PodName:     incident.PodName,
		ServiceName: incident.ServiceName,
		Message:     incident.Message,
		Severity:    incident.Severity,
	}); err != nil {
		return fmt.Errorf("report incident to core: %w", err)
	}
	return nil
}

// Compile-time assertion that CoreReporter satisfies the Reporter interface.
var _ Reporter = (*CoreReporter)(nil)
