package storage

import (
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
)

import _ "github.com/jinzhu/gorm/dialects/sqlite"

func MigrateDB() {
	/*log.Info().Msg("Migrating DB")
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate()*/
}
