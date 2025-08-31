package tools

import (
	"net/url"
	"path/filepath"
)

const (
	dataDir      string = "data"
	logsDir      string = "logs"
	templatesDir string = "templates"
)

func GetDataPath() string {
	return dataDir
}

func GetDataSubPath(path string) string {
	return filepath.Join(GetDataPath(), path)
}

func GetLogsPath() string {
	return GetDataSubPath(logsDir)
}

func GetTemplatesPath() string {
	l := GetLogger()
	s := GetSettings()

	repositoryURL := s.String("templates.repository")

	u, err := url.Parse(repositoryURL)
	if err != nil {
		l.Fatal().
			Err(err).
			Msg("invalid repository URL")
	}

	return filepath.Join(GetDataSubPath(templatesDir), u.Host, u.Path)
}
