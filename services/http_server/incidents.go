package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


func (s *HttpServer) listIncidentsHandler(c *gin.Context) {
	incidents, err := s.db.ListIncidents()
	if err != nil {
		s.logger.Error("list incidents failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list incidents"})
		return
	}
	c.JSON(http.StatusOK, incidents)
}
