package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/middlewarr/server/internal/middlewares"
	"github.com/middlewarr/server/internal/models"
	"github.com/middlewarr/server/internal/proxy"
	"github.com/middlewarr/server/internal/store"
)

type servicesHandlerV1 struct {
	repository *store.ConfigurationRepository
}

func newServicesRoutesV1(repository *store.ConfigurationRepository) chi.Router {
	h := &servicesHandlerV1{repository}

	r := chi.NewRouter()

	r.Get("/", h.getService)   // GET /service
	r.Post("/", h.postService) // POST /service

	r.With(middlewares.WithID).Group(func(r chi.Router) {
		r.Route("/{id:[0-9]+}", func(r chi.Router) {
			r.Get("/", h.getServiceById)       // GET /service/{id}
			r.Put("/", h.putServiceById)       // PUT /service/{id}
			r.Delete("/", h.deleteServiceById) // DELETE /service/{id}
		})
	})

	return r
}

func (h servicesHandlerV1) getService(w http.ResponseWriter, r *http.Request) {
	services, err := h.repository.ReadServices()
	if err != nil {
		errorHandler(w, err)
		return
	}

	responseHandler(w, http.StatusOK, services)
}

func (h servicesHandlerV1) postService(w http.ResponseWriter, r *http.Request) {
	var service *models.Service

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = json.Unmarshal(body, &service)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = h.repository.CreateService(service)
	if err != nil {
		errorHandler(w, err)
		return
	}

	proxy.LoadProxy(h.repository)
	responseHandler(w, http.StatusOK, service)
}

func (h servicesHandlerV1) getServiceById(w http.ResponseWriter, r *http.Request) {
	id := middlewares.GetIDFromContext(r.Context())

	service, err := h.repository.ReadService(id)
	if err != nil {
		errorHandler(w, err)
		return
	}

	responseHandler(w, http.StatusOK, service)
}

func (h servicesHandlerV1) putServiceById(w http.ResponseWriter, r *http.Request) {
	id := middlewares.GetIDFromContext(r.Context())

	var service *models.Service

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = json.Unmarshal(body, &service)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = h.repository.UpdateService(id, service)
	if err != nil {
		errorHandler(w, err)
		return
	}

	proxy.LoadProxy(h.repository)
	responseHandler(w, http.StatusOK, true)
}

func (h servicesHandlerV1) deleteServiceById(w http.ResponseWriter, r *http.Request) {
	id := middlewares.GetIDFromContext(r.Context())

	err := h.repository.DestroyService(id)
	if err != nil {
		errorHandler(w, err)
		return
	}

	proxy.LoadProxy(h.repository)
	responseHandler(w, http.StatusOK, true)
}
