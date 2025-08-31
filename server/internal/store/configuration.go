package store

import (
	"context"

	"github.com/middlewarr/server/internal/models"
	"github.com/middlewarr/server/internal/tools"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ConfigurationRepository struct {
	db  *gorm.DB
	ctx context.Context
}

const (
	configurationDb string = "middlewarr.db?_foreign_keys=on"
)

func NewConfigurationRepository() *ConfigurationRepository {
	ctx := context.Background()

	dbPath := tools.GetDataSubPath(configurationDb)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		panic("failed to connect database")
	}

	if err := db.AutoMigrate(&models.Service{}); err != nil {
		panic("failed to migrate Services")
	}

	if err := db.AutoMigrate(&models.App{}); err != nil {
		panic("failed to migrate Apps")
	}

	if err := db.AutoMigrate(&models.Proxy{}); err != nil {
		panic("failed to migrate Proxies")
	}

	if err := db.AutoMigrate(&models.Notification{}); err != nil {
		panic("failed to migrate Notifications")
	}

	c := &ConfigurationRepository{db, ctx}

	return c
}
