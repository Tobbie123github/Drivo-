package repository

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RideRepo struct {
	db *app.App
}

func NewRideRepo(db *app.App) *RideRepo {
	return &RideRepo{
		db: db,
	}
}

func (r *RideRepo) CreateRide(ctx context.Context, ride models.Ride) (models.Ride, error) {
	if err := r.db.DB.WithContext(ctx).Create(&ride).Error; err != nil {
		return models.Ride{}, fmt.Errorf("failed to create ride: %v", err)
	}
	return ride, nil
}

func (r *RideRepo) GetRideByID(ctx context.Context, rideID uuid.UUID) (models.Ride, error) {
	var ride models.Ride
	if err := r.db.DB.WithContext(ctx).
		Preload("Rider").
		Where("id = ?", rideID).
		First(&ride).Error; err != nil {
		return models.Ride{}, fmt.Errorf("ride not found: %v", err)
	}
	return ride, nil
}

func (r *RideRepo) UpdateRideStatus(ctx context.Context, rideID uuid.UUID, status models.RideStatus, driverID *uuid.UUID) error {
	updates := map[string]interface{}{
		"status":     string(status),
		"updated_at": time.Now().UTC(),
	}

	if driverID != nil {
		updates["driver_id"] = driverID
	}

	result := r.db.DB.WithContext(ctx).
		Model(&models.Ride{}).
		Where("id = ?", rideID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update ride status: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no ride found with id: %s", rideID)
	}

	return nil
}

func (r *RideRepo) SaveRideCandidates(ctx context.Context, rideID uuid.UUID, driverIDs []uuid.UUID) error {
	candidates := make([]string, len(driverIDs))
	for i, id := range driverIDs {
		candidates[i] = id.String()
	}

	bytes, err := json.Marshal(candidates)
	if err != nil {
		return fmt.Errorf("failed to marshal candidates: %v", err)
	}

	key := fmt.Sprintf("ride:candidates:%s", rideID.String())
	return r.db.Redis.Set(ctx, key, bytes, 2*time.Minute).Err()
}

func (r *RideRepo) GetNextCandidate(ctx context.Context, rideID uuid.UUID) (uuid.UUID, error) {
	key := fmt.Sprintf("ride:candidates:%s", rideID.String())

	val, err := r.db.Redis.Get(ctx, key).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("no candidates found: %v", err)
	}

	var candidates []string
	if err := json.Unmarshal([]byte(val), &candidates); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal candidates: %v", err)
	}

	if len(candidates) == 0 {
		return uuid.Nil, fmt.Errorf("no more candidates available")
	}

	nextDriver := candidates[0]

	candidates = candidates[1:]
	updated, _ := json.Marshal(candidates)
	r.db.Redis.Set(ctx, key, updated, 2*time.Minute)

	return uuid.MustParse(nextDriver), nil
}

func (r *RideRepo) DeleteRideCandidates(ctx context.Context, rideID uuid.UUID) error {
	key := fmt.Sprintf("ride:candidates:%s", rideID.String())
	return r.db.Redis.Del(ctx, key).Err()
}

func (r *RideRepo) GetOngoingRide(ctx context.Context, driverID uuid.UUID) (models.Ride, error) {
	var ride models.Ride
	err := r.db.DB.WithContext(ctx).
		Where("driver_id = ? AND status IN ?", driverID, []string{"accepted", "ongoing"}).
		First(&ride).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Ride{}, err
		}
		return models.Ride{}, fmt.Errorf("failed to get ongoing ride: %v", err)
	}

	return ride, nil
}

func (r *RideRepo) CompleteRide(ctx context.Context, rideID uuid.UUID, actualFare float64) error {
	updates := map[string]interface{}{
		"status":      string(models.RideStatusCompleted),
		"actual_fare": actualFare,
		"updated_at":  time.Now().UTC(),
	}

	result := r.db.DB.WithContext(ctx).
		Model(&models.Ride{}).
		Where("id = ? AND status = ?", rideID, string(models.RideStatusOngoing)).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to complete ride: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("ride %s is not ongoing", rideID)
	}

	return nil
}

func (r *RideRepo) RiderHistory(ctx context.Context, riderID uuid.UUID) ([]models.Ride, error) {

	var rides []models.Ride
	err := r.db.DB.WithContext(ctx).
		Where("rider_id = ?", riderID).
		Order("created_at DESC").
		Find(&rides).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch rider history: %v", err)
	}

	return rides, nil
}

func (r *RideRepo) DriverHistory(ctx context.Context, driverID uuid.UUID) ([]models.Ride, error) {

	var rides []models.Ride

	err := r.db.DB.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Order("created_at DESC").
		Find(&rides).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch driver history: %v", err)
	}

	return rides, nil
}

func (r *RideRepo) UpdatePoolRidesFare(ctx context.Context, poolID uuid.UUID, newFarePerHead float64) error {
	return r.db.DB.WithContext(ctx).Model(&models.PoolGroup{}).Where("id = ?", poolID).Update("fare_per_head", newFarePerHead).Error
}

func (r *RideRepo) GetRidesByPoolID(ctx context.Context, poolID uuid.UUID) ([]models.Ride, error) {
	var pool models.PoolGroup

	err := r.db.DB.WithContext(ctx).
		Where("id = ?", poolID).
		Preload("Rides").
		First(&pool).Error

	if err != nil {
		return nil, err
	}

	return pool.Rides, nil
}

func (r *RideRepo) GetDueScheduledRides(ctx context.Context) ([]models.Ride, error) {
	var rides []models.Ride
	now := time.Now().UTC()

	err := r.db.DB.WithContext(ctx).
		Where("status = ?", models.RideStatusScheduled).
		Where("scheduled_at <= ?", now).
		Find(&rides).Error

	return rides, err
}
