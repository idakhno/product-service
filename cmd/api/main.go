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
func main() {
	if err := run(); err != nil {
		log.Fatalf("server returned an error: %v", err)
	}
}

func run() error {
	cfg := config.MustLoad()

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Environment:      cfg.Env,
	}); err != nil {
		return fmt.Errorf("sentry initialization failed: %w", err)
	}
	defer sentry.Flush(2 * time.Second)

	dbpool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	logger := logger.NewSlogAdapter(cfg.Env)
	logger.Info("logger initialized", "environment", cfg.Env)

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

	userRepo := postgresrepo.NewUserRepository(dbpool)
	productRepo := postgresrepo.NewProductRepository(dbpool)
	orderRepo := postgresrepo.NewOrderRepository(dbpool)
	productService := service.NewProductService(productRepo)
	orderService := service.NewOrderService(dbpool, orderRepo, productRepo, logger)      // Если dbpool нужен для транзакций, это ок
	usersService := service.NewUsersService(userRepo, []byte(cfg.JWTSecret), cfg.JWTTTL) // Используем cfg.JWTSecret
	userHandler := handler.NewUserHandler(usersService, logger)
	productHandler := handler.NewProductHandler(productService, logger)
	orderHandler := handler.NewOrderHandler(orderService, logger)
	sentryHandler := sentryhttp.New(sentryhttp.Options{})

	router := setupRouter(sentryHandler, userHandler, productHandler, orderHandler, cfg)

	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting server", "address", cfg.HTTPServer.Address)
		serverErrors <- server.ListenAndServe()
	}()

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

func setupRouter(sentryHandler *sentryhttp.Handler, userHandler *handler.UserHandler, productHandler *handler.ProductHandler, orderHandler *handler.OrderHandler, cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	r.Use(sentryHandler.Handle)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "server")
	})

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Post("/users/register", userHandler.Register)
	r.Post("/users/login", userHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(handler.JWTMiddleware([]byte(cfg.JWTSecret)))

		r.Post("/products", productHandler.Create)
		r.Get("/products/{id}", productHandler.GetByID)

		r.Post("/orders", orderHandler.Create)
	})

	return r
}

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
