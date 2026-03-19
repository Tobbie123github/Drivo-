package app

import (
	"context"
	"drivo/internal/config"
	"drivo/internal/database"
	"drivo/internal/models"
	"fmt"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type App struct {
	Config config.Config
	DB     *gorm.DB

	Redis *redis.Client
}

func NewApp(ctx context.Context) (*App, error) {

	cfg, err := config.Load()

	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	redisCli, err := database.Redis(ctx, cfg)

	if err != nil {
		return nil, fmt.Errorf("Cant load db: %v", err)
	}

	// connect to database

	db, err := database.Connect(cfg)

	if err != nil {
		return nil, fmt.Errorf("error connecting to database, %v", err)
	}

	// automigrate user model

	if err := db.Exec(`CREATE TYPE user_role AS ENUM ('user', 'driver', 'admin')`).Error; err != nil {
		fmt.Println("Enum may already exist, continuing...")
	}

	if err := db.Exec(`
CREATE TYPE driver_status AS ENUM ('pending','offline','active', 'suspended', 'banned')
`).Error; err != nil {
		fmt.Println("driver_status enum may already exist, continuing...")
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, fmt.Errorf("error running user migrations: %v", err)
	}
	if err := db.AutoMigrate(&models.Driver{}); err != nil {
		return nil, fmt.Errorf("error running driver migrations: %v", err)
	}
	if err := db.AutoMigrate(&models.Vehicle{}); err != nil {
		return nil, fmt.Errorf("error running vehicle migrations: %v", err)
	}
	if err := db.AutoMigrate(&models.Ride{}); err != nil {
		return nil, fmt.Errorf("error running ride migrations: %v", err)
	}
	if err := db.AutoMigrate(&models.Rating{}); err != nil {
		return nil, fmt.Errorf("error running rating migrations: %v", err)
	}
	fmt.Println("Migrations ran successfully")

	return &App{
		Config: cfg,
		DB:     db,
		Redis:  redisCli,
	}, nil
}
