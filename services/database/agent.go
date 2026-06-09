package database

import (
	"errors"

	"github.com/sathwikshetty33/hydranet/services/models"
	"gorm.io/gorm"
)

func (db *Database) GetAgentByID(id uint) (*models.Agent, error) {
	var agent models.Agent
	result := db.Connection.First(&agent, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &agent, nil
}
