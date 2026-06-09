package httpserver

import (
	"log/slog"
	"errors"
	"net/http"
	"time"
	"context"
	"github.com/gin-gonic/gin"

	"github.com/sathwikshetty33/hydranet/services/database"
)


type HttpServer struct {
	Router *gin.Engine
	db     *database.Database
	logger *slog.Logger
}

func NewHttpServer(db *database.Database, logger *slog.Logger) *HttpServer {
	s := &HttpServer{
		Router: gin.Default(),
		db:     db,
		logger: logger,
	}
	s.registerRoutes()
	return s
}

func (s *HttpServer) registerRoutes() {
	v1 := s.Router.Group("/api/v1")
	{
		v1.POST("/agents", s.createAgentHandler)
		v1.GET("/agents", s.listAgentsHandler)
	}
}


func (s *HttpServer) Serve(ctx context.Context, addr string) error {
    srv := &http.Server{
        Addr:    addr,
        Handler: s.Router,
    }

    serveErr := make(chan error, 1)
    go func() {
        s.logger.Info("http server listening", "addr", addr)
        err := srv.ListenAndServe()
        if errors.Is(err, http.ErrServerClosed) {
            err = nil
        }
        serveErr <- err
    }()

    select {
    case <-ctx.Done():
        s.logger.Info("http server stopping", "reason", ctx.Err())
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        return srv.Shutdown(shutdownCtx)
    case err := <-serveErr:
        return err
    }
}