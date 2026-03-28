package repository

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/models"
	"time"

	"github.com/google/uuid"
)

type RecurringRepo struct {
	db *app.App
}

func NewRecurringRepo(db *app.App) *RecurringRepo {
	return &RecurringRepo{db: db}
}

func (r *RecurringRepo) Create(ctx context.Context, ride *models.RecurringRide) error {
	return r.db.DB.WithContext(ctx).Create(ride).Error
}

func (r *RecurringRepo) GetByRiderID(ctx context.Context, riderID uuid.UUID) ([]models.RecurringRide, error) {
	var rides []models.RecurringRide
	err := r.db.DB.WithContext(ctx).
		Where("rider_id = ? AND is_active = true", riderID).
		Find(&rides).Error
	return rides, err
}

func (r *RecurringRepo) GetAllActive(ctx context.Context) ([]models.RecurringRide, error) {
	var rides []models.RecurringRide
	err := r.db.DB.WithContext(ctx).
		Where("is_active = true").
		Find(&rides).Error
	return rides, err
}

func (r *RecurringRepo) Update(ctx context.Context, ride *models.RecurringRide) error {
	return r.db.DB.WithContext(ctx).Save(ride).Error
}

func (r *RecurringRepo) Deactivate(ctx context.Context, id uuid.UUID, riderID uuid.UUID) error {
	return r.db.DB.WithContext(ctx).
		Model(&models.RecurringRide{}).
		Where("id = ? AND rider_id = ?", id, riderID).
		Update("is_active", false).Error
}

// AlreadyBookedForDate — prevents duplicate if cron runs twice
func (r *RecurringRepo) AlreadyBookedForDate(ctx context.Context, recurringID uuid.UUID, date time.Time) (bool, error) {
	var count int64
	dateStr := date.Format("2006-01-02")
	err := r.db.DB.WithContext(ctx).
		Model(&models.Ride{}).
		Where("recurring_ride_id = ?", recurringID).
		Where("DATE(scheduled_at) = ?", dateStr).
		Count(&count).Error
	return count > 0, err
}
