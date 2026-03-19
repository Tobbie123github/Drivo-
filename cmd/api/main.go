package main

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/config"
	"drivo/internal/server"
	"fmt"
	"log"
)

func main() {

	cfg, err := config.Load()

	if err != nil {
		log.Fatalf("Error occured loading config: %v", err)
	}

	a, err := app.NewApp(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(a)

	router := server.NewRouter(a, cfg)
	

	add := fmt.Sprintf(":%s", cfg.PORT)

	if err := router.Run(add); err != nil {
		log.Fatalf("Server Error")
	}

}
