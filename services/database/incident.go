package database

import (

	"github.com/sathwikshetty33/hydranet/services/models"
)

func (db *Database) CreateIncident(incident *models.Incident) error {
	return db.Connection.Create(incident).Error
}
