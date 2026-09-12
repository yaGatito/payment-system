package domain

import (
	"fmt"
	"strings"
)

const (
	AddCardEventType = "add-card"
	PaymentEventType = "payment"
)

const (
	PendingPaymentStatus = "pending"
	FailurePaymentStatus = "failure"
	SuccessPaymentStatus = "success"
)

func BuildExternalID(eventType, customerID, entityID string) string {
	return strings.Join([]string{eventType, customerID, entityID}, "_")
}

func ParseExternalID(externalID string) (eventType, customerID, entityID string, err error) {
	parts := strings.Split(externalID, "_")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("invalid external_id format")
	}
	return parts[0], parts[1], parts[2], nil
}
