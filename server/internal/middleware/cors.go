package middleware

import (
	"net/http"
	"strings"
)

type CORS struct {
	allowedOrigins         map[string]struct{}
	allowFirefoxExtensions bool
}

func NewCORS(
	allowedOrigins []string,
	allowFirefoxExtensions bool,
) *CORS {
	origins := make(map[string]struct{}, len(allowedOrigins))

	for _, origin := range allowedOrigins {
		origin = strings.TrimSuffix(
			strings.TrimSpace(origin),
			"/",
		)

		if origin != "" {
			origins[origin] = struct{}{}
		}
	}

	return &CORS{
		allowedOrigins:          origins,
		allowFirefoxExtensions: allowFirefoxExtensions,
	}
}

func (c *CORS) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSuffix(
			strings.TrimSpace(r.Header.Get("Origin")),
			"/",
		)

		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		if !c.originAllowed(origin) {
			http.Error(
				w,
				"origin not allowed",
				http.StatusForbidden,
			)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Add("Vary", "Origin")

		if r.Method == http.MethodOptions {
			handlePreflight(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c *CORS) originAllowed(origin string) bool {
	if _, allowed := c.allowedOrigins[origin]; allowed {
		return true
	}

	return c.allowFirefoxExtensions &&
		strings.HasPrefix(origin, "moz-extension://")
}

func handlePreflight(
	w http.ResponseWriter,
	r *http.Request,
) {
	requestedMethod := r.Header.Get("Access-Control-Request-Method")

	switch requestedMethod {
	case http.MethodGet, http.MethodPost:
	default:
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "600")

	w.Header().Add("Vary", "Access-Control-Request-Method")
	w.Header().Add("Vary", "Access-Control-Request-Headers")

	w.WriteHeader(http.StatusNoContent)
}