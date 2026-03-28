package models

import (
	"time"

	"github.com/google/uuid"
)

type PoolStatus string

const (
	PoolStatusOpen      PoolStatus = "open"
	PoolStatusFull      PoolStatus = "full"
	PoolStatusActive    PoolStatus = "active"
	PoolStatusCompleted PoolStatus = "completed"
	PoolStatusCancelled PoolStatus = "cancelled"
)

type PoolGroup struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DriverID    *uuid.UUID `gorm:"type:uuid;index" json:"driver_id"`
	Status      PoolStatus `gorm:"type:pool_status;default:'open'" json:"status"`
	MaxRiders   int        `gorm:"default:3" json:"max_riders"`
	CurrentSize int        `gorm:"default:0" json:"current_size"`
	BaseFare    float64    `json:"base_fare"`
	FarePerHead float64    `json:"fare_per_head"`
	OriginLat   float64    `json:"origin_lat"`
	OriginLng   float64    `json:"origin_lng"`
	DestLat     float64    `json:"dest_lat"`
	DestLng     float64    `json:"dest_lng"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Rides       []Ride     `gorm:"foreignKey:PoolGroupID" json:"rides,omitempty"`
}
