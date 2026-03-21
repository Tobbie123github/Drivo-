package app

import (
	"context"
	"drivo/internal/config"
	"drivo/internal/database"
	"drivo/internal/models"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type App struct {
	Config config.Config
	DB     *gorm.DB
	Redis  *redis.Client
}

func NewApp(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	redisCli, err := database.Redis(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("cant connect to redis: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		log.Println("Running migrations...")
		if err := runMigrations(db); err != nil {
			return nil, err
		}
		log.Println("Migrations complete ✅")
	} else {
		log.Println("Skipping migrations (RUN_MIGRATIONS not set)")
	}

	return &App{
		Config: cfg,
		DB:     db,
		Redis:  redisCli,
	}, nil
}

func runMigrations(db *gorm.DB) error {
	
	enums := []string{
		`CREATE TYPE user_role AS ENUM ('user', 'driver', 'admin')`,
		`CREATE TYPE driver_status AS ENUM ('pending','offline','active','suspended','banned')`,
	}
	for _, sql := range enums {
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("Enum may already exist, continuing: %v", err)
		}
	}

	
	if err := db.AutoMigrate(
		&models.User{},
		&models.Driver{},
		&models.Vehicle{},
		&models.Ride{},
		&models.Rating{},
	); err != nil {
		return fmt.Errorf("error running migrations: %v", err)
	}

	return nil
}
