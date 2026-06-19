package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sathwikshetty33/hydranet/services/database"
	"github.com/sathwikshetty33/hydranet/services/models"
	hydranetv1 "github.com/sathwikshetty33/hydranet/services/proto/hydranet/v1"
	"github.com/sathwikshetty33/hydranet/services/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Quarantine policy knobs. Deliberately constants for v1; promote to config
// when there's a real need to tune per-deployment.
const (
	quarantineWindow    = 5 * time.Minute
	quarantineThreshold = 3 // N high-or-critical incidents in window → quarantine
)

type GrpcServer struct {
	hydranetv1.UnimplementedIncidentServiceServer

	db     *database.Database
	logger *slog.Logger
	grpc   *grpc.Server
}

func NewGrpcServer(db *database.Database, logger *slog.Logger) *GrpcServer {
	s := &GrpcServer{db: db, logger: logger}
	g := grpc.NewServer()
	hydranetv1.RegisterIncidentServiceServer(g, s)
	s.grpc = g
	return s
}

func (s *GrpcServer) Serve(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	serveErr := make(chan error, 1)
	go func() {
		s.logger.Info("grpc server listening", "addr", addr)
		serveErr <- s.grpc.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("grpc server stopping", "reason", ctx.Err())
		s.grpc.GracefulStop()
		return nil
	case err := <-serveErr:
		return err
	}
}


var errInvalidCreds = status.Error(codes.Unauthenticated, "invalid agent credentials")

func (s *GrpcServer) ReportIncident(ctx context.Context, req *hydranetv1.IncidentReport) (*hydranetv1.IncidentAck, error) {
	agentID, token, err := agentCredsFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	agent, err := s.db.GetAgentByID(agentID)
	if err != nil {
		s.logger.Error("agent lookup failed", "err", err, "agent_id", agentID)
		return nil, status.Error(codes.Internal, "agent lookup failed")
	}
	if agent == nil {
		s.logger.Warn("auth rejected: agent not registered", "agent_id", agentID)
		return nil, errInvalidCreds
	}

	if !utils.VerifyToken(token, agent.AuthTokenHash) {
		s.logger.Warn("auth rejected: token mismatch", "agent_id", agent.ID)
		return nil, errInvalidCreds
	}

	incident := &models.Incident{
		AgentId:     agent.ID,
		Namespace:   req.GetNamespace(),
		PodName:     req.GetPodName(),
		ServiceName: req.GetServiceName(),
		Message:     req.GetMessage(),
		Severity:    req.GetSeverity(),
	}
	if err := s.db.CreateIncident(incident); err != nil {
		s.logger.Error("create incident failed", "err", err, "agent_id", agent.ID)
		return nil, status.Error(codes.Internal, "create incident failed")
	}

	quarantine := s.decideQuarantine(incident.Namespace, incident.PodName, incident.Severity)

	s.logger.Info("incident recorded",
		"incident_id", incident.ID,
		"agent_id", agent.ID,
		"pod", incident.PodName,
		"service", incident.ServiceName,
		"severity", incident.Severity,
		"quarantine", quarantine,
	)
	return &hydranetv1.IncidentAck{
		IncidentId: uint64(incident.ID),
		Quarantine: quarantine,
	}, nil
}

// decideQuarantine is the core's policy brain for "should the reporting
// agent isolate this pod?". v1 ruleset:
//
//   - A "critical" severity event is sufficient on its own (escape patterns,
//     credential reads — the low-FP set).
//   - Otherwise, accumulate: ≥quarantineThreshold high-or-critical events
//     in the last quarantineWindow trips quarantine.
//
// Called AFTER the current incident is persisted so the count includes it.
// Fail-closed: DB errors return false (don't isolate) and are logged.
func (s *GrpcServer) decideQuarantine(namespace, podName, severity string) bool {
	if severity == "critical" {
		return true
	}
	count, err := s.db.CountRecentHighSeverityIncidents(namespace, podName, quarantineWindow)
	if err != nil {
		s.logger.Error("quarantine history query failed",
			"err", err, "namespace", namespace, "pod", podName)
		return false
	}
	return count >= quarantineThreshold
}

func agentCredsFromMetadata(ctx context.Context) (uint, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, "", errInvalidCreds
	}

	idVals := md.Get("x-agent-id")
	if len(idVals) == 0 || idVals[0] == "" {
		return 0, "", errInvalidCreds
	}
	id, err := strconv.ParseUint(idVals[0], 10, 64)
	if err != nil {
		return 0, "", errInvalidCreds
	}

	authVals := md.Get("authorization")
	if len(authVals) == 0 {
		return 0, "", errInvalidCreds
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authVals[0], prefix) {
		return 0, "", errInvalidCreds
	}
	token := strings.TrimPrefix(authVals[0], prefix)
	if token == "" {
		return 0, "", errInvalidCreds
	}

	return uint(id), token, nil
}
