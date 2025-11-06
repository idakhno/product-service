package domain

import "github.com/google/uuid"

// Product represents a product in the system.
type Product struct {
	ID          uuid.UUID
	Description string
	Tags        []string
	Quantity    int     // Product quantity in stock
	Price       float64 // Product price
}
