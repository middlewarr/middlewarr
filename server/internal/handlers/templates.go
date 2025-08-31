package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/middlewarr/server/internal/templates"
)

type templatesHandlerV1 struct {
	//
}

func newTemplatesRoutesV1() chi.Router {
	h := &templatesHandlerV1{}

	r := chi.NewRouter()

	r.Get("/", h.getTemplate)             // GET /template
	r.Get("/{type}", h.getTemplateByType) // GET /template/{type}

	return r
}

func (h templatesHandlerV1) getTemplate(w http.ResponseWriter, r *http.Request) {
	templates, err := templates.ReadTemplates()
	if err != nil {
		errorHandler(w, err)
		return
	}

	responseHandler(w, http.StatusOK, templates)
}

func (h templatesHandlerV1) getTemplateByType(w http.ResponseWriter, r *http.Request) {
	templateType := chi.URLParam(r, "type")

	template, err := templates.ReadTemplate(templateType)
	if err != nil {
		errorHandler(w, err)
		return
	}

	responseHandler(w, http.StatusOK, template)
}
