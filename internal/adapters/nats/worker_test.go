package natsadp

import (
	"testing"

	"payment-system/internal/domain"
)

func TestBuildExternalID(t *testing.T) {
	got := domain.BuildExternalID(domain.AddCardEventType, "customer-123", "payment-456")
	want := "add-card_customer-123_payment-456"
	if got != want {
		t.Fatalf("BuildExternalID() = %q, want %q", got, want)
	}
}

func TestParseExternalID(t *testing.T) {
	eventType, customerID, entityID, err := domain.ParseExternalID("payment_customer-123_order-456")
	if err != nil {
		t.Fatalf("ParseExternalID returned error: %v", err)
	}
	if eventType != "payment" || customerID != "customer-123" || entityID != "order-456" {
		t.Fatalf(
			"ParseExternalID() = %q, %q, %q; want payment/customer-123/order-456",
			eventType,
			customerID,
			entityID,
		)
	}
}
