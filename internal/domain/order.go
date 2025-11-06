package domain

import (
	"time"

	"github.com/google/uuid"
)

// Order represents a user's order.
type Order struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Items       []OrderItem
	CreatedAt   time.Time
	TotalAmount float64 // Total order amount
}

// OrderItem represents a single item in an order.
// PriceAtPurchase stores the product price at the time of purchase.
type OrderItem struct {
	ID              uuid.UUID
	ProductID       uuid.UUID
	Quantity        int
	PriceAtPurchase float64 // Price at time of purchase
}
