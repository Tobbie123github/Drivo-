package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type DriverStatus string

const (
	DriverPending   DriverStatus = "pending"
	DriverActive    DriverStatus = "active"
	DriverOffline   DriverStatus = "offline"
	DriverSuspended DriverStatus = "suspended"
	DriverBanned    DriverStatus = "banned"
)

type Driver struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`

	DOB     *time.Time
	Gender  *string `gorm:"type:varchar(255)" json:"gender"`
	Address *string `gorm:"type:varchar(255)" json:"address"`
	City    *string `gorm:"type:varchar(255)" json:"city"`
	State   *string `gorm:"type:varchar(255)" json:"state"`
	Country *string `gorm:"type:varchar(255)" json:"country"`

	Status          DriverStatus `gorm:"type:driver_status;not null;default:pending;index"`
	LicenseNumber   string       `gorm:"uniqueIndex;not null;size:50"`
	LicenseExpiry   string       `gorm:"not null"`
	LicenseImage    *string
	LicenseVerified bool `gorm:"default:false"`

	NationalIdImage *string
	SelfieImage     *string
	ProofOfAddress  *string

	AgreeTerms bool `gorm:"not null;default:false"`

	IsOnline         bool     `gorm:"not null;default:false;index"`
	CurrentLat       *float64 `gorm:"type:decimal(9,6)"`
	CurrentLng       *float64 `gorm:"type:decimal(9,6)"`
	LastLocationAt   *time.Time
	Rating           float64 `gorm:"type:decimal(3,2);not null;default:5.00"`
	TotalTrips       int     `gorm:"not null;default:0"`
	AcceptanceRate   float64 `gorm:"type:decimal(5,2);not null;default:100.00"`
	CancellationRate float64 `gorm:"type:decimal(5,2);not null;default:0.00"`

	CreatedAt time.Time
	UpdatedAt time.Time

	IsIdentityVerified bool `gorm:"default:false"`
	IsVehicleVerified  bool `gorm:"default:false"`

	OnboardingStep        int  `gorm:"not null;default:1"`
	IsOnboardingCompleted bool `gorm:"not null;default:false"`

	User     User `gorm:"foreignKey:UserID" json:"-"`
	Vehicles []Vehicle
}

type Vehicle struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DriverID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Make         string    `gorm:"not null;size:50"`
	Model        string    `gorm:"not null;size:50"`
	Year         int16     `gorm:"not null"`
	Color        string    `gorm:"not null;size:30"`
	PlateNumber  string    `gorm:"uniqueIndex;not null;size:20"`
	VehicleImage *string
	Category     string `gorm:"not null;size:20"`
	Seats        int16  `gorm:"not null;default:4"`
	IsVerified   bool   `gorm:"not null;default:false"`
	IsActive     bool   `gorm:"not null;default:true"`
	CreatedAt    time.Time

	Driver Driver `gorm:"foreignKey:DriverID" json:"-"`
}

type PendingDriver struct {
	Name         string `json:"name" redis:"name"`
	Email        string `json:"email" redis:"email"`
	HashPassword string `json:"hashPassword" redis:"hashPassword"`
	Phone        string `json:"phone" redis:"phone"`
	OTP          string `json:"otp" redis:"otp"`
	IsVerified   bool   `json:"is_verified" redis:"is_verified"`
	IsActive     bool   `json:"is_active" redis:"is_active"`
}

type DriverRegisterInput struct {
	Name     string `json:"name"  validate:"required,min=2,max=100"`
	Password string `json:"password"  validate:"required"`
	Phone    string `json:"phone" validate:"required,e164"`
	Email    string `json:"email" validate:"omitempty,email"`
}

type DriverVerifyEmail struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type DriverLoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResult struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type DriverProfileInput struct {
	FullName string                `form:"fullname"`
	DOB      string                `form:"dob"`
	Gender   string                `form:"gender"`
	Address  string                `form:"address"`
	City     string                `form:"city"`
	State    string                `form:"state"`
	Country  string                `form:"country"`
	Avatar   *multipart.FileHeader `form:"avatar"`
}

type DriverLicence struct {
	LicenseNumber string                `form:"licensenumber"`
	LicenseExpiry string                `form:"licenseexpiry"`
	LicenseImage  *multipart.FileHeader `form:"licenseimage"`
}

type VehicleInput struct {
	Make         string                `form:"make"`
	Model        string                `form:"model"`
	Year         int16                 `form:"year"`
	Color        string                `form:"color"`
	PlateNumber  string                `form:"plate_number"`
	Category     string                `form:"category"`
	Seats        int16                 `form:"seats"`
	VehicleImage *multipart.FileHeader `form:"vehicle_image"`
}

type DocumentUploadInput struct {
	NationalIdImage *multipart.FileHeader `form:"national_id_image"`
	SelfieImage     *multipart.FileHeader `form:"selfie_image"`
	ProofOfAddress  *multipart.FileHeader `form:"proof_of_address"`
}

type StatusUpdate struct {
	IsOnline bool `json:"is_online"`
}

type LocationData struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	UpdatedAt time.Time `json:"updated_at"`
}
