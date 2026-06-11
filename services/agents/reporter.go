package agents

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	hydranetv1 "github.com/sathwikshetty33/hydranet/services/proto/hydranet/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Reporter is the sink-side contract every agent uses to forward incidents
// to the core. The transport is whatever the implementation chooses — gRPC,
// stdout, an in-memory queue for tests — so agents stay decoupled from how
// "the core" is reached.
type Reporter interface {
	// Report forwards an incident and returns the core's decision back so
	// the caller can react (e.g., quarantine the pod). Returns nil ack on
	// transport error.
	Report(ctx context.Context, incident IncidentDetails) (*ReportAck, error)
	io.Closer
}

// ReportAck is the transport-agnostic shape of the core's response to a
// Report() call. The gRPC impl maps IncidentAck → ReportAck; future
// transports can do the same translation.
type ReportAck struct {
	IncidentID uint64
	// Quarantine is the core's signal to the agent that this pod should
	// be isolated via NetworkPolicy. The agent's Isolator (if configured)
	// is the actuator.
	Quarantine bool
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
	// Scheme drives transport selection:
	//   "https://host[:port]"  /  "grpcs://host[:port]"  → TLS, port defaults to 443
	//   "http://host[:port]"   /  "grpc://host[:port]"   → plaintext, port defaults to 80
	//   "host:port"            (no scheme)               → plaintext (backwards compat)
	// Examples:
	//   "https://hydranet.example.com:443"     (ingress / ngrok)
	//   "host.minikube.internal:50051"          (in-cluster, plaintext)
	//   "localhost:50051"                       (local dev)
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

	target, useTLS, err := resolveCoreTarget(cfg.CoreAddr)
	if err != nil {
		return nil, err
	}

	var transportCreds credentials.TransportCredentials
	if useTLS {
		// System root CAs — fine for public certs (Let's Encrypt, ngrok).
		// For a private CA, plug in a RootCAs pool here.
		transportCreds = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	} else {
		transportCreds = insecure.NewCredentials()
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		return nil, fmt.Errorf("dial core %s: %w", target, err)
	}
	return &CoreReporter{
		conn:    conn,
		client:  hydranetv1.NewIncidentServiceClient(conn),
		agentID: cfg.AgentID,
		token:   cfg.AuthToken,
	}, nil
}

// resolveCoreTarget turns a user-facing address into a gRPC dial target plus
// a TLS decision. gRPC's NewClient expects "host:port" (or its own scheme
// like dns:///, unix:///), so https:// / http:// have to be stripped before
// being passed through.
func resolveCoreTarget(addr string) (target string, useTLS bool, err error) {
	if !strings.Contains(addr, "://") {
		// Bare "host:port" stays bare → plaintext.
		return addr, false, nil
	}

	u, err := url.Parse(addr)
	if err != nil {
		return "", false, fmt.Errorf("parse core address: %w", err)
	}

	host := u.Hostname()
	port := u.Port()
	if host == "" {
		return "", false, fmt.Errorf("core address %q missing host", addr)
	}

	switch strings.ToLower(u.Scheme) {
	case "https", "grpcs":
		if port == "" {
			port = "443"
		}
		return host + ":" + port, true, nil
	case "http", "grpc":
		if port == "" {
			port = "80"
		}
		return host + ":" + port, false, nil
	default:
		return "", false, fmt.Errorf("unsupported scheme %q in core address (use http, https, grpc, grpcs, or bare host:port)", u.Scheme)
	}
}

func (r *CoreReporter) Close() error {
	if r.conn == nil {
		return nil
	}
	return r.conn.Close()
}

func (r *CoreReporter) Report(ctx context.Context, incident IncidentDetails) (*ReportAck, error) {
	md := metadata.Pairs(
		"authorization", "Bearer "+r.token,
		"x-agent-id", r.agentID,
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	ack, err := r.client.ReportIncident(ctx, &hydranetv1.IncidentReport{
		PodName:     incident.PodName,
		ServiceName: incident.ServiceName,
		Message:     incident.Message,
		Severity:    incident.Severity,
	})
	if err != nil {
		return nil, fmt.Errorf("report incident to core: %w", err)
	}
	return &ReportAck{
		IncidentID: ack.GetIncidentId(),
		Quarantine: ack.GetQuarantine(),
	}, nil
}

// Compile-time assertion that CoreReporter satisfies the Reporter interface.
var _ Reporter = (*CoreReporter)(nil)
