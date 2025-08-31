package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/middlewarr/server/internal/middlewares"
	"github.com/middlewarr/server/internal/proxy"
	"github.com/middlewarr/server/internal/store"
	"github.com/middlewarr/server/internal/templates"
	"github.com/middlewarr/server/internal/tools"
)

func SetupAdminRoutes(r chi.Router, c *store.ConfigurationRepository) {
	r.With(middlewares.WithAPIKey).Route("/admin", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// Misc
			r.Mount("/", newRoutesV1(c))

			// Template
			r.Mount("/template", newTemplatesRoutesV1())

			// Configuration
			r.Mount("/service", newServicesRoutesV1(c))
			r.Mount("/app", newAppsRoutesV1(c))
			r.Mount("/proxy", newProxiesRoutesV1(c))
		})
	})
}

func newRoutesV1(repository *store.ConfigurationRepository) chi.Router {
	r := chi.NewRouter()

	// Reload
	r.Post("/reload", func(w http.ResponseWriter, r *http.Request) {
		templates.LoadTemplates()
		proxy.LoadProxy(repository)
	})

	// Specs
	r.Get("/specs/{type}", getSpecsByType) // GET /specs/{type}

	return r
}

func getSpecsByType(w http.ResponseWriter, r *http.Request) {
	serviceType := chi.URLParam(r, "type")

	spec, err := tools.GetOpenAPISpecs(serviceType)
	if err != nil {
		errorHandler(w, err)
		return
	}

	responseHandler(w, http.StatusOK, spec.Paths)
}
