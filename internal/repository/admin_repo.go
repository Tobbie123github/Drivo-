package repository

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminRepo struct {
	db *app.App
}

func NewAdminRepo(db *app.App) *AdminRepo {
	return &AdminRepo{db: db}
}

// -----------------------------DRIVERS ---------------------------//

func (r *AdminRepo) GetAllDrivers(ctx context.Context, status string) ([]models.Driver, error) {
	var drivers []models.Driver

	query := r.db.DB.WithContext(ctx).Preload("User").Preload("Vehicles")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&drivers).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch drivers: %v", err)
	}

	return drivers, nil
}

func (r *AdminRepo) UpdateDriverStatus(ctx context.Context, driverID uuid.UUID, status models.DriverStatus) error {
	result := r.db.DB.WithContext(ctx).
		Model(&models.Driver{}).
		Where("id = ?", driverID).
		Exec("UPDATE drivers SET status = ?::driver_status WHERE id = ?", string(status), driverID)

	if result.Error != nil {
		return fmt.Errorf("failed to update driver status: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("driver not found: %s", driverID)
	}

	return nil
}

func (r *AdminRepo) VerifyDriverIdentity(ctx context.Context, driverID uuid.UUID) error {
	result := r.db.DB.WithContext(ctx).
		Model(&models.Driver{}).
		Where("id = ?", driverID).
		Update("is_identity_verified", true)

	if result.RowsAffected == 0 {
		return fmt.Errorf("driver not found: %s", driverID)
	}

	return result.Error


	
}

func (r *AdminRepo) VerifyDriverVehicle(ctx context.Context, driverID uuid.UUID) error {
	return r.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		
		vehicleResult := tx.Model(&models.Vehicle{}).
			Where("driver_id = ?", driverID).
			Update("is_verified", true)

		if vehicleResult.Error != nil {
			return vehicleResult.Error
		}

		if vehicleResult.RowsAffected == 0 {
			return fmt.Errorf("no vehicles found for driver: %s", driverID)
		}

		driverResult := tx.Model(&models.Driver{}).
			Where("id = ?", driverID).
			Update("is_vehicle_verified", true)

		if driverResult.Error != nil {
			return driverResult.Error
		}

		if driverResult.RowsAffected == 0 {
			return fmt.Errorf("driver not found: %s", driverID)
		}

		return nil
	})
}

func (r *AdminRepo) VerifyDriverLicense(ctx context.Context, driverID uuid.UUID) error {
	result := r.db.DB.WithContext(ctx).
		Model(&models.Driver{}).
		Where("id = ?", driverID).
		Update("license_verified", true)

	if result.RowsAffected == 0 {
		return fmt.Errorf("driver not found: %s", driverID)
	}

	return result.Error
}

// -------------------------RIDERS------------------------//

func (r *AdminRepo) GetAllRiders(ctx context.Context) ([]models.User, error) {
	var riders []models.User

	if err := r.db.DB.WithContext(ctx).
		Where("role = ?", "user").
		Order("created_at DESC").
		Find(&riders).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch riders: %v", err)
	}

	return riders, nil
}

func (r *AdminRepo) GetAllRides(ctx context.Context, status string) ([]models.Ride, error) {
	var rides []models.Ride

	query := r.db.DB.WithContext(ctx).
		Preload("Rider")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&rides).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch rides: %v", err)
	}

	return rides, nil
}

// ---------------------STATS------------------------------//

type DashboardStats struct {
	TotalDrivers     int64   `json:"total_drivers"`
	PendingDrivers   int64   `json:"pending_drivers"`
	ActiveDrivers    int64   `json:"active_drivers"`
	SuspendedDrivers int64   `json:"suspended_drivers"`
	TotalRiders      int64   `json:"total_riders"`
	TotalRides       int64   `json:"total_rides"`
	CompletedRides   int64   `json:"completed_rides"`
	CancelledRides   int64   `json:"cancelled_rides"`
	TotalEarnings    float64 `json:"total_earnings"`
	OnlineDrivers    int64   `json:"online_drivers"`
}

func (r *AdminRepo) GetDashboardStats(ctx context.Context) (DashboardStats, error) {
	var stats DashboardStats

	// Total drivers
	r.db.DB.WithContext(ctx).Model(&models.Driver{}).Count(&stats.TotalDrivers)

	// Drivers by status
	r.db.DB.WithContext(ctx).Model(&models.Driver{}).Where("status = ?", "pending").Count(&stats.PendingDrivers)
	r.db.DB.WithContext(ctx).Model(&models.Driver{}).Where("status = ?", "active").Count(&stats.ActiveDrivers)
	r.db.DB.WithContext(ctx).Model(&models.Driver{}).Where("status = ?", "suspended").Count(&stats.SuspendedDrivers)

	// Online drivers
	r.db.DB.WithContext(ctx).Model(&models.Driver{}).Where("is_online = ?", true).Count(&stats.OnlineDrivers)

	// Total riders
	r.db.DB.WithContext(ctx).Model(&models.User{}).Where("role = ?", "user").Count(&stats.TotalRiders)

	// Rides
	r.db.DB.WithContext(ctx).Model(&models.Ride{}).Count(&stats.TotalRides)
	r.db.DB.WithContext(ctx).Model(&models.Ride{}).Where("status = ?", "completed").Count(&stats.CompletedRides)
	r.db.DB.WithContext(ctx).Model(&models.Ride{}).Where("status = ?", "cancelled").Count(&stats.CancelledRides)

	// Total earnings — sum of all completed ride fares
	r.db.DB.WithContext(ctx).Model(&models.Ride{}).
		Where("status = ?", "completed").
		Select("COALESCE(SUM(actual_fare), 0)").
		Scan(&stats.TotalEarnings)

	return stats, nil
}

