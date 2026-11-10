package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/web3-frozen/demo-api/internal/downstream"
	"github.com/web3-frozen/demo-api/internal/model"
)

type CheckoutHandler struct {
	inventory *downstream.Client
	logger    *slog.Logger
}

func NewCheckoutHandler(inventory *downstream.Client, logger *slog.Logger) *CheckoutHandler {
	return &CheckoutHandler{inventory: inventory, logger: logger}
}

func (h *CheckoutHandler) Create(w http.ResponseWriter, r *http.Request) {
	if h.inventory == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "inventory service is not configured"})
		return
	}

	var request model.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if message := request.Validate(); message != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	checkoutID := uuid.New().String()
	reservation, err := h.inventory.Reserve(ctx, model.ReservationRequest{
		CheckoutID: checkoutID,
		Priority:   request.Priority,
	})
	if err != nil {
		h.logger.Error("checkout reservation failed", "checkout_id", checkoutID, "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "checkout dependencies unavailable"})
		return
	}

	writeJSON(w, http.StatusCreated, model.CheckoutResult{
		CheckoutID:  checkoutID,
		Reservation: *reservation,
	})
}
