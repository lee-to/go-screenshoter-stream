package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"screenshoter/internal"
	"syscall"
	"time"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env")
	}

	port := os.Getenv("APP_PORT")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(internal.Handle),
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("Server started on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	<-done
	fmt.Printf("Shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Printf("Server stopped gracefully")
}
