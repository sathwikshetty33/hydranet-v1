package models

type Agent struct {
	BaseModel
	Name          string `gorm:"type:varchar(255);not null" json:"name"`
	ClusterName   string `gorm:"type:varchar(255)" json:"cluster_name"`
	AuthTokenHash string `gorm:"type:varchar(255);not null" json:"-"`
}
