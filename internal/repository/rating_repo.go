package repository

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/models"
	"fmt"

	"github.com/google/uuid"
)

type RatingRepo struct {
	db *app.App
}

func NewRatingRepo(db *app.App) *RatingRepo {
	return &RatingRepo{
		db: db,
	}
}

func (r *RatingRepo) CreateRating(ctx context.Context, rating models.Rating) error {

	var count int64
	r.db.DB.WithContext(ctx).
		Model(&models.Rating{}).
		Where("ride_id = ? AND rater_id = ?", rating.RideID, rating.RaterID).
		Count(&count)

	if count > 0 {
		return fmt.Errorf("you have already rated this ride")
	}

	return r.db.DB.WithContext(ctx).Create(&rating).Error
}

func (r *RatingRepo) UpdateDriverRating(ctx context.Context, driverUserID uuid.UUID, avg float64) error {
	return r.db.DB.WithContext(ctx).
		Model(&models.Driver{}).
		Where("user_id = ?", driverUserID).
		Update("rating", avg).Error
}

func (r *RatingRepo) GetAverageRating(ctx context.Context, rateeID uuid.UUID) (float64, error) {
    var avg float64
    err := r.db.DB.WithContext(ctx).
        Model(&models.Rating{}).
        Where("ratee_id = ?", rateeID).
        Select("COALESCE(AVG(score), 5.0)"). 
        Scan(&avg).Error
    return avg, err
}
