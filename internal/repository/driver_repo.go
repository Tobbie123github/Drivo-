package repository

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DriverRepo struct {
	db *app.App
}

func NewDriverRepo(db *app.App) *DriverRepo {
	return &DriverRepo{
		db: db,
	}
}

func (r *DriverRepo) StageDriver(ctx context.Context, email string, token string, u models.PendingDriver) error {

	childCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	key := "pending:" + email

	data, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return r.db.Redis.Set(childCtx, key, data, 15*time.Minute).Err()
}

func (r *DriverRepo) DeleteStagedDriver(ctx context.Context, email string) error {
	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	key := "pending:" + email

	return r.db.Redis.Del(childCtx, key).Err()
}

func (r *DriverRepo) Verify(ctx context.Context, key string) (models.PendingDriver, error) {

	childCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	var user models.PendingDriver

	data, err := r.db.Redis.Get(childCtx, key).Bytes()
	if err != nil {
		return user, err
	}
	err = json.Unmarshal(data, &user)
	return user, err
}

func (r *DriverRepo) Register(u models.User) (models.User, error) {

	result := r.db.DB.Create(&u)

	if result.Error != nil {
		return models.User{}, fmt.Errorf("error registering user:%v", result.Error)
	}

	return u, nil

}

func (r *DriverRepo) ExistInRedis(email string) (bool, error) {

	ctx := context.Background()
	key := "pending:" + email

	exists, err := r.db.Redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if exists > 0 {
		return false, nil
	}

	return true, nil
}

func (r *DriverRepo) FindEmail(email string) error {

	var count int64

	r.db.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)

	if count > 0 {

		return errors.New("User already exist with this email, please use another one")
	}

	return nil
}

func (r *DriverRepo) FindPhone(phone string) error {

	var count int64

	result := r.db.DB.Model(&models.User{}).Where("phone = ?", phone).Count(&count)

	if result.Error != nil {
		return result.Error
	}

	if count > 0 {
		return errors.New("User already exist with this phone, please use another one")
	}

	return nil
}

func (r *DriverRepo) FindUserEmail(email string) (models.User, error) {

	var user models.User

	result := r.db.DB.Where("email = ?", email).First(&user)

	if result.Error != nil {

		return models.User{}, errors.New("User already exist with this email, please use another one")
	}

	return user, nil
}

func (r *DriverRepo) Driver(d models.Driver) (models.Driver, error) {

	result := r.db.DB.Create(&d)

	if result.Error != nil {
		return models.Driver{}, fmt.Errorf("error registering user:%v", result.Error)
	}

	return d, nil
}

func (r *DriverRepo) Update(user models.User, driver models.Driver, userId uuid.UUID) error {

	tx := r.db.DB.Begin()

	if err := tx.Model(&models.User{}).
		Where("id = ?", userId).
		Updates(user).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.Driver{}).
		Where("user_id = ?", userId).
		Updates(driver).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error

}

func (r *DriverRepo) UpdateDriverLicense(driver models.Driver, userId uuid.UUID) error {

	tx := r.db.DB.Begin()

	if err := tx.Model(&models.Driver{}).
		Where("user_id = ?", userId).
		Updates(driver).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error

}

func (r *DriverRepo) GetDriver(ctx context.Context, userId uuid.UUID) (models.Driver, error) {
	var driver models.Driver

	err := r.db.DB.WithContext(ctx).
		Preload("Vehicles").
		Where("user_id = ?", userId).
		First(&driver).Error

	return driver, err
}

func (r *DriverRepo) GetDriverByUserID(userID uuid.UUID) (models.Driver, error) {
	var driver models.Driver
	if err := r.db.DB.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		return models.Driver{}, fmt.Errorf("driver not found: %v", err)
	}
	return driver, nil
}

func (r *DriverRepo) GetDriverByID(driverID uuid.UUID) (models.Driver, error) {
	var driver models.Driver
	if err := r.db.DB.Where("id = ?", driverID).First(&driver).Error; err != nil {
		return models.Driver{}, fmt.Errorf("driver not found: %v", err)
	}
	return driver, nil
}

func (r *DriverRepo) GetUserByDriverID(driverID uuid.UUID) (models.Driver, error) {
	var driver models.Driver

	err := r.db.DB.Preload("User").
		First(&driver, "id = ?", driverID).Error
	if err != nil {
		return models.Driver{}, fmt.Errorf("user not found: %v", err)
	}
	return driver, nil
}

func (r *DriverRepo) GetDriverVehicle(driverID uuid.UUID) (models.Vehicle, error) {
	var vehicle models.Vehicle

	err := r.db.DB.
		Where("driver_id = ?", driverID).
		First(&vehicle).Error

	return vehicle, err
}

func (r *DriverRepo) UpdateDriver(driver models.Driver, userId uuid.UUID) error {

	return r.db.DB.Model(&models.Driver{}).
		Where("user_id = ?", userId).
		Updates(driver).
		Error
}

func (r *DriverRepo) UpdateOnlineStatus(userId uuid.UUID, isOnline bool) error {
	return r.db.DB.Model(&models.Driver{}).
		Where("user_id = ?", userId).
		Update("is_online", isOnline).
		Error
}

func (r *DriverRepo) AddVehicle(vehicle models.Vehicle) error {

	return r.db.DB.Create(&vehicle).Error

}

func (r *DriverRepo) SaveLocationToRedis(ctx context.Context, driverID uuid.UUID, lat, lng float64) error {

	data := models.LocationData{
		Latitude:  lat,
		Longitude: lng,
		UpdatedAt: time.Now().UTC(),
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %v", err)
	}

	key := fmt.Sprintf("driver:location:%s", driverID.String())

	return r.db.Redis.Set(ctx, key, bytes, 5*time.Minute).Err()
}

func (r *DriverRepo) GetLocationFromRedis(ctx context.Context, driverID uuid.UUID) (*models.LocationData, error) {
	key := fmt.Sprintf("driver:location:%s", driverID.String())

	val, err := r.db.Redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var loc models.LocationData

	if err := json.Unmarshal([]byte(val), &loc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location: %v", err)
	}

	return &loc, nil
}

func (r *DriverRepo) IncreaseDriverTrips(ctx context.Context, driverID uuid.UUID) error {

	return r.db.DB.WithContext(ctx).Model(&models.Driver{}).Where("id = ?", driverID).Update("total_trips", gorm.Expr("total_trips + 1")).Error
}

func (r *DriverRepo) IncrementCancellationRate(ctx context.Context, driverID uuid.UUID) error {
	return r.db.DB.WithContext(ctx).
		Model(&models.Driver{}).
		Where("id = ?", driverID).
		UpdateColumn("cancellation_rate", gorm.Expr("cancellation_rate + 1")).
		Error
}

func (r *DriverRepo) SaveRiderLocationToRedis(ctx context.Context, riderID uuid.UUID, lat, lng float64) error {

	data := models.LocationData{
		Latitude:  lat,
		Longitude: lng,
		UpdatedAt: time.Now().UTC(),
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %v", err)
	}

	key := fmt.Sprintf("rider:location:%s", riderID.String())

	return r.db.Redis.Set(ctx, key, bytes, 5*time.Minute).Err()
}

// will fix, send broadcast pool to online riders and location must match
func (r *DriverRepo) GetRiderLocationFromRedis(ctx context.Context, riderID uuid.UUID) (*models.LocationData, error) {
	key := fmt.Sprintf("rider:location:%s", riderID.String())

	val, err := r.db.Redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var loc models.LocationData

	if err := json.Unmarshal([]byte(val), &loc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location: %v", err)
	}

	return &loc, nil
}

func (r *DriverRepo) GetNearbyRiderIDs(ctx context.Context, lat, lng, radiusKm float64, onlineRiderIDs []uuid.UUID) ([]uuid.UUID, error) {
	var nearby []uuid.UUID

	for _, riderID := range onlineRiderIDs {
		loc, err := r.GetRiderLocationFromRedis(ctx, riderID)
		if err != nil {
			continue
		}

		dist := haversineKmDriver(lat, lng, loc.Latitude, loc.Longitude)
		if dist <= radiusKm {
			nearby = append(nearby, riderID)
		}
	}

	return nearby, nil
}

func haversineKmDriver(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}


func (r *DriverRepo) UpdatePassword(userID uuid.UUID, newPassword string) error {

	return r.db.DB.Model(&models.User{}).Where("id = ?", userID).Update("password_hash", newPassword).Error

}