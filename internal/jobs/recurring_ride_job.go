package jobs

import (
	"context"
	"drivo/internal/models"
	"drivo/internal/repository"
	"fmt"
	"math"
	"strings"
	"time"
)

type RecurringRideJob struct {
	recurringRepo *repository.RecurringRepo
	rideRepo      *repository.RideRepo
}

func NewRecurringRideJob(
	recurringRepo *repository.RecurringRepo,
	rideRepo *repository.RideRepo,
) *RecurringRideJob {
	return &RecurringRideJob{
		recurringRepo: recurringRepo,
		rideRepo:      rideRepo,
	}
}

func (j *RecurringRideJob) Run() {
	ctx := context.Background()
	now := time.Now()

	fmt.Printf("RecurringRideJob running at %s\n", now.Format("2006-01-02 15:04:05"))

	rides, err := j.recurringRepo.GetAllActive(ctx)
	if err != nil {
		fmt.Printf("RecurringRideJob error fetching recurring rides: %v\n", err)
		return
	}

	fmt.Printf("RecurringRideJob processing %d active recurring rides\n", len(rides))

	tomorrow := now.AddDate(0, 0, 1)

	for _, r := range rides {
		j.process(ctx, r, tomorrow)
	}
}

func (j *RecurringRideJob) process(ctx context.Context, r models.RecurringRide, targetDate time.Time) {

	if r.EndDate != nil && targetDate.After(*r.EndDate) {
		fmt.Printf("RecurringRideJob recurring %s expired — deactivating\n", r.ID)
		r.IsActive = false
		j.recurringRepo.Update(ctx, &r)
		return
	}

	if !matchesDay(targetDate, r.DaysOfWeek) {
		fmt.Printf("[RecurringRideJob] recurring %s — %s not in '%s'\n",
			r.ID, targetDate.Weekday().String(), r.DaysOfWeek)
		return
	}

	alreadyBooked, err := j.recurringRepo.AlreadyBookedForDate(ctx, r.ID, targetDate)
	if err != nil {
		fmt.Printf("[RecurringRideJob] duplicate check error for %s: %v\n", r.ID, err)
		return
	}
	if alreadyBooked {
		fmt.Printf("[RecurringRideJob] recurring %s already booked for %s — skipping\n",
			r.ID, targetDate.Format("2006-01-02"))
		return
	}

	scheduledAt, err := buildScheduledTime(targetDate, r.PickupTime, r.Timezone)
	if err != nil {
		fmt.Printf("[RecurringRideJob] bad time config for %s: %v\n", r.ID, err)
		return
	}

	distanceKm := haversine(r.PickupLat, r.PickupLng, r.DropoffLat, r.DropoffLng)
	estimatedFare := calcFare(distanceKm)

	ride := models.Ride{
		RiderID:         r.RiderID,
		PickupLat:       r.PickupLat,
		PickupLng:       r.PickupLng,
		DropoffLat:      r.DropoffLat,
		DropoffLng:      r.DropoffLng,
		PickupAddress:   r.PickupAddress,
		DropoffAddress:  r.DropoffAddress,
		Status:          models.RideStatusScheduled,
		EstimatedFare:   estimatedFare,
		DistanceKm:      distanceKm,
		IsScheduled:     true,
		ScheduledAt:     &scheduledAt,
		RecurringRideID: &r.ID,
	}

	created, err := j.rideRepo.CreateRide(ctx, ride)
	if err != nil {
		fmt.Printf("[RecurringRideJob] failed to create ride for %s: %v\n", r.ID, err)
		return
	}

	now := time.Now()
	r.LastBookedDate = &now
	r.TotalBooked++
	j.recurringRepo.Update(ctx, &r)

	fmt.Printf("RecurringRideJob created ride %s for recurring %s at %s\n",
		created.ID, r.ID, scheduledAt.Format("2006-01-02 15:04"))
}

func matchesDay(date time.Time, days string) bool {
	dayMap := map[time.Weekday]string{
		time.Monday:    "mon",
		time.Tuesday:   "tue",
		time.Wednesday: "wed",
		time.Thursday:  "thu",
		time.Friday:    "fri",
		time.Saturday:  "sat",
		time.Sunday:    "sun",
	}
	target := dayMap[date.Weekday()]
	for _, d := range strings.Split(days, ",") {
		if strings.TrimSpace(strings.ToLower(d)) == target {
			return true
		}
	}
	return false
}

func buildScheduledTime(date time.Time, pickupTime string, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}

	var hour, min int
	_, err = fmt.Sscanf(pickupTime, "%d:%d", &hour, &min)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid pickup_time format '%s', expected HH:MM", pickupTime)
	}

	return time.Date(
		date.Year(), date.Month(), date.Day(),
		hour, min, 0, 0, loc,
	), nil
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func calcFare(distanceKm float64) float64 {
	estimatedMinutes := (distanceKm / 30.0) * 60
	fare := 500.0 + (distanceKm * 150.0) + (estimatedMinutes * 20.0)
	return math.Round(fare/50) * 50
}
