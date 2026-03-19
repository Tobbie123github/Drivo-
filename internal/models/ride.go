package models

import (
	"time"

	"github.com/google/uuid"
)

type RideStatus string

const (
	RideStatusPending   RideStatus = "pending"
	RideStatusAccepted  RideStatus = "accepted"
	RideStatusOngoing   RideStatus = "ongoing"
	RideStatusCompleted RideStatus = "completed"
	RideStatusCancelled RideStatus = "cancelled"
)

type Ride struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RiderID  uuid.UUID  `gorm:"type:uuid;not null;index"`
	DriverID *uuid.UUID `gorm:"type:uuid;index"`

	PickupLat  float64 `gorm:"type:decimal(9,6);not null"`
	PickupLng  float64 `gorm:"type:decimal(9,6);not null"`
	DropoffLat float64 `gorm:"type:decimal(9,6);not null"`
	DropoffLng float64 `gorm:"type:decimal(9,6);not null"`

	PickupAddress  string `gorm:"type:varchar(255)"`
	DropoffAddress string `gorm:"type:varchar(255)"`

	Status        RideStatus `gorm:"column:status;type:varchar(20);not null;default:pending"`
	EstimatedFare float64    `gorm:"type:decimal(10,2);not null"`
	ActualFare    *float64   `gorm:"type:decimal(10,2)"`
	DistanceKm    float64    `gorm:"type:decimal(10,2);not null"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Rider  User    `gorm:"foreignKey:RiderID" json:"-"`
	Driver *Driver `gorm:"foreignKey:DriverID" json:"-"`
}

type RideRequestInput struct {
	PickupLat      float64 `json:"pickup_lat"  binding:"required"`
	PickupLng      float64 `json:"pickup_lng"  binding:"required"`
	DropoffLat     float64 `json:"dropoff_lat" binding:"required"`
	DropoffLng     float64 `json:"dropoff_lng" binding:"required"`
	PickupAddress  string  `json:"pickup_address"`
	DropoffAddress string  `json:"dropoff_address"`
}

type RideRequestNotification struct {
	RideID         uuid.UUID `json:"ride_id"`
	PickupLat      float64   `json:"pickup_lat"`
	PickupLng      float64   `json:"pickup_lng"`
	DropoffLat     float64   `json:"dropoff_lat"`
	DropoffLng     float64   `json:"dropoff_lng"`
	PickupAddress  string    `json:"pickup_address"`
	DropoffAddress string    `json:"dropoff_address"`
	EstimatedFare  float64   `json:"estimated_fare"`
	DistanceKm     float64   `json:"distance_km"`
	RiderName      string    `json:"rider_name"`
	RiderRating    float64   `json:"rider_rating"`
}

type RideResponseInput struct {
	RideID uuid.UUID `json:"ride_id"`
	Action string    `json:"action"`
}
