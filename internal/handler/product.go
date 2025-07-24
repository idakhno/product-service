package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"product-api/internal/logger"
	"product-api/internal/service"
	customvalidator "product-api/pkg/validator"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CreateProductRequest struct {
	Description string   `json:"description" example:"High-quality wireless headphones" validate:"required"`
	Tags        []string `json:"tags" example:"audio,electronics,wireless" validate:"required"`
	Quantity    int      `json:"quantity" example:"100" validate:"required,gt=0"`
	Price       float64  `json:"price" example:"99.99" validate:"required,gt=0"`
}

type ProductHandler struct {
	service *service.ProductService
	logger  logger.Logger
}

func NewProductHandler(s *service.ProductService, l logger.Logger) *ProductHandler {
	return &ProductHandler{service: s, logger: l}
}

// Create godoc
// @Summary Create a new product
// @Tags products
// @Accept  json
// @Produce  json
// @Param   product  body      CreateProductRequest  true  "Product details"
// @Security ApiKeyAuth
// @Success 201  {object}  domain.Product
// @Failure 400  {string}  string "Invalid request body"
// @Failure 401  {string}  string "Unauthorized"
// @Failure 500  {string}  string "Internal server error"
// @Router /products [post]
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "ProductHandler.Create"
	log := h.logger.WithTrace(r.Context())

	var req CreateProductRequest
	if err := customvalidator.DecodeAndValidate(r, &req); err != nil {
		customvalidator.HandleValidationError(w, err)
		return
	}

	product, err := h.service.CreateProduct(r.Context(), req.Description, req.Tags, req.Quantity, req.Price)
	if err != nil {
		log.Error("failed to create product", "op", op, "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(product); err != nil {
		log.Error("failed to encode product response", "op", op, "err", err)
	}
}

// GetByID godoc
// @Summary Get a product by ID
// @Tags products
// @Produce  json
// @Param   id   path      string  true  "Product ID"
// @Security ApiKeyAuth
// @Success 200  {object}  domain.Product
// @Failure 400  {string}  string "Invalid product ID"
// @Failure 401  {string}  string "Unauthorized"
// @Failure 404  {string}  string "Product not found"
// @Failure 500  {string}  string "Internal server error"
// @Router /products/{id} [get]
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	const op = "ProductHandler.GetByID"
	log := h.logger.WithTrace(r.Context())

	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			http.Error(w, "product not found", http.StatusNotFound)
			return
		}
		log.Error("failed to get product by id", "op", op, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		log.Error("failed to encode product response", "op", op, "err", err)
	}
}
