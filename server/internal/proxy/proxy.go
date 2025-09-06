package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/middlewarr/server/internal/models"
	"github.com/middlewarr/server/internal/store"
	"github.com/middlewarr/server/internal/templates"
	"github.com/middlewarr/server/internal/tools"
)

const (
	apiKeyHeaderKey     string = "X-Api-Key"
	apiKeyQueryParamKey string = "apikey"
)

type ProxyEndpoint struct {
	Method    string
	PathRegex *regexp.Regexp
}

type ProxyEndpoints []ProxyEndpoint

type ProxyConfig struct {
	proxy     models.Proxy
	endpoints ProxyEndpoints
}

type ProxyRouter struct {
	ProxyByKey map[string]*ProxyConfig
}

var proxyRouter atomic.Value

func getProxyRouter() *ProxyRouter {
	return proxyRouter.Load().(*ProxyRouter)
}

func setProxyRouter(pr *ProxyRouter) {
	proxyRouter.Store(pr)
}

func LoadProxy(c *store.ConfigurationRepository) {
	l := tools.GetLogger()

	l.Info().
		Msg("Loading configuration...")

	pr := &ProxyRouter{
		ProxyByKey: make(map[string]*ProxyConfig),
	}

	store.ValidateConfig(c)

	// TODO: Handle errors
	apps, _ := c.ReadApps()

	for _, app := range *apps {
		template, err := templates.ReadTemplate(app.Template)
		if err != nil {
			l.Error().
				Err(err).
				Str("app_name", app.Name).
				Str("app_template", app.Template).
				Msg("Invalid template, no proxy will be configured")
			continue
		}

		for _, proxy := range app.Proxies {
			parsedEndpoints, err := parseEndpoints(template.Endpoints[proxy.Service.Type])
			if err != nil {
				l.Error().
					Err(err).
					Str("app_name", app.Name).
					Str("app_template", app.Template).
					Msg("Invalid template endpoint, no proxy will be configured")

				continue
			}

			if len(parsedEndpoints) == 0 {
				l.Warn().
					Str("proxy_service", proxy.Service.Name).
					Str("proxy_app", app.Name).
					Str("proxy_type", proxy.Service.Type).
					Str("proxy_url", proxy.Service.URL).
					Msg("No endpoints present")
			}

			if proxy.APIKey == "" {
				l.Error().
					Err(err).
					Str("proxy_service", proxy.Service.Name).
					Str("proxy_app", app.Name).
					Str("proxy_type", proxy.Service.Type).
					Str("proxy_url", proxy.Service.URL).
					Msg("Missing apiKey, no proxy will be configured")

				continue
			}

			pr.ProxyByKey[proxy.APIKey] = &ProxyConfig{
				proxy:     proxy,
				endpoints: parsedEndpoints,
			}
		}
	}

	setProxyRouter(pr)

	l.Info().
		Msg("Configuration loaded")
}

func parseEndpoints(endpoints map[string][]string) (ProxyEndpoints, error) {
	var proxyEndpoints ProxyEndpoints

	for path, methods := range endpoints {
		regexStr := regexp.QuoteMeta(path)
		// Replace placeholders like {id} with regex group
		regexStr = strings.ReplaceAll(regexStr, `\{`, "{")
		regexStr = strings.ReplaceAll(regexStr, `\}`, "}")
		regexStr = regexp.MustCompile(`\{[^/]+\}`).ReplaceAllString(regexStr, `[^/]+`)
		// Make the check case-insensitive
		regexStr = `(?i)^` + regexStr + `$`

		re, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid endpoint pattern %s: %w", path, err)
		}

		for _, m := range methods {
			proxyEndpoints = append(proxyEndpoints, ProxyEndpoint{
				Method:    strings.ToUpper(m),
				PathRegex: re,
			})
		}
	}

	return proxyEndpoints, nil
}

func getApiKey(apiKeyHeaderValue string, apiKeyQueryParamValue string) *string {
	if apiKeyHeaderValue != "" {
		return &apiKeyHeaderValue
	}

	if apiKeyQueryParamValue != "" {
		return &apiKeyQueryParamValue
	}

	return nil
}

func GetProxyHandle(w http.ResponseWriter, r *http.Request) {
	l := tools.GetLogger()

	apiKeyHeaderValue := r.Header.Get(apiKeyHeaderKey)
	apiKeyQueryParamValue := r.URL.Query().Get(apiKeyQueryParamKey)

	apiKey := getApiKey(apiKeyHeaderValue, apiKeyQueryParamValue)
	if apiKey == nil {
		http.Error(w, "", http.StatusUnauthorized)

		l.Error().
			Str("request_client", r.RemoteAddr).
			Str("request_method", r.Method).
			Str("request_url", tools.SanitizeURI(r)).
			Msg("Missing API key")
		return
	}

	proxyConfig, ok := getProxyRouter().ProxyByKey[*apiKey]
	if !ok {
		http.Error(w, "", http.StatusUnauthorized)

		l.Error().
			Str("request_client", r.RemoteAddr).
			Str("request_method", r.Method).
			Str("request_url", tools.SanitizeURI(r)).
			Msg("Invalid API key")
		return
	}

	app := proxyConfig.proxy.App
	service := proxyConfig.proxy.Service
	proxyID := fmt.Sprintf("%03d_%03d", app.ID, service.ID)

	if !*app.IsActive {
		http.Error(w, "", http.StatusUnauthorized)

		l.Warn().
			Str("request_client", r.RemoteAddr).
			Str("request_method", r.Method).
			Str("request_url", tools.SanitizeURI(r)).
			Bool("app_is_active", *app.IsActive).
			Msg("Inactive API key")
		return
	}

	l.Info().
		Str("proxy_service", service.Name).
		Str("proxy_app", app.Name).
		Str("proxy_type", service.Type).
		Str("proxy_url", service.URL).
		Str("request_client", r.RemoteAddr).
		Str("request_method", r.Method).
		Str("request_url", tools.SanitizeURI(r)).
		Msg("Proxing request")

	serviceUrl, err := url.Parse(service.URL)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)

		l.Error().
			Err(err).
			Str("proxy_service", service.Name).
			Str("proxy_app", app.Name).
			Str("proxy_type", service.Type).
			Str("request_client", r.RemoteAddr).
			Str("request_method", r.Method).
			Str("request_url", tools.SanitizeURI(r)).
			Msg("Invalid service URL")
		return
	}

	var proxyHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(serviceUrl)
		originalDirector := proxy.Director

		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Remove the `apikey` query param.
			q := req.URL.Query()
			q.Del(apiKeyQueryParamKey)

			req.URL.RawQuery = q.Encode()

			// Provide the API key using the X-Api-Key header.
			req.Header.Set(apiKeyHeaderKey, service.APIKey)

			req.Host = serviceUrl.Host
		}

		r.Header.Set("X-Proxy-Id", proxyID)
		r.Header.Set("X-Proxy-App", app.Name)
		r.Header.Set("X-Proxy-Service", service.Name)

		handler := chainMiddlewares(
			proxy,
			middlewareLogRequest(),
			middlewareValidateRequest(proxyConfig.endpoints),
		)

		handler.ServeHTTP(w, r)
	})

	proxyHandler.ServeHTTP(w, r)
}
