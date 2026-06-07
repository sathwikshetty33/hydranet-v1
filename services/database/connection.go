package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Connection *gorm.DB
}

func NewDatabase(dbHost, dbUser, dbPassword, dbName, dbPort string) (*Database, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost,
		dbUser,
		dbPassword,
		dbName,
		dbPort,
	)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("db connection failed: %w", err)
	}

	if err := conn.AutoMigrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &Database{Connection: conn}, nil
}
