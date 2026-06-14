package linkedin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/time/rate"
)

// ErrBlocked signals that LinkedIn refused the page: an HTTP 999 bot status, a
// redirect to the sign-in/authwall, or a body that holds no public data where a
// JSON-LD graph was expected. Callers map it to the "blocked" exit code.
var ErrBlocked = errors.New("blocked: LinkedIn refused the page (sign-in wall or bot block)")

// ErrRateLimited signals a sustained HTTP 429 after retries.
var ErrRateLimited = errors.New("rate limited (HTTP 429)")

// ErrNotFound signals a 404. Callers map it to the no-data exit code.
var ErrNotFound = errors.New("not found")

// Client performs polite, retrying HTTP GETs against linkedin.com.
type Client struct {
	http       *http.Client
	userAgents []string
	limiter    *rate.Limiter
	retries    int
}

// newLimiter builds a token-bucket limiter: rate = workers/delay, burst = workers,
// so every worker can fire at startup but the average pace stays polite.
func newLimiter(cfg Config) *rate.Limiter {
	if cfg.Delay <= 0 || cfg.Workers <= 0 {
		return rate.NewLimiter(rate.Inf, 1)
	}
	r := rate.Limit(float64(cfg.Workers) / cfg.Delay.Seconds())
	return rate.NewLimiter(r, cfg.Workers)
}

func transport(cfg Config) *http.Transport {
	return &http.Transport{
		MaxIdleConns:        cfg.Workers + 4,
		MaxConnsPerHost:     cfg.Workers + 4,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

// NewClient builds an anonymous client.
func NewClient(cfg Config) *Client {
	retries := cfg.Retries
	if retries <= 0 {
		retries = DefaultRetries
	}
	return &Client{
		http:       &http.Client{Timeout: cfg.Timeout, Transport: transport(cfg)},
		userAgents: userAgents,
		limiter:    newLimiter(cfg),
		retries:    retries,
	}
}

// NewClientWithCookies builds a client pre-loaded with a lent session.
func NewClientWithCookies(cfg Config, cookies []*http.Cookie) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(BaseURL)
	jar.SetCookies(u, cookies)
	c := NewClient(cfg)
	c.http.Jar = jar
	return c, nil
}

// Fetch returns the raw body and status for a URL, retrying transient failures.
// A 404 returns (nil, 404, nil). A sign-in redirect or HTTP 999 returns
// ErrBlocked.
func (c *Client) Fetch(ctx context.Context, rawurl string) ([]byte, int, error) {
	max := c.retries
	if max < 1 {
		max = 1
	}
	for attempt := 1; attempt <= max; attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, 0, err
		}
		body, code, err := c.doGet(ctx, rawurl)
		if err != nil {
			if errors.Is(err, ErrBlocked) || attempt == max {
				return nil, code, err
			}
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		switch {
		case code == 429:
			if attempt == max {
				return nil, code, ErrRateLimited
			}
			time.Sleep(time.Duration(attempt*attempt) * 10 * time.Second)
			continue
		case code == 404:
			return nil, code, nil
		case code >= 500:
			if attempt == max {
				return nil, code, fmt.Errorf("server error HTTP %d", code)
			}
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
			continue
		}
		return body, code, nil
	}
	return nil, 0, fmt.Errorf("all %d attempts failed", max)
}

// FetchHTML fetches a URL and parses it into a goquery document.
func (c *Client) FetchHTML(ctx context.Context, rawurl string) (*goquery.Document, int, error) {
	body, code, err := c.Fetch(ctx, rawurl)
	if err != nil {
		return nil, code, err
	}
	if code == 404 {
		return nil, code, nil
	}
	if code != 200 {
		return nil, code, fmt.Errorf("unexpected HTTP %d for %s", code, rawurl)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, code, fmt.Errorf("parse HTML: %w", err)
	}
	return doc, code, nil
}

func (c *Client) doGet(ctx context.Context, rawurl string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", c.userAgents[rand.Intn(len(c.userAgents))])
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// The guest job endpoints expect a same-site referer; sending it on every
	// request is harmless and matches what a browser does.
	req.Header.Set("Referer", BaseURL+"/")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	// HTTP 999 is LinkedIn's own "we do not like this request" status. There is
	// no body worth reading and no point retrying the same way.
	if resp.StatusCode == 999 {
		return nil, 999, fmt.Errorf("%w (HTTP 999 for %s)", ErrBlocked, rawurl)
	}

	// The client follows redirects; landing on an authwall/login path means the
	// public view is gated.
	final := resp.Request.URL.String()
	if isAuthWall(final) {
		return nil, 401, fmt.Errorf("%w (%s redirected to %s)", ErrBlocked, rawurl, final)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// isAuthWall reports whether a final URL is one of LinkedIn's gate paths.
func isAuthWall(u string) bool {
	for _, p := range []string{"/authwall", "/uas/login", "/login", "/checkpoint", "/signup"} {
		if strings.Contains(u, p) {
			return true
		}
	}
	return false
}
