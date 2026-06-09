package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sathwikshetty33/hydranet/services/models"
	"github.com/sathwikshetty33/hydranet/services/utils"
)

type createAgentRequest struct {
	Name        string `json:"name" binding:"required"`
	ClusterName string `json:"cluster_name"`
}

// createAgentResponse flattens the new Agent's public fields and appends the
// raw auth token, which is returned exactly once.
type createAgentResponse struct {
	*models.Agent
	AuthToken string `json:"auth_token"`
}

func (s *HttpServer) createAgentHandler(c *gin.Context) {
	var req createAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateAgentToken()
	if err != nil {
		s.logger.Error("generate agent token failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	hash, err := utils.HashToken(token)
	if err != nil {
		s.logger.Error("hash agent token failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash token"})
		return
	}

	agent := &models.Agent{
		Name:          req.Name,
		ClusterName:   req.ClusterName,
		AuthTokenHash: hash,
	}
	if err := s.db.CreateAgent(agent); err != nil {
		s.logger.Error("create agent failed", "err", err, "name", req.Name)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create agent"})
		return
	}

	s.logger.Info("agent created", "agent_id", agent.ID, "name", agent.Name)
	c.JSON(http.StatusCreated, createAgentResponse{
		Agent:     agent,
		AuthToken: token,
	})
}

func (s *HttpServer) listAgentsHandler(c *gin.Context) {
	agents, err := s.db.ListAgents()
	if err != nil {
		s.logger.Error("list agents failed", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list agents"})
		return
	}
	c.JSON(http.StatusOK, agents)
}
