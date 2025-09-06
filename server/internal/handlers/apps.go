package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/middlewarr/server/internal/models"
	"github.com/middlewarr/server/internal/proxy"
	"github.com/middlewarr/server/internal/store"
)

type appsHandlerV1 struct {
	repository *store.ConfigurationRepository
}

func newAppsRoutesV1(repository *store.ConfigurationRepository) chi.Router {
	h := &appsHandlerV1{repository}

	r := chi.NewRouter()

	r.Get("/", h.getApp)   // GET /app
	r.Post("/", h.postApp) // POST /app

	r.With(withID).Group(func(r chi.Router) {
		r.Route("/{id:[0-9]+}", func(r chi.Router) {
			r.Get("/", h.getAppById)       // GET /app/{id}
			r.Put("/", h.putAppById)       // PUT /app/{id}
			r.Delete("/", h.deleteAppById) // DELETE /app/{id}
		})
	})

	return r
}

func (h appsHandlerV1) getApp(w http.ResponseWriter, r *http.Request) {
	apps, err := h.repository.ReadApps()
	if err != nil {
		errorHandler(w, err)
		return
	}

	responseHandler(w, http.StatusOK, apps)
}

func (h appsHandlerV1) postApp(w http.ResponseWriter, r *http.Request) {
	var app *models.App

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = json.Unmarshal(body, &app)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = h.repository.CreateApp(app)
	if err != nil {
		errorHandler(w, err)
		return
	}

	proxy.LoadProxy(h.repository)
	responseHandler(w, http.StatusOK, app)
}

func (h appsHandlerV1) getAppById(w http.ResponseWriter, r *http.Request) {
	id := getIDFromContext(r.Context())

	app, err := h.repository.ReadApp(id)
	if err != nil {
		errorHandler(w, err)
		return
	}

	responseHandler(w, http.StatusOK, app)
}

func (h appsHandlerV1) putAppById(w http.ResponseWriter, r *http.Request) {
	id := getIDFromContext(r.Context())

	var app *models.App

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = json.Unmarshal(body, &app)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = h.repository.UpdateApp(id, app)
	if err != nil {
		errorHandler(w, err)
		return
	}

	proxy.LoadProxy(h.repository)
	responseHandler(w, http.StatusOK, true)
}

func (h appsHandlerV1) deleteAppById(w http.ResponseWriter, r *http.Request) {
	id := getIDFromContext(r.Context())

	err := h.repository.DestroyApp(id)
	if err != nil {
		errorHandler(w, err)
		return
	}

	proxy.LoadProxy(h.repository)
	responseHandler(w, http.StatusOK, true)
}
