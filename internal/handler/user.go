package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"product-api/internal/logger"
	"product-api/internal/service"
	customvalidator "product-api/pkg/validator"
)

type RegisterRequest struct {
	Email     string `json:"email" example:"user@example.com" validate:"required,email"`
	Password  string `json:"password" example:"password123" validate:"required,min=8"`
	Firstname string `json:"firstname" example:"John" validate:"required"`
	Lastname  string `json:"lastname" example:"Doe" validate:"required"`
	Age       int    `json:"age" example:"25" validate:"required,gte=18"`
	IsMarried bool   `json:"is_married" example:"false"`
}

type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
	Password string `json:"password" example:"password123" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

type UserHandler struct {
	service *service.UsersService
	logger  logger.Logger
}

func NewUserHandler(s *service.UsersService, l logger.Logger) *UserHandler {
	return &UserHandler{service: s, logger: l}
}

// Register godoc
// @Summary Register a new user
// @Tags users
// @Accept  json
// @Produce  json
// @Param   user  body      RegisterRequest  true  "User registration details"
// @Success 201   {object}  domain.User
// @Failure 400   {string}  string "Invalid request body or validation error"
// @Failure 409   {string}  string "User with this email already exists"
// @Failure 500   {string}  string "Internal server error"
// @Router /users/register [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	log := h.logger.WithTrace(r.Context())

	var req RegisterRequest
	if err := customvalidator.DecodeAndValidate(r, &req); err != nil {
		customvalidator.HandleValidationError(w, err)
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Password, req.Firstname, req.Lastname, req.Age, req.IsMarried)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			log.Error("failed to register user", "err", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Error("failed to encode user response", "err", err)
	}
}

// Login godoc
// @Summary Log in a user
// @Tags users
// @Accept  json
// @Produce  json
// @Param   credentials  body      LoginRequest  true  "User credentials"
// @Success 200        {object}  LoginResponse
// @Failure 400        {string}  string "Invalid request body"
// @Failure 401        {string}  string "Invalid email or password"
// @Failure 500        {string}  string "Internal server error"
// @Router /users/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.Login"
	log := h.logger.WithTrace(r.Context())

	var req LoginRequest
	if err := customvalidator.DecodeAndValidate(r, &req); err != nil {
		customvalidator.HandleValidationError(w, err)
		return
	}

	token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		log.Error("failed to login user", "op", op, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(LoginResponse{Token: token}); err != nil {
		log.Error("failed to write login response", "op", op, "error", err)
	}
}
