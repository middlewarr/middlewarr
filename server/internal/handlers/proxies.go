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

type proxiesHandlerV1 struct {
	repository *store.ConfigurationRepository
}

func newProxiesRoutesV1(repository *store.ConfigurationRepository) chi.Router {
	h := &proxiesHandlerV1{repository}

	r := chi.NewRouter()

	r.Get("/", h.getProxy)   // GET /proxy
	r.Post("/", h.postProxy) // POST /proxy

	r.With(middlewares.WithID).Group(func(r chi.Router) {
		r.Route("/{id:[0-9]+}", func(r chi.Router) {
			r.Get("/", h.getProxyById)       // GET /proxy/{id}
			r.Put("/", h.putProxyById)       // PUT /proxy/{id}
			r.Delete("/", h.deleteProxyById) // DELETE /proxy/{id}
		})
	})

	return r
}

func (h proxiesHandlerV1) getProxy(w http.ResponseWriter, r *http.Request) {
	proxies, err := h.repository.ReadProxies()
	if err != nil {
		errorHandler(w, err)
		return
	}

	responseHandler(w, http.StatusOK, proxies)
}

func (h proxiesHandlerV1) postProxy(w http.ResponseWriter, r *http.Request) {
	var p *models.Proxy

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = h.repository.CreateProxy(p)
	if err != nil {
		errorHandler(w, err)
		return
	}

	proxy.LoadProxy(h.repository)
	responseHandler(w, http.StatusOK, p)
}

func (h proxiesHandlerV1) getProxyById(w http.ResponseWriter, r *http.Request) {
	id := middlewares.GetIDFromContext(r.Context())

	p, err := h.repository.ReadProxy(id)
	if err != nil {
		errorHandler(w, err)
		return
	}

	responseHandler(w, http.StatusOK, p)
}

func (h proxiesHandlerV1) putProxyById(w http.ResponseWriter, r *http.Request) {
	id := middlewares.GetIDFromContext(r.Context())

	var p *models.Proxy

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		errorHandler(w, err)
		return
	}

	err = h.repository.UpdateProxy(id, p)
	if err != nil {
		errorHandler(w, err)
		return
	}

	proxy.LoadProxy(h.repository)
	responseHandler(w, http.StatusOK, true)
}

func (h proxiesHandlerV1) deleteProxyById(w http.ResponseWriter, r *http.Request) {
	id := middlewares.GetIDFromContext(r.Context())

	err := h.repository.DestroyProxy(id)
	if err != nil {
		errorHandler(w, err)
		return
	}

	proxy.LoadProxy(h.repository)
	responseHandler(w, http.StatusOK, true)
}
