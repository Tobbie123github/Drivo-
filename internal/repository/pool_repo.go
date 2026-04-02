package repository

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/models"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PoolRepo struct {
	db *app.App
}

func NewPoolRepo(db *app.App) *PoolRepo {
	return &PoolRepo{db: db}
}

func (r *PoolRepo) CreatePool(ctx context.Context, pool *models.PoolGroup) error {
	return r.db.DB.WithContext(ctx).Create(pool).Error
}

func (r *PoolRepo) GetOpenPools(ctx context.Context) ([]models.PoolGroup, error) {
	var pools []models.PoolGroup
	err := r.db.DB.WithContext(ctx).Where("status = ?", models.PoolStatusOpen).Where("current_size < max_riders").Preload("Rides").Find(&pools).Error
	if err != nil {
		return nil, err
	}

	return pools, nil
}

func (r *PoolRepo) GetPoolByID(ctx context.Context, id uuid.UUID) (*models.PoolGroup, error) {
	var pool models.PoolGroup
	err := r.db.DB.WithContext(ctx).
		Where("id = ?", id).
		Preload("Rides").
		First(&pool).Error
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

func (r *PoolRepo) UpdatePool(ctx context.Context, pool *models.PoolGroup) error {
	return r.db.DB.WithContext(ctx).Save(pool).Error
}

func (r *PoolRepo) IncrementPoolSize(ctx context.Context, poolID uuid.UUID) error {
	return r.db.DB.WithContext(ctx).
		Model(&models.PoolGroup{}).
		Where("id = ?", poolID).
		UpdateColumn("current_size", gorm.Expr("current_size + 1")).Error
}

func (r *PoolRepo) FindCompatiblePool(ctx context.Context, originLat, originLng, destLat, destLng float64) (*models.PoolGroup, error) {
	pools, err := r.GetOpenPools(ctx)
	if err != nil {
		return nil, err
	}

	for _, pool := range pools {
		originDist := haversineKm(originLat, originLng, pool.OriginLat, pool.OriginLng)
		destDist := haversineKm(destLat, destLng, pool.DestLat, pool.DestLng)

		// Origin within 1.5km and destination within 2.5km = compatible
		if originDist <= 1.5 && destDist <= 2.5 {
			return &pool, nil
		}
	}
	return nil, nil
}

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func (r *PoolRepo) CreatePoolRide(ctx context.Context, ride *models.Ride) error {
	return r.db.DB.WithContext(ctx).Create(ride).Error
}

// func (r *PoolRepo) AssignDriver(ctx context.Context, poolID uuid.UUID, driverID uuid.UUID) error {
// 	return r.db.DB.WithContext(ctx).
// 		Model(&models.PoolGroup{}).
// 		Where("id = ?", poolID).
// 		Updates(map[string]interface{}{
// 			"driver_id": driverID,
// 			"status":    models.PoolStatusOpen,
// 		}).Error
// }
