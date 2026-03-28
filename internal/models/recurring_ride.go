package models

import (
	"time"

	"github.com/google/uuid"
)

type RecurringRide struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RiderID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"rider_id"`
	PickupLat      float64    `json:"pickup_lat"`
	PickupLng      float64    `json:"pickup_lng"`
	DropoffLat     float64    `json:"dropoff_lat"`
	DropoffLng     float64    `json:"dropoff_lng"`
	PickupAddress  string     `gorm:"type:varchar(300)" json:"pickup_address"`
	DropoffAddress string     `gorm:"type:varchar(300)" json:"dropoff_address"`
	DaysOfWeek     string     `gorm:"type:varchar(50)" json:"days_of_week"` // "mon,tue,wed,thu,fri"
	PickupTime     string     `gorm:"type:varchar(10)" json:"pickup_time"`  // "07:00"
	Timezone       string     `gorm:"type:varchar(50);default:'Africa/Lagos'" json:"timezone"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	LastBookedDate *time.Time `json:"last_booked_date"`
	TotalBooked    int        `gorm:"default:0" json:"total_booked"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
