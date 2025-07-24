package domain

import "github.com/google/uuid"

type Product struct {
	ID          uuid.UUID
	Description string
	Tags        []string
	Quantity    int
	Price       float64
}
