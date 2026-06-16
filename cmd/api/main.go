package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"jobscout/internal/applications"
	"jobscout/internal/auth"
	"jobscout/internal/config"
	"jobscout/internal/database"
	"jobscout/internal/jobs"
	"jobscout/internal/jobsource"
	"jobscout/internal/middleware"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	authSvc := auth.NewService(pool, cfg)
	authHandler := auth.NewHandler(authSvc)

	// Job source + search
	jobSource := jobsource.NewRemotive(cfg.RemotiveBaseURL)
	jobSvc := jobs.NewService(jobSource)
	jobHandler := jobs.NewHandler(jobSvc)

	// Applications
	appSvc := applications.NewService(pool)
	appHandler := applications.NewHandler(appSvc)

	r := chi.NewRouter()

	// Public routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
		})
	})

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg))

		r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
			userID := middleware.UserIDFromContext(r.Context())
			resp, err := authSvc.GetUser(r.Context(), userID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		})

		r.Get("/jobs/search", jobHandler.Search)

		// Applications
		r.Route("/applications", func(r chi.Router) {
			r.Post("/", appHandler.Save)
		})
	})

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped")
}