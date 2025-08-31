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
	apiKeyHeader     string = "X-Api-Key"
	apiKeyQueryParam string = "apikey"
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
				proxy,
				parsedEndpoints,
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

func isRequestAllowed(endpoints ProxyEndpoints, method string, path string) bool {
	for _, endpoint := range endpoints {
		if method == endpoint.Method && endpoint.PathRegex.MatchString(path) {
			return true
		}
	}

	return false
}

func getApiKey(apiKey string, apiKeyParam string) string {
	if apiKey != "" {
		return apiKey
	}

	return apiKeyParam
}

func GetProxyHandle(w http.ResponseWriter, r *http.Request) {
	l := tools.GetLogger()

	apiKey := r.Header.Get(apiKeyHeader)
	apiKeyParam := r.URL.Query().Get(apiKeyQueryParam)

	if apiKey == "" && apiKeyParam == "" {
		http.Error(w, "", http.StatusUnauthorized)

		l.Error().
			Str("request_client", r.RemoteAddr).
			Str("request_method", r.Method).
			Str("request_url", tools.SanitizeURI(r)).
			Msg("Missing API key")
		return
	}

	proxyConfig, ok := getProxyRouter().ProxyByKey[getApiKey(apiKey, apiKeyParam)]
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

	if !isRequestAllowed(proxyConfig.endpoints, r.Method, r.URL.Path) {
		w.Header().Set("Cache-Control", "no-cache, no-store")
		w.Header().Set("Expires", "-1")
		w.Header().Set("Pragma", "no-cache")

		http.Error(w, "", http.StatusNotFound)

		l.Error().
			Str("proxy_service", service.Name).
			Str("proxy_app", app.Name).
			Str("proxy_type", service.Type).
			Str("proxy_url", service.URL).
			Str("request_client", r.RemoteAddr).
			Str("request_method", r.Method).
			Str("request_url", tools.SanitizeURI(r)).
			Msg("Forbidden, endpoint not allowed")
		return
	}

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

			if req.URL.Query().Get(apiKeyQueryParam) != "" {
				q := req.URL.Query()
				q.Set(apiKeyQueryParam, service.APIKey)

				req.URL.RawQuery = q.Encode()

				// Delete the unwanted `X-Api-Key` header if the `apikey` param is set.
				req.Header.Del(apiKeyHeader)
			} else {
				req.Header.Set(apiKeyHeader, service.APIKey)

				// Delete the unwanted `apikey` param if the `X-Api-Key` header is set.
				q := req.URL.Query()
				q.Del(apiKeyQueryParam)

				req.URL.RawQuery = q.Encode()
			}

			req.Host = serviceUrl.Host
		}

		proxy.ServeHTTP(w, r)
	})

	proxyHandler.ServeHTTP(w, r)
}
