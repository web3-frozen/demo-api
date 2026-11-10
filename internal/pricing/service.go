package pricing

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/web3-frozen/demo-api/internal/cache"
	"github.com/web3-frozen/demo-api/internal/model"
)

type Service struct {
	client      *http.Client
	providerURL string
	cache       *cache.CheckoutCache
	logger      *slog.Logger
}

type providerResponse struct {
	Rates map[string]float64 `json:"rates"`
}

func NewService(providerURL string, rc *cache.CheckoutCache, logger *slog.Logger) *Service {
	return &Service{
		client:      &http.Client{Timeout: 1500 * time.Millisecond},
		providerURL: providerURL,
		cache:       rc,
		logger:      logger,
	}
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /quote", s.quoteHandler)
	return mux
}

func (s *Service) quoteHandler(w http.ResponseWriter, r *http.Request) {
	priority := r.URL.Query().Get("priority")
	if priority == "" {
		priority = "medium"
	}
	if priority != "low" && priority != "medium" && priority != "high" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "priority must be low, medium, or high"})
		return
	}
	writeJSON(w, http.StatusOK, s.quote(r.Context(), priority))
}

func (s *Service) quote(ctx context.Context, priority string) model.Quote {
	baseCents := map[string]int{"low": 199, "medium": 299, "high": 399}[priority]
	rate := 0.92
	provider := "fallback"

	if s.providerURL != "" {
		if liveRate, liveProvider, err := s.fetchProvider(ctx); err == nil {
			rate = liveRate
			provider = liveProvider
		} else {
			s.logger.Warn("market data provider unavailable; using fallback rate", "error", err)
		}
	}

	quote := model.Quote{
		AmountCents: int(math.Round(float64(baseCents) * rate)),
		Currency:    "EUR",
		Rate:        rate,
		Provider:    provider,
	}
	if s.cache != nil {
		if err := s.cache.SetJSON(ctx, "pricing:last", quote, 10*time.Minute); err != nil {
			s.logger.Warn("cache quote failed", "error", err)
		}
	}
	return quote
}

func (s *Service) fetchProvider(ctx context.Context) (float64, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.providerURL, nil)
	if err != nil {
		return 0, "", fmt.Errorf("create provider request: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("request provider: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return 0, "", fmt.Errorf("provider returned %d", resp.StatusCode)
	}

	var payload providerResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, "", fmt.Errorf("decode provider response: %w", err)
	}
	rate := payload.Rates["EUR"]
	if rate <= 0 {
		return 0, "", fmt.Errorf("provider response did not include a positive EUR rate")
	}

	parsed, err := url.Parse(s.providerURL)
	if err != nil {
		return rate, "market-data", nil
	}
	return rate, parsed.Hostname(), nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
