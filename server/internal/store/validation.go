package store

import (
	"errors"

	"github.com/middlewarr/server/internal/models"
	"github.com/middlewarr/server/internal/tools"
)

var errForeignKeyConstraintFailed = errors.New("FOREIGN KEY constraint failed")
var errUniqueConstraintFailed = errors.New("UNIQUE constraint failed")

func ValidateConfig(c *ConfigurationRepository) {
	// TODO: Handle errors
	services, _ := c.ReadServices()
	apps, _ := c.ReadApps()
	proxies, _ := c.ReadProxies()

	validateServices(*services)
	validateApps(*apps)
	validateProxies(*proxies)
}

func validateServices(services []models.Service) {
	l := tools.GetLogger()

	for _, service := range services {
		err := tools.ValidateServiceType(service.Type)
		if err != nil {
			l.Warn().
				Str("service", service.Name).
				Msg("Invalid service type")

			continue
		}

		err = tools.ValidateServiceHealth(service)
		if err != nil {
			l.Warn().
				Str("service_id", service.Name).
				Str("type", service.Type).
				Msg("Service not healty")
		}
	}
}

func validateApps(apps []models.App) {
	l := tools.GetLogger()

	for _, app := range apps {
		if !*app.IsActive {
			l.Warn().
				Str("app", app.Name).
				Str("app", app.Template).
				Int("proxies", len(app.Proxies)).
				Msg("App not active")
		}
	}
}

func validateProxies(proxies []models.Proxy) {
	// TODO
}
