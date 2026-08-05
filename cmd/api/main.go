// Package main is the entrypoint for the hospital middleware API server.
//
// It owns the process lifecycle and the dependency wiring: load configuration,
// open the database, build each layer from the bottom up, serve, and shut down
// cleanly. Go has no dependency injection container, so the graph is assembled
// by hand here - the one place that is allowed to know about every layer.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LidTleJao/hospital-middleware-api/internal/config"
	"github.com/LidTleJao/hospital-middleware-api/internal/database"
	"github.com/LidTleJao/hospital-middleware-api/internal/repository"
	"github.com/LidTleJao/hospital-middleware-api/internal/service"
	httptransport "github.com/LidTleJao/hospital-middleware-api/internal/transport/http"
	"github.com/LidTleJao/hospital-middleware-api/internal/transport/http/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	db, err := database.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	// Bottom up: repositories talk to the pool, services talk to repositories,
	// handlers talk to services. Nothing points back up the chain.
	hospitalRepo := repository.NewHospital(db)
	staffRepo := repository.NewStaff(db)

	staffService := service.NewStaff(hospitalRepo, staffRepo)

	staffHandler := handler.NewStaff(staffService)

	router := httptransport.NewRouter(httptransport.Handlers{
		Staff: staffHandler,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// ListenAndServe blocks, so it runs in its own goroutine and main is left
	// free to wait for a shutdown signal.
	go func() {
		log.Printf("api listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("api stopped")
}
