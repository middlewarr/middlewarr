package store

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/middlewarr/server/internal/models"
	"gorm.io/gorm"
)

func generateApiKey() string {
	apiKey := strings.ReplaceAll(uuid.New().String(), "-", "")

	return apiKey
}

func (c *ConfigurationRepository) CreateProxy(proxy *models.Proxy) error {
	if proxy.APIKey == "" {
		proxy.APIKey = generateApiKey()
	}

	err := gorm.G[models.Proxy](c.db).Create(c.ctx, proxy)
	if err != nil {
		if strings.HasPrefix(err.Error(), errUniqueConstraintFailed.Error()) {
			return errors.New("a proxy for the selected app and service already exists")
		}

		return err
	}

	return nil
}

func (c *ConfigurationRepository) ReadProxy(id int) (*models.Proxy, error) {
	proxy, err := gorm.G[models.Proxy](c.db).
		Preload("App", nil).
		Preload("Service", nil).
		Where("id = ?", id).
		First(c.ctx)
	if err != nil {
		return nil, err
	}

	return &proxy, nil
}

func (c *ConfigurationRepository) ReadProxies() (*[]models.Proxy, error) {
	proxies, err := gorm.G[models.Proxy](c.db).
		Preload("App", nil).
		Preload("Service", nil).
		Find(c.ctx)
	if err != nil {
		return nil, err
	}

	return &proxies, nil
}

func (c *ConfigurationRepository) UpdateProxy(id int, proxy *models.Proxy) error {
	_, err := gorm.G[models.Proxy](c.db).Where("id = ?", id).Updates(c.ctx, *proxy)
	if err != nil {
		if strings.HasPrefix(err.Error(), errUniqueConstraintFailed.Error()) {
			return errors.New("a proxy for the selected app and service already exists")
		}

		return err
	}

	return nil
}

func (c *ConfigurationRepository) DestroyProxy(id int) error {
	_, err := gorm.G[models.Proxy](c.db).Where("id = ?", id).Delete(c.ctx)
	if err != nil {
		return err
	}

	return nil
}
