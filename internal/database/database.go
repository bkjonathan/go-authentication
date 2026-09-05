package database

import (
	"fmt"

	"github.com/bkjonathan/go-authentication/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DSN renders the connection string for a database config. Exported because
// the schema tooling in cmd/schema needs the same connection with a search_path
// appended, and one spelling of the DSN is easier to keep correct than two.
func DSN(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC", cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode)
}

func New(cfg *config.DatabaseConfig) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.Open(DSN(cfg)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Println("Database is connected")
	return db, nil
}
