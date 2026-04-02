package jobs

type EmailType string

const (
	EmailTypeOTP              EmailType = "otp"
	EmailTypeWelcome          EmailType = "welcome"
	EmailTypeDriverWelcome    EmailType = "driver_welcome"
	EmailTypeRideConfirmation EmailType = "ride_confirmation"
	EmailTypeRideCompleted    EmailType = "ride_completed"
	EmailTypeDriverApproved   EmailType = "driver_approved"
	EmailTypePasswordReset	EmailType = "password_reset"
)

type EmailJob struct {
	Type                 EmailType
	To                   string
	Name                 string
	OTP                  string
	RideConfirmationData RideConfirmationData
	RideCompletedData    RideCompletedData
	ResetLink string
}

type RideConfirmationData struct {
	RiderName      string
	DriverName     string
	VehicleMake    string
	VehicleModel   string
	PlateNumber    string
	VehicleColor   string
	PickupAddress  string
	DropoffAddress string
	EstimatedFare  float64
	ETA            int
	Year           int
}

type RideCompletedData struct {
	RiderName      string
	PickupAddress  string
	DropoffAddress string
	ActualFare     float64
	DistanceKm     float64
	Year           int
}

type Mailer interface {
	SendOTPEmail(to, name, otp string) error
	SendWelcomeEmail(to, name string) error
	SendDriverWelcomeEmail(to, name string) error
	SendRideConfirmationEmail(to string, data RideConfirmationData) error
	SendRideCompletedEmail(to string, data RideCompletedData) error
	SendDriverApprovedEmail(to string, text string) error
	SendPasswordResetEmail(to string, otp string) error
}
