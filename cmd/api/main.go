package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"product-api/internal/config"
	"product-api/internal/handler"
	"product-api/internal/logger"
	postgresrepo "product-api/internal/repository/postgres"
	"product-api/internal/service"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"

	_ "product-api/docs"
)

// @title Product API
// @version 1.0
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// main is the entry point of the application.
// Initializes all service components and starts the HTTP server.
func main() {
	if err := run(); err != nil {
		log.Fatalf("server returned an error: %v", err)
	}
}

// run initializes all application components and starts the HTTP server.
// Performs graceful shutdown when receiving termination signals.
func run() error {
	// Load configuration from environment variables
	cfg := config.MustLoad()

	// Initialize Sentry for error monitoring
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Environment:      cfg.Env,
	}); err != nil {
		return fmt.Errorf("sentry initialization failed: %w", err)
	}
	defer sentry.Flush(2 * time.Second)

	// Create database connection pool
	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}
	defer dbpool.Close()

	// Verify database connection
	if err := dbpool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	// Initialize logger
	logger := logger.NewSlogAdapter(cfg.Env)
	logger.Info("logger initialized", "environment", cfg.Env)

	// Initialize OpenTelemetry tracer
	tp, err := initTracer()
	if err != nil {
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	otel.SetTracerProvider(tp)

	// Initialize repositories
	userRepo := postgresrepo.NewUserRepository(dbpool)
	productRepo := postgresrepo.NewProductRepository(dbpool)
	orderRepo := postgresrepo.NewOrderRepository(dbpool)

	// Initialize services
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(dbpool, orderRepo, productRepo, logger)
	usersService := service.NewUsersService(userRepo, []byte(cfg.JWTSecret), cfg.JWTTTL)

	// Initialize HTTP handlers
	userHandler := handler.NewUserHandler(usersService, logger)
	productHandler := handler.NewProductHandler(productService, logger)
	orderHandler := handler.NewOrderHandler(orderService, logger)
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	// Setup router
	router := setupRouter(sentryHandler, userHandler, productHandler, orderHandler, cfg)

	// Create HTTP server with timeout settings
	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a separate goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting server", "address", cfg.HTTPServer.Address)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for either server error or shutdown signal
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
	}

	return nil
}

// setupRouter configures HTTP router with middleware and routes.
// Public routes: user registration and authentication.
// Protected routes (require JWT token): product and order operations.
func setupRouter(sentryHandler *sentryhttp.Handler, userHandler *handler.UserHandler, productHandler *handler.ProductHandler, orderHandler *handler.OrderHandler, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware for error handling and monitoring
	r.Use(sentryHandler.Handle)           // Sentry for error tracking
	r.Use(middleware.Recoverer)           // Panic recovery
	r.Use(middleware.RequestID)           // Generate unique ID for each request
	r.Use(middleware.RealIP)              // Get real client IP
	r.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "server") // OpenTelemetry tracing
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Public routes (no authentication required)
	r.Post("/users/register", userHandler.Register)
	r.Post("/users/login", userHandler.Login)

	// Protected routes (require JWT token)
	r.Group(func(r chi.Router) {
		r.Use(handler.JWTMiddleware([]byte(cfg.JWTSecret)))

		// Product routes
		r.Post("/products", productHandler.Create)
		r.Get("/products/{id}", productHandler.GetByID)

		// Order routes
		r.Post("/orders", orderHandler.Create)
	})

	return r
}

// initTracer initializes OpenTelemetry tracer for request tracing.
// Uses stdout exporter to output traces to console.
func initTracer() (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	return tp, nil
}
