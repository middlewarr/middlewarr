package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/joho/godotenv/autoload"
	"github.com/middlewarr/server/internal/handlers"
	"github.com/middlewarr/server/internal/proxy"
	"github.com/middlewarr/server/internal/store"
	"github.com/middlewarr/server/internal/templates"
	"github.com/middlewarr/server/internal/tools"
)

func main() {
	l := tools.GetLogger()
	s := tools.GetSettings()

	templates.SyncTemplates()
	templates.LoadTemplates()

	l.Info().Msg("Starting Middlewarr server...")

	c := store.NewConfigurationRepository()
	proxy.LoadProxy(c)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Api-Key"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}))

		// Internal: /api/admin/*
		handlers.SetupAdminRoutes(r, c)

		// Proxy: /api/*
		r.HandleFunc("/*", proxy.GetProxyHandle)
	})

	graceful(s.String("host"), "80", r)
}

func graceful(host string, port string, handler http.Handler) {
	l := tools.GetLogger()

	server := &http.Server{
		Addr:         host + ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverError := make(chan error, 1)

	go func() {
		l.Info().Msg(fmt.Sprintf("Server is running on %s", server.Addr))
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverError:
		l.Fatal().Err(err)
	case sig := <-stop:
		l.Info().Msg(fmt.Sprintf("Received shutdown signal %v", sig))
	}

	l.Info().Msg("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		l.Error().Msg(fmt.Sprintf("Server shutdown error: %v", err))
		return
	}

	l.Info().Msg("Server exited properly")
}
