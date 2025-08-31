package tools

import "net/http"

func SanitizeURI(r *http.Request) string {
	url := *r.URL

	query := url.Query()
	if apiKey := query.Get("apikey"); apiKey != "" {
		redactedApiKey := apiKey[:7] + "..." + apiKey[len(apiKey)-7:]
		query.Set("apikey", redactedApiKey)

		url.RawQuery = query.Encode()
	}

	return url.RequestURI()
}
