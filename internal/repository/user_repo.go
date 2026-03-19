package repository

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type UserRepo struct {
	db *app.App
}

func NewUserRepo(db *app.App) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) StageUser(ctx context.Context, email string, token string, u models.PendingUser) error {

	childCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	key := "pending:" + email

	data, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return r.db.Redis.Set(childCtx, key, data, 15*time.Minute).Err()
}

func (r *UserRepo) DeleteStagedUser(ctx context.Context, email string) error {
	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	key := "pending:" + email

	return r.db.Redis.Del(childCtx, key).Err()
}

func (r *UserRepo) Verify(ctx context.Context, key string) (models.PendingUser, error) {

	childCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	var user models.PendingUser
	
	data, err := r.db.Redis.Get(childCtx, key).Bytes()
	if err != nil {
		return user, err
	}
	err = json.Unmarshal(data, &user)
	return user, err
}

func (r *UserRepo) Register(u models.User) (models.User, error) {

	result := r.db.DB.Create(&u)

	if result.Error != nil {
		return models.User{}, fmt.Errorf("error registering user:%v", result.Error)
	}

	return u, nil

}

func (r *UserRepo) ExistInRedis(email string) (bool, error) {

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

func (r *UserRepo) FindEmail(email string) error {

	var count int64

	r.db.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)

	if count > 0 {

		return errors.New("User already exist with this email, please use another one")
	}

	return nil
}

func (r *UserRepo) FindUserEmail(email string) (models.User, error) {

	var user models.User

	result := r.db.DB.Where("email = ?", email).First(&user)

	if result.Error != nil {

		return models.User{}, errors.New("User already exist with this email, please use another one")
	}

	return user, nil
}

