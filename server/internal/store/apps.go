package store

import (
	"errors"
	"strings"

	"github.com/middlewarr/server/internal/models"
	"gorm.io/gorm"
)

func (c *ConfigurationRepository) CreateApp(app *models.App) error {
	// TODO: Validate app template

	err := gorm.G[models.App](c.db).Create(c.ctx, app)
	if err != nil {
		if strings.HasPrefix(err.Error(), errUniqueConstraintFailed.Error()) {
			return errors.New("an app with the same name already exists")
		}

		return err
	}

	return nil
}

func (c *ConfigurationRepository) ReadApp(id int) (*models.App, error) {
	app, err := gorm.G[models.App](c.db).
		Preload("Proxies", nil).
		Preload("Proxies.App", nil).
		Preload("Proxies.Service", nil).
		Where("id = ?", id).
		First(c.ctx)
	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (c *ConfigurationRepository) ReadApps() (*[]models.App, error) {
	apps, err := gorm.G[models.App](c.db).
		Preload("Proxies", nil).
		Preload("Proxies.App", nil).
		Preload("Proxies.Service", nil).
		Find(c.ctx)
	if err != nil {
		return nil, err
	}

	return &apps, nil
}

func (c *ConfigurationRepository) UpdateApp(id int, app *models.App) error {
	// TODO: Validate fields

	_, err := gorm.G[models.App](c.db).Where("id = ?", id).Updates(c.ctx, *app)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigurationRepository) DestroyApp(id int) error {
	_, err := gorm.G[models.App](c.db).Where("id = ?", id).Delete(c.ctx)
	if err != nil {
		if strings.HasPrefix(err.Error(), errForeignKeyConstraintFailed.Error()) {
			return errors.New("remove all proxies before deleting the app")
		}

		return err
	}

	return nil
}
