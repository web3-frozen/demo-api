package inventory

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/web3-frozen/demo-api/internal/cache"
	"github.com/web3-frozen/demo-api/internal/downstream"
	"github.com/web3-frozen/demo-api/internal/model"
	"github.com/web3-frozen/demo-api/internal/store"
)

type Service struct {
	store   *store.PostgresStore
	cache   *cache.CheckoutCache
	pricing *downstream.Client
	logger  *slog.Logger
}

func NewService(pg *store.PostgresStore, rc *cache.CheckoutCache, pricing *downstream.Client, logger *slog.Logger) *Service {
	return &Service{store: pg, cache: rc, pricing: pricing, logger: logger}
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := s.store.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("POST /reserve", s.reserve)
	return mux
}

func (s *Service) reserve(w http.ResponseWriter, r *http.Request) {
	var request model.ReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if message := request.Validate(); message != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
		return
	}

	quote, err := s.pricing.Quote(r.Context(), request.Priority)
	if err != nil {
		s.logger.Error("pricing quote failed", "checkout_id", request.CheckoutID, "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "pricing service unavailable"})
		return
	}

	reservation, err := s.store.CreateReservation(r.Context(), request.CheckoutID, request.Priority, *quote)
	if err != nil {
		s.logger.Error("persist reservation failed", "checkout_id", request.CheckoutID, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist reservation"})
		return
	}

	if s.cache != nil {
		payload := map[string]any{
			"status":      reservation.Status,
			"priority":    reservation.Priority,
			"quote":       reservation.Quote,
			"created_at":  reservation.CreatedAt,
			"checkout_id": reservation.ID,
		}
		if err := s.cache.SetJSON(r.Context(), "checkout:"+reservation.ID, payload, 10*time.Minute); err != nil {
			s.logger.Warn("cache reservation failed", "checkout_id", reservation.ID, "error", err)
		}
	}

	writeJSON(w, http.StatusCreated, reservation)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
