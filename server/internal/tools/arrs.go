package tools

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/middlewarr/server/internal/models"
)

type ServiceType string

const (
	Lidarr   ServiceType = "lidarr"
	Prowlarr ServiceType = "prowlarr"
	Radarr   ServiceType = "radarr"
	Sonarr   ServiceType = "sonarr"
)

type ServiceSystemStatusResponse struct {
	AppName string `json:"appName"`
	Version string `json:"version"`
}

type ServiceOpenAPISpec struct {
	Paths map[string]map[string]any `json:"paths"` // method, operation
}

func ValidateServiceType(serviceType string) error {
	s := ServiceType(serviceType)

	switch s {
	case Lidarr:
	case Prowlarr:
	case Radarr:
	case Sonarr:
		return nil
	default:
		return errors.New("invalid service type")
	}

	return nil
}

func getServiceHealthPath(serviceType string) string {
	s := ServiceType(serviceType)

	switch s {
	case "lidarr", "prowlarr":
		return "/api/v1/system/status"
	case "radarr", "sonarr":
		return "/api/v3/system/status"
	}

	return ""
}

func ValidateServiceHealth(service models.Service) error {
	l := GetLogger()
	client := http.Client{}

	healthCheckUrl := service.URL + getServiceHealthPath(service.Type)

	req, err := http.NewRequest("GET", healthCheckUrl, nil)
	if err != nil {
		l.Error().
			Err(err).
			Str("service_name", service.Name).
			Str("service_type", service.Type).
			Str("service_url", service.URL).
			Msg("cannot contact the service")

		return err
	}

	req.Header = http.Header{
		"X-Api-Key": {service.APIKey},
	}

	res, err := client.Do(req)
	if err != nil {
		l.Warn().
			Str("service_name", service.Name).
			Str("service_type", service.Type).
			Str("service_url", service.URL).
			Msg("Cannot contact the service")
		return err
	}

	if res.StatusCode == http.StatusUnauthorized {
		l.Warn().
			Str("service_name", service.Name).
			Str("service_type", service.Type).
			Str("service_url", service.URL).
			Msg("Invalid service API key")
		return err
	}

	if res.StatusCode == http.StatusOK {
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			l.Warn().
				Str("service_name", service.Name).
				Str("service_type", service.Type).
				Str("service_url", service.URL).
				Msg("Cannot ReadAll")
			return err
		}

		var result ServiceSystemStatusResponse
		err = json.Unmarshal(body, &result)
		if err != nil {
			l.Warn().
				Str("service_name", service.Name).
				Str("service_type", service.Type).
				Str("service_url", service.URL).
				Msg("Cannot Unmarshal")
			return err
		}

		if result.Version == "" {
			l.Warn().
				Str("service_name", service.Name).
				Str("service_type", service.Type).
				Str("service_url", service.URL).
				Msg("Cannot retrieve service version")
			return err
		}

		l.Info().
			Str("service_name", service.Name).
			Str("service_type", service.Type).
			Str("AppName", result.AppName).
			Str("Version", result.Version).
			Str("service_url", service.URL).
			Msg("Service healty")
	} else {
		l.Warn().
			Str("service_name", service.Name).
			Str("service_type", service.Type).
			Str("service_url", service.URL).
			Msg("Service not healty")
	}

	return nil
}

func getOpenAPISpecsURL(serviceType string) string {
	s := ServiceType(serviceType)

	switch s {
	case Lidarr:
		return "https://raw.githubusercontent.com/lidarr/Lidarr/develop/src/Lidarr.Api.V1/openapi.json"
	case Prowlarr:
		return "https://raw.githubusercontent.com/Prowlarr/Prowlarr/develop/src/Prowlarr.Api.V1/openapi.json"
	case Radarr:
		return "https://raw.githubusercontent.com/Radarr/Radarr/develop/src/Radarr.Api.V3/openapi.json"
	case Sonarr:
		return "https://raw.githubusercontent.com/Sonarr/Sonarr/develop/src/Sonarr.Api.V3/openapi.json"
	default:
		return ""
	}
}

func GetOpenAPISpecs(serviceType string) (*ServiceOpenAPISpec, error) {
	l := GetLogger()

	specsURL := getOpenAPISpecsURL(serviceType)
	if specsURL == "" {
		l.Error().
			Msg("Invalid service OpenAPI specs URL")

		return nil, errors.New("invalid service OpenAPI specs URL")
	}

	res, err := http.Get(specsURL)
	if err != nil {
		l.Error().
			Err(err).
			Msg("cannot get OpenAPI specs")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		l.Error().
			Err(err).
			Msg("cannot read OpenAPI specs body")
		return nil, err
	}

	var result *ServiceOpenAPISpec
	err = json.Unmarshal(body, &result)
	if err != nil {
		l.Error().
			Err(err).
			Msg("cannot unmarshal OpenAPI specs JSON")
		return nil, err
	}

	return result, nil
}
