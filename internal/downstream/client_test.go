package downstream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/web3-frozen/demo-api/internal/model"
)

func TestClientReserveAndQuote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/reserve":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			var request model.ReservationRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Fatal(err)
			}
			if request.CheckoutID != "checkout-1" {
				t.Fatalf("checkout_id = %q", request.CheckoutID)
			}
			_ = json.NewEncoder(w).Encode(model.Reservation{ID: request.CheckoutID, Status: "reserved"})
		case "/quote":
			if got := r.URL.Query().Get("priority"); got != "high" {
				t.Fatalf("priority = %q, want high", got)
			}
			_ = json.NewEncoder(w).Encode(model.Quote{AmountCents: 399, Currency: "EUR", Provider: "test"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	reservation, err := client.Reserve(context.Background(), model.ReservationRequest{CheckoutID: "checkout-1", Priority: "high"})
	if err != nil {
		t.Fatal(err)
	}
	if reservation.ID != "checkout-1" || reservation.Status != "reserved" {
		t.Fatalf("unexpected reservation: %+v", reservation)
	}

	quote, err := client.Quote(context.Background(), "high")
	if err != nil {
		t.Fatal(err)
	}
	if quote.AmountCents != 399 || quote.Provider != "test" {
		t.Fatalf("unexpected quote: %+v", quote)
	}
}
