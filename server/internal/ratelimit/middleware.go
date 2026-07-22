package ratelimit

import "net/http"

type KeyFunc func(*http.Request) string

type Middleware struct {
	store   *Store
	enabled bool
}

func NewMiddleware(store *Store, enabled bool) *Middleware {
	return &Middleware{
		store:   store,
		enabled: enabled,
	}
}

func (m *Middleware) Limit(
	namespace string,
	policy Policy,
	keyFunc KeyFunc,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !m.enabled {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)

			if key == "" {
				WriteExceeded(w, policy)
				return
			}

			fullKey := namespace + ":" + key

			if !m.store.Allow(fullKey, policy) {
				WriteExceeded(w, policy)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) ByIP(namespace string, policy Policy) func(http.Handler) http.Handler {
	return m.Limit(namespace, policy, ClientIP)
}