package models

import (
    "time"
    "github.com/google/uuid"
)

type Rating struct {
    ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    RideID     uuid.UUID `gorm:"type:uuid;not null;index"`
    RaterID    uuid.UUID `gorm:"type:uuid;not null"`  
    RateeID    uuid.UUID `gorm:"type:uuid;not null"` 
    Score      int       `gorm:"not null"`           
    Comment    string    `gorm:"type:varchar(500)"`
    RaterRole  string    `gorm:"type:varchar(20);not null"` 
    CreatedAt  time.Time

    Ride Ride `gorm:"foreignKey:RideID" json:"-"`
}


type RatingInput struct {
    RideID  string `json:"ride_id"  binding:"required"`
    Score   int    `json:"score"    binding:"required,min=1,max=5"`
    Comment string `json:"comment"`
}