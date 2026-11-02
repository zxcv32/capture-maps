package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimit(t *testing.T) {
	calls := 0
	handler := RateLimit(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})

	newRequest := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/print", nil)
		req.RemoteAddr = "203.0.113.1:54321"
		return req
	}

	for i := 0; i < rateLimitBurst; i++ {
		rec := httptest.NewRecorder()
		handler(rec, newRequest())
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d within burst: want 200, got %d", i, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	handler(rec, newRequest())
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request past burst: want 429, got %d", rec.Code)
	}

	if calls != rateLimitBurst {
		t.Fatalf("want handler invoked %d times, got %d", rateLimitBurst, calls)
	}

	otherIP := httptest.NewRequest(http.MethodPost, "/print", nil)
	otherIP.RemoteAddr = "203.0.113.2:54321"
	rec = httptest.NewRecorder()
	handler(rec, otherIP)
	if rec.Code != http.StatusOK {
		t.Fatalf("different IP should have its own bucket: want 200, got %d", rec.Code)
	}
}
