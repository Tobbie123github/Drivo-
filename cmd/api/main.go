package main

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/config"
	"drivo/internal/server"
	"drivo/pkg/fcm"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {

	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("Error occured loading config: %v", err)
	}

	if cfg.FirebaseCredentialsPath != "" {
		if err := fcm.Init(cfg.FirebaseCredentialsPath); err != nil {
			log.Printf("FCM init failed: %v", err)
		}
	}

	a, err := app.NewApp(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(a)

	router, scheduler := server.NewRouter(a, cfg)

	addr := fmt.Sprintf(":%s", cfg.PORT)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server Error: %v", err)
		}
	}()

	fmt.Println("Server started on:", addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Shutting down server...")

	scheduler.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

}
