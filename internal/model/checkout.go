package model

import "time"

type CheckoutRequest struct {
	Priority string `json:"priority"`
}

func (r *CheckoutRequest) Validate() string {
	if r.Priority == "" {
		r.Priority = "medium"
	}
	if !contains(ValidPriorities, r.Priority) {
		return "priority must be low, medium, or high"
	}
	return ""
}

type ReservationRequest struct {
	CheckoutID string `json:"checkout_id"`
	Priority   string `json:"priority"`
}

func (r *ReservationRequest) Validate() string {
	if r.CheckoutID == "" {
		return "checkout_id is required"
	}
	if !contains(ValidPriorities, r.Priority) {
		return "priority must be low, medium, or high"
	}
	return ""
}

type Quote struct {
	AmountCents int     `json:"amount_cents"`
	Currency    string  `json:"currency"`
	Rate        float64 `json:"rate"`
	Provider    string  `json:"provider"`
}

type Reservation struct {
	ID        string    `json:"id"`
	Priority  string    `json:"priority"`
	Status    string    `json:"status"`
	Quote     Quote     `json:"quote"`
	CreatedAt time.Time `json:"created_at"`
}

type CheckoutResult struct {
	CheckoutID  string      `json:"checkout_id"`
	Reservation Reservation `json:"reservation"`
}
