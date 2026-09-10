package domain

import (
	"time"

	"github.com/google/uuid"
)

type Card struct {
	CustomerID uuid.UUID
	CardID     uuid.UUID

	Token    string
	Type     string
	LinkedAt time.Time
	Last4    string
}

type WalletCard struct {
	CustomerID  uuid.UUID
	RedirectURL string
}

type Payment struct {
	CustomerID uuid.UUID
	CardID     uuid.UUID
	OrderID    uuid.UUID

	Amount    float64
	Currency  string
	Status    string
	CreatedAt time.Time
}
