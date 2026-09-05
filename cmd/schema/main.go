// Command schema rebuilds the shadow schema: a scratch Postgres schema holding
// exactly the tables the GORM models describe.
//
// Atlas writes a migration by diffing two states, and one of them has to be
// "whatever the models say right now". Materialising that with AutoMigrate,
// rather than describing the schema a second time in HCL, keeps the models the
// only definition of the schema in this project.
//
//	go run ./cmd/schema
//	go run ./cmd/schema -schema other_name
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/bkjonathan/go-authentication/cmd/schema/models"
	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/bkjonathan/go-authentication/internal/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// A schema name cannot be a bind parameter, so it is checked rather than
// escaped: nothing but a plain identifier reaches the DROP below.
var plainIdentifier = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

func main() {
	name := flag.String("schema", "atlas_shadow", "scratch schema to rebuild")
	flag.Parse()

	if err := rebuild(*name); err != nil {
		fmt.Fprintln(os.Stderr, "schema:", err)
		os.Exit(1)
	}
}

func rebuild(name string) error {
	if !plainIdentifier.MatchString(name) {
		return fmt.Errorf("%q is not a plain schema name", name)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := gorm.Open(
		postgres.Open(database.DSN(&cfg.Database)+" search_path="+name),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}

	// Dropped and recreated rather than migrated in place: AutoMigrate never
	// drops a column, so a shadow that is only added to would slowly stop
	// describing the models.
	if err := db.Exec("DROP SCHEMA IF EXISTS " + name + " CASCADE").Error; err != nil {
		return fmt.Errorf("drop schema: %w", err)
	}
	if err := db.Exec("CREATE SCHEMA " + name).Error; err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	if err := models.Migrate(db); err != nil {
		return fmt.Errorf("migrate models: %w", err)
	}

	fmt.Printf("shadow schema %q rebuilt from %d models\n", name, len(models.All()))

	return nil
}
