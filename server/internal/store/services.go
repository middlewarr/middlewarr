package store

import (
	"errors"
	"strings"

	"github.com/middlewarr/server/internal/models"
	"github.com/middlewarr/server/internal/tools"
	"gorm.io/gorm"
)

func (c *ConfigurationRepository) CreateService(service *models.Service) error {
	err := tools.ValidateServiceType(service.Type)
	if err != nil {
		return err
	}

	err = gorm.G[models.Service](c.db).Create(c.ctx, service)
	if err != nil {
		if strings.HasPrefix(err.Error(), errUniqueConstraintFailed.Error()) {
			return errors.New("a service with the same name already exists")
		}

		return err
	}

	return nil
}

func (c *ConfigurationRepository) ReadService(id int) (*models.Service, error) {
	service, err := gorm.G[models.Service](c.db).
		Preload("Proxies", nil).
		Preload("Proxies.App", nil).
		Preload("Proxies.Service", nil).
		Where("id = ?", id).
		First(c.ctx)
	if err != nil {
		return nil, err
	}

	return &service, nil
}

func (c *ConfigurationRepository) ReadServices() (*[]models.Service, error) {
	services, err := gorm.G[models.Service](c.db).
		Preload("Proxies", nil).
		Preload("Proxies.App", nil).
		Preload("Proxies.Service", nil).
		Find(c.ctx)
	if err != nil {
		return nil, err
	}

	return &services, nil
}

func (c *ConfigurationRepository) UpdateService(id int, service *models.Service) error {
	// TODO: Validate fields

	_, err := gorm.G[models.Service](c.db).Where("id = ?", id).Updates(c.ctx, *service)
	if err != nil {
		if strings.HasPrefix(err.Error(), errUniqueConstraintFailed.Error()) {
			return errors.New("a service with the same name already exists")
		}

		return err
	}

	return nil
}

func (c *ConfigurationRepository) DestroyService(id int) error {
	_, err := gorm.G[models.Service](c.db).Where("id = ?", id).Delete(c.ctx)
	if err != nil {
		if strings.HasPrefix(err.Error(), errForeignKeyConstraintFailed.Error()) {
			return errors.New("remove all proxies before deleting the service")
		}

		return err
	}

	return nil
}
