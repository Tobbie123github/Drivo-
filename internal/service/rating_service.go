package service

import (
	"context"
	"drivo/internal/models"
	"drivo/internal/repository"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type RatingService struct {
	ratingRepo *repository.RatingRepo
	rideRepo   *repository.RideRepo
	driverRepo *repository.DriverRepo
}

func NewRatingService(
	ratingRepo *repository.RatingRepo,
	rideRepo *repository.RideRepo,
	driverRepo *repository.DriverRepo,
) *RatingService {
	return &RatingService{
		ratingRepo: ratingRepo,
		rideRepo:   rideRepo,
		driverRepo: driverRepo,
	}
}


func (s *RatingService) RateDriver(ctx context.Context, riderUserID uuid.UUID, input models.RatingInput) error {

	 rideID, err := uuid.Parse(input.RideID)
    if err != nil {
        return errors.New("invalid ride id")
    }

	 ride, err := s.rideRepo.GetRideByID(ctx, rideID)
    if err != nil {
        return fmt.Errorf("ride not found: %v", err)
    }

	 if ride.Status != models.RideStatusCompleted {
        return errors.New("can only rate a completed ride")
    }

	// Confirm the rater is the rider on this ride
    if ride.RiderID != riderUserID {
        return errors.New("you are not the rider on this ride")
    }

	if ride.DriverID == nil {
        return errors.New("no driver assigned to this ride")
    }

	driver, err := s.driverRepo.GetDriverByID(*ride.DriverID)
    if err != nil {
        return fmt.Errorf("driver not found: %v", err)
    }

	rating := models.Rating{
        RideID:    rideID,
        RaterID:   riderUserID,
        RateeID:   driver.UserID, 
        Score:     input.Score,
        Comment:   input.Comment,
        RaterRole: "rider",
    }

	if err := s.ratingRepo.CreateRating(ctx, rating); err != nil {
        return fmt.Errorf("failed to create rating: %v", err)
    }

	avg, err := s.ratingRepo.GetAverageRating(ctx, driver.UserID)
    if err != nil {
        return fmt.Errorf("failed to calculate average: %v", err)
    }

    if err := s.ratingRepo.UpdateDriverRating(ctx, driver.UserID, avg); err != nil {
        return fmt.Errorf("failed to update driver rating: %v", err)
    }

	fmt.Printf("Driver %s rated %.1f by rider %s\n", driver.UserID, float64(input.Score), riderUserID)

    return nil
}