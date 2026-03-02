package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	iap "github.com/meetup/iap-api"
	"github.com/meetup/iap-service/handlers"
)

const (
	defaultPort   = 9090
	defaultCacheDir = "./tmp"
)

func main() {
	// Get configuration from environment
	port := getEnv("PORT", fmt.Sprintf("%d", defaultPort))
	cacheDir := getEnv("CACHE_DIR", defaultCacheDir)

	// Create cache
	cache, err := iap.NewCache(cacheDir)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}

	// Create and start biller
	biller := iap.NewBiller(cache)
	if err := biller.Start(); err != nil {
		log.Fatalf("Failed to start biller: %v", err)
	}

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create Gin server
	gin.SetMode(getEnv("GIN_MODE", "release"))
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginLogger())

	// Register handlers
	h := handlers.New(biller)
	h.RegisterRoutes(r)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting IAP Mock Server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Shutdown biller (save state)
	if err := biller.Shutdown(); err != nil {
		log.Printf("Warning: failed to shutdown biller: %v", err)
	}

	log.Println("Server exited")
}

// getEnv gets an environment variable or returns the default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ginLogger returns a gin middleware that logs requests.
func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		log.Printf("[%s] %s %s | %d | %v | %s",
			c.Request.Method,
			path,
			query,
			c.Writer.Status(),
			latency,
			c.ClientIP(),
		)
	}
}
