package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatSession struct {
	ID        uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RideID    uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex" json:"ride_id"` 
	DriverID  uuid.UUID     `gorm:"type:uuid;not null" json:"driver_id"`
	RiderID   uuid.UUID     `gorm:"type:uuid;not null" json:"rider_id"`
	IsActive  bool          `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Messages  []ChatMessage `gorm:"foreignKey:SessionID" json:"messages,omitempty"`
}

type SenderType string

const (
	SenderTypeDriver SenderType = "driver"
	SenderTypeRider  SenderType = "rider"
)

type ChatMessage struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SessionID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"session_id"`
	SenderID   uuid.UUID  `gorm:"type:uuid;not null" json:"sender_id"`
	SenderType SenderType `gorm:"type:varchar(10)" json:"sender_type"`
	Message    string     `gorm:"type:text;not null" json:"message"`
	IsRead     bool       `gorm:"default:false" json:"is_read"`
	CreatedAt  time.Time  `json:"created_at"`
}
