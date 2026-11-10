package pricing

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQuoteUsesProviderRate(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rates":{"EUR":0.8}}`))
	}))
	defer provider.Close()

	service := NewService(provider.URL, nil, slog.Default())
	quote := service.quote(context.Background(), "high")

	if quote.AmountCents != 319 {
		t.Fatalf("amount = %d, want 319", quote.AmountCents)
	}
	if quote.Currency != "EUR" || quote.Provider != "127.0.0.1" {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}

func TestQuoteFallsBackWhenProviderFails(t *testing.T) {
	service := NewService("http://127.0.0.1:1", nil, slog.Default())
	quote := service.quote(context.Background(), "low")

	if quote.AmountCents != 183 || quote.Provider != "fallback" {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}
