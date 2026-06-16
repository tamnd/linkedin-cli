package linkedin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestFetch_OK checks that the client reads a plain 200 body.
func TestFetch_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := NewClient(DefaultConfig())
	body, code, err := c.Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if code != 200 {
		t.Errorf("code = %d, want 200", code)
	}
	if string(body) != "ok" {
		t.Errorf("body = %q, want ok", body)
	}
}

// TestFetch_Blocked999 checks that LinkedIn's HTTP 999 bot-block status
// returns ErrBlocked with no retry.
func TestFetch_Blocked999(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(999)
	}))
	defer srv.Close()

	c := NewClient(DefaultConfig())
	_, _, err := c.Fetch(context.Background(), srv.URL)
	if !errors.Is(err, ErrBlocked) {
		t.Errorf("want ErrBlocked for HTTP 999, got %v", err)
	}
	if hits != 1 {
		t.Errorf("server hit %d times, want 1 (no retry on 999)", hits)
	}
}

// TestFetch_BlockedAuthwall checks that a redirect to /authwall triggers
// ErrBlocked.
func TestFetch_BlockedAuthwall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/authwall" {
			// Serve the authwall page (as LinkedIn does).
			_, _ = w.Write([]byte("sign in to continue"))
			return
		}
		// Redirect the initial request to the authwall.
		http.Redirect(w, r, "/authwall", http.StatusFound)
	}))
	defer srv.Close()

	c := NewClient(DefaultConfig())
	_, _, err := c.Fetch(context.Background(), srv.URL+"/in/someone")
	if !errors.Is(err, ErrBlocked) {
		t.Errorf("want ErrBlocked for authwall redirect, got %v", err)
	}
}

// TestFetch_RateLimited checks that a sustained 429 after retries returns
// ErrRateLimited.
func TestFetch_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Retries = 2
	cfg.Delay = 0
	c := NewClient(cfg)
	_, _, err := c.Fetch(context.Background(), srv.URL)
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("want ErrRateLimited after sustained 429, got %v", err)
	}
}

// TestFetch_404 checks that a 404 returns (nil, 404, nil) without error.
func TestFetch_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClient(DefaultConfig())
	body, code, err := c.Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Fetch 404: unexpected error %v", err)
	}
	if code != 404 {
		t.Errorf("code = %d, want 404", code)
	}
	if body != nil {
		t.Errorf("body = %q, want nil for 404", body)
	}
}

// TestFetch_ServerError checks that 5xx responses are retried.
func TestFetch_ServerError(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("recovered"))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Retries = 5
	cfg.Delay = 0
	c := NewClient(cfg)
	body, code, err := c.Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Fetch after 5xx retries: %v", err)
	}
	if code != 200 || string(body) != "recovered" {
		t.Errorf("code = %d body = %q, want 200 recovered", code, body)
	}
	if hits != 3 {
		t.Errorf("server hit %d times, want 3", hits)
	}
}

// TestIsAuthWall covers the known login-gate path patterns.
func TestIsAuthWall(t *testing.T) {
	gates := []string{
		"https://www.linkedin.com/authwall?trk=test",
		"https://www.linkedin.com/uas/login",
		"https://www.linkedin.com/login",
		"https://www.linkedin.com/checkpoint/lg/sign-in",
		"https://www.linkedin.com/signup/cold-join",
	}
	for _, u := range gates {
		if !isAuthWall(u) {
			t.Errorf("isAuthWall(%q) = false, want true", u)
		}
	}
	if isAuthWall("https://www.linkedin.com/company/microsoft") {
		t.Error("isAuthWall(company page) = true, want false")
	}
}
