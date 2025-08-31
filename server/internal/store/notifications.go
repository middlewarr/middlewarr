package store

import (
	"github.com/middlewarr/server/internal/models"
	"gorm.io/gorm"
)

func (c *ConfigurationRepository) CreateNotification(notification *models.Notification) error {
	result := c.db.Create(notification)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (c *ConfigurationRepository) ReadNotifications() (*[]models.Notification, error) {
	var notifications []models.Notification

	err := c.db.Find(&notifications).Error
	if err != nil {
		return nil, err
	}

	return &notifications, nil
}

func (c *ConfigurationRepository) UpdateNotification(id int, notification *models.Notification) error {
	newNotification := &models.Notification{}

	result := c.db.First(&newNotification, id)
	err := result.Error
	if err != nil {
		return err
	}

	// TODO: Validate fields
	newNotification.URL = notification.URL

	result = c.db.Save(&newNotification)
	err = result.Error
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigurationRepository) DestroyNotification(id int) error {
	_, err := gorm.G[models.Notification](c.db).Where("id = ?", id).Delete(c.ctx)
	if err != nil {
		return err
	}

	return nil
}
