package model

import "testing"

func TestCheckoutRequestValidate(t *testing.T) {
	tests := []struct {
		name     string
		request  CheckoutRequest
		want     string
		priority string
	}{
		{name: "default priority", request: CheckoutRequest{}, want: "", priority: "medium"},
		{name: "valid priority", request: CheckoutRequest{Priority: "high"}, want: "", priority: "high"},
		{name: "invalid priority", request: CheckoutRequest{Priority: "urgent"}, want: "priority must be low, medium, or high", priority: "urgent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.request.Validate(); got != tt.want {
				t.Fatalf("Validate() = %q, want %q", got, tt.want)
			}
			if tt.request.Priority != tt.priority {
				t.Fatalf("priority = %q, want %q", tt.request.Priority, tt.priority)
			}
		})
	}
}

func TestReservationRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		request ReservationRequest
		want    string
	}{
		{name: "valid", request: ReservationRequest{CheckoutID: "checkout-1", Priority: "medium"}, want: ""},
		{name: "missing id", request: ReservationRequest{Priority: "medium"}, want: "checkout_id is required"},
		{name: "invalid priority", request: ReservationRequest{CheckoutID: "checkout-1", Priority: "urgent"}, want: "priority must be low, medium, or high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.request.Validate(); got != tt.want {
				t.Fatalf("Validate() = %q, want %q", got, tt.want)
			}
		})
	}
}
