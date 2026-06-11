package database

import (
	"time"

	"github.com/sathwikshetty33/hydranet/services/models"
)

func (db *Database) CreateIncident(incident *models.Incident) error {
	return db.Connection.Create(incident).Error
}

// CountRecentHighSeverityIncidents returns how many incidents at severity
// "high" or "critical" have been recorded for podName within the last
// `window`. Used by the quarantine decision: a pod accumulating high-tier
// activity in a short window is treated as compromised even if no single
// event was critical on its own.
func (db *Database) CountRecentHighSeverityIncidents(podName string, window time.Duration) (int64, error) {
	var count int64
	err := db.Connection.
		Table("incidents").
		Where("pod_name = ?", podName).
		Where("severity IN ?", []string{"high", "critical"}).
		Where("created_at > ?", time.Now().Add(-window)).
		Where("deleted_at IS NULL").
		Count(&count).Error
	return count, err
}


type IncidentListItem struct {
	ID          uint      `json:"id"`
	AgentID     uint      `json:"agent_id"`
	AgentName   string    `json:"agent_name"`
	ClusterName string    `json:"cluster_name"`
	PodName     string    `json:"pod_name"`
	ServiceName string    `json:"service_name"`
	Message     string    `json:"message"`
	Severity    string    `json:"severity"`
	CreatedAt   time.Time `json:"created_at"`
}


func (db *Database) ListIncidents() ([]IncidentListItem, error) {
	var rows []IncidentListItem
	err := db.Connection.
		Table("incidents").
		Select(`incidents.id,
		        incidents.agent_id,
		        agents.name AS agent_name,
		        agents.cluster_name,
		        incidents.pod_name,
		        incidents.service_name,
		        incidents.message,
		        incidents.severity,
		        incidents.created_at`).
		Joins("LEFT JOIN agents ON agents.id = incidents.agent_id").
		Where("incidents.deleted_at IS NULL").
		Order("incidents.id DESC").
		Limit(200).
		Scan(&rows).Error
	return rows, err
}
