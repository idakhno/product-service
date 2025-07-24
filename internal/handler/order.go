package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"product-api/internal/logger"
	"product-api/internal/service"
	customvalidator "product-api/pkg/validator"

	"github.com/google/uuid"
)

type OrderItemInput struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type CreateOrderRequest struct {
	Items []OrderItemInput `json:"items" validate:"required,min=1,dive"`
}

type OrderHandler struct {
	service *service.OrderService
	logger  logger.Logger
}

func NewOrderHandler(s *service.OrderService, l logger.Logger) *OrderHandler {
	return &OrderHandler{service: s, logger: l}
}

// Create godoc
// @Summary Create a new order
// @Tags orders
// @Accept  json
// @Produce  json
// @Param   order  body      CreateOrderRequest  true  "Order details"
// @Security ApiKeyAuth
// @Success 201  {object}  domain.Order
// @Failure 400  {string}  string "Invalid request body or product not found"
// @Failure 401  {string}  string "Unauthorized"
// @Failure 500  {string}  string "Internal server error"
// @Router /orders [post]
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "OrderHandler.Create"
	log := h.logger.WithTrace(r.Context())

	var req CreateOrderRequest
	if err := customvalidator.DecodeAndValidate(r, &req); err != nil {
		customvalidator.HandleValidationError(w, err)
		return
	}

	userIDStr, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		log.Error("failed to get user id from context", "op", op)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Error("failed to parse user id", "op", op, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Convert handler DTO to service DTO
	serviceItems := make([]service.OrderItemInput, len(req.Items))
	for i, item := range req.Items {
		serviceItems[i] = service.OrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	order, err := h.service.CreateOrder(r.Context(), userID, serviceItems)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			http.Error(w, "one or more products not found", http.StatusBadRequest)
		case errors.Is(err, service.ErrInsufficientStock):
			http.Error(w, "insufficient stock for one or more products", http.StatusConflict)
		default:
			log.Error("failed to create order", "op", op, "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Error("failed to encode order response", "op", op, "error", err)
	}
}
