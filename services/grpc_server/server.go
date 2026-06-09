package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/sathwikshetty33/hydranet/services/database"
	"github.com/sathwikshetty33/hydranet/services/models"
	hydranetv1 "github.com/sathwikshetty33/hydranet/services/proto/hydranet/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

func (s *GrpcServer) ReportIncident(ctx context.Context, req *hydranetv1.IncidentReport) (*hydranetv1.IncidentAck, error) {
	agentID, err := agentIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	agent, err := s.db.GetAgentByID(agentID)
	if err != nil {
		s.logger.Error("agent lookup failed", "err", err, "agent_id", agentID)
		return nil, status.Error(codes.Internal, "agent lookup failed")
	}
	if agent == nil {
		return nil, status.Errorf(codes.Unauthenticated, "agent %d not registered", agentID)
	}

	incident := &models.Incident{
		AgentId:     agent.ID,
		PodName:     req.GetPodName(),
		ServiceName: req.GetServiceName(),
		Message:     req.GetMessage(),
		Severity:    req.GetSeverity(),
	}
	if err := s.db.CreateIncident(incident); err != nil {
		s.logger.Error("create incident failed", "err", err, "agent_id", agent.ID)
		return nil, status.Error(codes.Internal, "create incident failed")
	}

	s.logger.Info("incident recorded",
		"incident_id", incident.ID,
		"agent_id", agent.ID,
		"pod", incident.PodName,
		"service", incident.ServiceName,
		"severity", incident.Severity,
	)
	return &hydranetv1.IncidentAck{IncidentId: uint64(incident.ID)}, nil
}

func agentIDFromMetadata(ctx context.Context) (uint, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("x-agent-id")
	if len(vals) == 0 || vals[0] == "" {
		return 0, status.Error(codes.Unauthenticated, "missing x-agent-id")
	}
	id, err := strconv.ParseUint(vals[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid x-agent-id: %v", err)
	}
	return uint(id), nil
}
