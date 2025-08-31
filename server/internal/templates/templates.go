package templates

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/middlewarr/server/internal/tools"
)

type TemplateEndpoints map[string]map[string][]string

type Template struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	URL       string            `json:"url"`
	Endpoints TemplateEndpoints `json:"endpoints"`
}

type TemplateFiles struct {
	templates   []Template
	templateIDs []string
}

var templateFiles atomic.Value

func getTemplateFiles() *TemplateFiles {
	return templateFiles.Load().(*TemplateFiles)
}

func setTemplateFiles(pr *TemplateFiles) {
	templateFiles.Store(pr)
}

func LoadTemplates() {
	l := tools.GetLogger()

	l.Info().
		Msg("Loading template files...")

	templatesFiles, err := initTemplateFiles()
	if err != nil {
		l.Panic().
			Err(err).
			Msg("failed to initialize templates")
	}

	setTemplateFiles(templatesFiles)

	l.Info().
		Msg("Template files loaded")
}

func initTemplateFiles() (*TemplateFiles, error) {
	templatesPath := tools.GetTemplatesPath()

	entries, err := os.ReadDir(templatesPath)
	if err != nil {
		return nil, err
	}

	var templateIDs []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		entryExt := filepath.Ext(entry.Name())

		if entryExt != ".json" {
			continue
		}

		templateIDs = append(templateIDs, strings.TrimSuffix(entry.Name(), entryExt))
	}

	var templates []Template

	for _, file := range templateIDs {
		f, err := os.Open(filepath.Join(templatesPath, file+".json"))
		if err != nil {
			return nil, err
		}
		defer f.Close()

		var template Template

		dec := json.NewDecoder(f)
		if err := dec.Decode(&template); err != nil {
			return nil, err
		}

		err = validateTemplateFile(file, template)
		if err != nil {
			continue
		}

		templates = append(templates, template)
	}

	return &TemplateFiles{
		templates,
		templateIDs,
	}, nil

}

func ReadTemplates() (*[]Template, error) {
	t := getTemplateFiles()

	return &t.templates, nil
}

func ReadTemplate(id string) (*Template, error) {
	t := getTemplateFiles()

	for _, template := range t.templates {
		if template.ID == id {
			return &template, nil
		}
	}

	return nil, errors.New("template not found")
}

func validateTemplateFile(file string, template Template) error {
	l := tools.GetLogger()

	if len(template.ID) == 0 {
		l.Warn().
			Str("template_file", file+".json").
			Msg("Missing template ID")
		return errors.New("missing template ID")
	}

	if template.ID != file {
		l.Warn().
			Str("template_file", file+".json").
			Msg("Invalid template ID")
		return errors.New("invalid template ID")
	}

	if len(template.Name) == 0 {
		l.Warn().
			Str("template_file", file+".json").
			Msg("Missing template Name")
		return errors.New("missing template Name")
	}

	if len(template.Endpoints) == 0 {
		l.Warn().
			Str("template_file", file+".json").
			Msg("Missing template Endpoints")
		return errors.New("missing template Endpoints")
	}

	templateEndpoints := template.Endpoints

	for serviceType, endpoints := range templateEndpoints {
		specs, err := tools.GetOpenAPISpecs(serviceType)
		if err != nil {
			l.Warn().
				Str("template_id", template.ID).
				Str("template_name", template.Name).
				Str("service_type", serviceType).
				Msg("Invalid service type")
			return errors.New("invalid service type")
		}

		for path, methods := range endpoints {
			specsMethods, ok := specs.Paths[path]
			if !ok {
				l.Warn().
					Str("path", path).
					Str("template_id", template.ID).
					Str("template_name", template.Name).
					Str("service_type", serviceType).
					Msg("Invalid path in route")
			} else {
				// Check methods only if path is valid.
				for _, configMethod := range methods {
					if _, ok := specsMethods[strings.ToLower(configMethod)]; !ok {
						l.Warn().
							Str("method", configMethod).
							Str("path", path).
							Str("template_id", template.ID).
							Str("template_name", template.Name).
							Str("service_type", serviceType).
							Msg("Invalid method path in route")
					}
				}
			}
		}
	}

	return nil
}
