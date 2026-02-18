package resolver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var errUnauthorized = errors.New("unauthorized")

const (
	maxRetries    = 10
	retryInterval = 30 * time.Second
)

type ApiResolver struct {
	baseURL    string
	accountID  string
	secret     string
	mu         sync.RWMutex
	routes     map[string]string
	token      string
	httpClient *http.Client
}

func (a *ApiResolver) login(ctx context.Context) error {
	body, err := json.Marshal(map[string]string{
		"user_account_id": a.accountID,
		"secret":          a.secret,
	})
	if err != nil {
		return fmt.Errorf("marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/service/login", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("login failed: unauthorized")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode login response: %w", err)
	}

	a.token = result.Token

	log.Debug().Msg("Successfully authenticated with MinecraftAdmin API")
	return nil
}

func (a *ApiResolver) fetchMapping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/service/mapping", nil)
	if err != nil {
		return fmt.Errorf("create mapping request: %w", err)
	}

	token := a.token

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("mapping request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return errUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mapping fetch failed: unexpected status %d", resp.StatusCode)
	}

	var routes map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return fmt.Errorf("decode mapping response: %w", err)
	}

	a.mu.Lock()
	a.routes = routes
	a.mu.Unlock()

	log.Debug().Int("RouteCount", len(routes)).Msg("Routing cache refreshed")
	return nil
}

func (a *ApiResolver) refresh(ctx context.Context) error {
	err := a.fetchMapping(ctx)
	if err == nil {
		return nil
	}
	if !errors.Is(err, errUnauthorized) {
		return err
	}

	log.Debug().Msg("Received 401 from mapping endpoint, re-authenticating")
	if err := a.login(ctx); err != nil {
		return fmt.Errorf("re-authentication failed: %w", err)
	}
	return a.fetchMapping(ctx)
}

func (a *ApiResolver) poll(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.refresh(ctx); err != nil {
				log.Warn().Err(err).Msg("Failed to refresh routing cache, keeping last-known-good cache")
				a.retryRefresh(ctx)
			}
		}
	}
}

func (a *ApiResolver) retryRefresh(ctx context.Context) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(retryInterval):
		}

		if err := a.refresh(ctx); err != nil {
			remaining := maxRetries - attempt
			if remaining == 0 {
				log.Fatal().
					Err(err).
					Int("Attempts", maxRetries).
					Msg("Routing cache has been stale for too long, shutting down")
			}
			log.Warn().
				Err(err).
				Int("RemainingAttempts", remaining).
				Msg("Retry failed, keeping last-known-good cache")
		} else {
			log.Info().Int("Attempt", attempt).Msg("Routing cache recovered after retry")
			return
		}
	}
}

func (a *ApiResolver) ResolveHostname(hostname string) (string, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	val, ok := a.routes[hostname]
	return val, ok
}

func (a *ApiResolver) MarshalZerologObject(e *zerolog.Event) {
	e.Str("ApiBaseUrl", a.baseURL)
}

func NewApiResolver(ctx context.Context, apiURL, accountID, secret string, pollInterval time.Duration) (*ApiResolver, error) {
	if _, err := url.Parse(apiURL); err != nil {
		return nil, fmt.Errorf("parse API URL: %w", err)
	}
	baseURL := strings.TrimRight(apiURL, "/")

	resolver := &ApiResolver{
		baseURL:   baseURL,
		accountID: accountID,
		secret:    secret,
		routes:    make(map[string]string),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	if err := resolver.login(ctx); err != nil {
		return nil, fmt.Errorf("initial authentication failed: %w", err)
	}

	if err := resolver.refresh(ctx); err != nil {
		return nil, fmt.Errorf("initial mapping fetch failed: %w", err)
	}

	go resolver.poll(ctx, pollInterval)

	return resolver, nil
}
