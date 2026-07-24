package ratelimit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// ensures middleware allows requests to pass through when not rate limited.
func TestMiddlewareAllowsRequest(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	middleware := NewMiddleware(store, true)
	handlerCalled := false

	handler := middleware.ByIP("test", testPolicy(1))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusNoContent)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.0.67:12345"

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNoContent)
	}

	if !handlerCalled {
		t.Fatal("next handler was not called")
	}
}

// ensure when burst limit is exceeded, middleware returns 429.
func TestMiddlewareReturns429AfterBurst(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	middleware := NewMiddleware(store, true)

	policy := Policy{
		Limit:		rate.Every(time.Hour),
		Burst:		1,
		RetryAfter: 17 * time.Second,
	}

	handler := middleware.ByIP("test", policy)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)

	firstReq := httptest.NewRequest(http.MethodGet, "/", nil)
	firstReq.RemoteAddr = "192.0.0.67:12345"

	firstRes := httptest.NewRecorder()
	handler.ServeHTTP(firstRes, firstReq)

	if firstRes.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d", firstRes.Code, http.StatusNoContent)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/", nil)
	secondReq.RemoteAddr = "192.0.0.67:12346"

	secondRes := httptest.NewRecorder()
	handler.ServeHTTP(secondRes, secondReq)

	if secondRes.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", secondRes.Code, http.StatusTooManyRequests)
	}

	if got := secondRes.Header().Get("Retry-After"); got != "17" {
		t.Fatalf("Retry-After = %q, want %q", got, "17")
	}

	if got := secondRes.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body map[string]string

	if err := json.NewDecoder(secondRes.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body["error"] != "too many requests" {
		t.Fatalf("error = %q, want %q", body["error"], "too many requests")
	}
}

// ensure middleware separates IPs and handles each independently.
func TestMiddlewareSeparatesIPs(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	middleware := NewMiddleware(store, true)

	handler := middleware.ByIP("test", testPolicy(1))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	)

	firstReq := httptest.NewRequest(http.MethodGet, "/", nil)
	firstReq.RemoteAddr = "192.0.0.67:12345"

	firstRes := httptest.NewRecorder()
	handler.ServeHTTP(firstRes, firstReq)

	secondReq := httptest.NewRequest(http.MethodGet, "/", nil)
	secondReq.RemoteAddr = "192.0.0.68:12345"

	secondRes := httptest.NewRecorder()
	handler.ServeHTTP(secondRes, secondReq)

	if firstRes.Code != http.StatusNoContent {
		t.Fatalf("first IP status = %d, want 204", firstRes.Code)
	}

	if secondRes.Code != http.StatusNoContent {
		t.Fatalf("second IP status = %d, want 204", secondRes.Code)
	}
}

// ensure middleware rate limits each namespace independently.
func TestMiddlewareSeparatesNamespaces(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	middleware := NewMiddleware(store, true)
	policy := testPolicy(1)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	loginHandler := middleware.ByIP("login", policy)(next)
	registerHandler := middleware.ByIP("request", policy)(next)

	loginReq := httptest.NewRequest(http.MethodGet, "/", nil)
	loginReq.RemoteAddr = "192.0.0.67:12345"

	loginRes := httptest.NewRecorder()
	loginHandler.ServeHTTP(loginRes, loginReq)

	registerReq := httptest.NewRequest(http.MethodGet, "/", nil)
	registerReq.RemoteAddr = "192.0.0.67:12346"

	registerRes := httptest.NewRecorder()
	registerHandler.ServeHTTP(registerRes, registerReq)

	if loginRes.Code != http.StatusNoContent {
		t.Fatalf("login status = %d, want 204", loginRes.Code)
	}

	if registerRes.Code != http.StatusNoContent {
		t.Fatalf("register status = %d, want 204", registerRes.Code)
	}
}

// ensure middleware rejects requests with an empty key.
func TestMiddlewareRejectsEmptyKey(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	middleware := NewMiddleware(store, true)
	handlerCalled := false

	handler := middleware.Limit("test", testPolicy(1), func(*http.Request) string {
		return ""
	})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusTooManyRequests)
	}

	if handlerCalled {
		t.Fatal("next handler was called with an empty key")
	}
}

// ensure disabled middleware always passes requests through.
func TestDisabledMiddlewarePassesThrough(t *testing.T) {
	store := NewStore(time.Minute, time.Minute)
	t.Cleanup(store.Close)

	middleware := NewMiddleware(store, false)
	callCount := 0

	handler := middleware.ByIP("test", testPolicy(1))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusNoContent)
		}),
	)

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.0.67:12345"

		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		if res.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want 204", i + 1, res.Code)
		}
	}

	if callCount != 10 {
		t.Fatalf("handler call count = %d, want 10", callCount)
	}
}