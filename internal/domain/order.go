package domain

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Items       []OrderItem
	CreatedAt   time.Time
	TotalAmount float64
}

type OrderItem struct {
	ID              uuid.UUID
	ProductID       uuid.UUID
	Quantity        int
	PriceAtPurchase float64
}
