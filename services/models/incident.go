package models

type Incident struct {
	BaseModel
	AgentId     uint   `gorm:"not null" json:"agent_id"`
	Namespace   string `gorm:"type:varchar(255);not null;default:''" json:"namespace"`
	PodName     string `gorm:"type:varchar(255);not null" json:"pod_name"`
	ServiceName string `gorm:"type:varchar(255);not null" json:"service_name"`
	Message     string `gorm:"type:varchar(255);not null" json:"message"`
	Severity    string `gorm:"type:varchar(50);not null" json:"severity"`
}
