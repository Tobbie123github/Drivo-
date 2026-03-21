package ws

type MessageType string

const (
	MessageTypeLocation      MessageType = "location_update"
	MessageTypeRideResponse  MessageType = "ride_response"
	MessageTypeRideRequest   MessageType = "ride_request"
	MessageTypeDriverArrived MessageType = "driver_arrived"
	MessageTypeStartTrip     MessageType = "start_trip"
	MessageTypeEndTrip       MessageType = "end_trip"

	MessageTypeRideAccepted  MessageType = "ride_accepted"
	MessageTypeRideCancelled MessageType = "ride_cancelled"
	MessageTypeRideStarted   MessageType = "ride_started"
	MessageTypeRideCompleted MessageType = "ride_completed"
	MessageTypeDriverIsHere  MessageType = "driver_is_here"

	MessageTypeNoCandidates MessageType = "no_candidates"

	MessageTypeRideCancelledByRider  MessageType = "ride_cancelled_by_rider"
	MessageTypeRideCancelledByDriver MessageType = "ride_cancelled_by_driver"
	MessageTypeDriverCancel          MessageType = "driver_cancel"

	MessageTypeRateDriver MessageType = "rate_driver"
	MessageTypeRateRider  MessageType = "rate_rider"

	MessageTypeDriverLocation MessageType = "driver_location"

	MessageTypeError MessageType = "error"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type TripActionPayload struct {
	RideID string `json:"ride_id"`
}

type RideCompletedPayload struct {
	RideID      string  `json:"ride_id"`
	ActualFare  float64 `json:"actual_fare"`
	DistanceKm  float64 `json:"distance_km"`
	DurationMin int     `json:"duration_minutes"`
}

type LocationPayload struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RideAcceptedPayload struct {
	RideID       string  `json:"ride_id"`
	DriverName   string  `json:"driver_name"`
	DriverPhone  string  `json:"driver_phone"`
	VehicleMake  string  `json:"vehicle_make"`
	VehicleModel string  `json:"vehicle_model"`
	PlateNumber  string  `json:"plate_number"`
	VehicleColor string  `json:"vehicle_color"`
	Rating       float64 `json:"rating"`
	ETA          int     `json:"eta_minutes"`
}
