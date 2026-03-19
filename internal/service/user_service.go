package service

import (
	"context"
	"crypto/rand"
	"drivo/internal/auth"
	"drivo/internal/jobs"
	"drivo/internal/models"
	"drivo/internal/repository"
	"drivo/internal/workers"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      *repository.UserRepo
	jwtSecret string
}

func NewUserService(repo *repository.UserRepo, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func GenerateToken(length int) (string, error) {

	max := big.NewInt(100000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", length, n), nil
}

func (svc *UserService) PreRegister(ctx context.Context, input models.UserRegisterInput) error {

	// validate inputs
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := strings.TrimSpace(strings.ToLower(input.Password))
	phone := strings.TrimSpace(strings.ToLower(input.Phone))
	name := strings.TrimSpace(strings.ToLower(input.Name))

	if email == "" || password == "" || name == "" || phone == "" {
		return errors.New("Name, Email, Phone and password must not be empty")
	}

	if len(password) < 6 {
		return errors.New("Password must be of lenght 6")
	}

	// check if user already exists
	if err := svc.repo.FindEmail(email); err != nil {
		return err
	}

	ok, err := svc.repo.ExistInRedis(email)

	if err != nil {
		return fmt.Errorf("Error with exist in redis :%v", err)
	}

	if ok == false {
		return errors.New("Email already exist, Verification already sent")
	}

	// hash password
	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return fmt.Errorf("Unable to Hash paswword:%v", err)
	}

	otp, _ := GenerateToken(5)

	u := models.PendingUser{
		Email:        email,
		HashPassword: string(hashPass),
		Phone:        phone,
		OTP:          otp,
		IsVerified:   false,
		IsActive:     false,
		Name:         name,
	}

	if err := svc.repo.StageUser(ctx, email, otp, u); err != nil {
		return errors.New("Error staging user")
	}

	workers.EmailQueue <- jobs.EmailJob{
		Type: jobs.EmailTypeOTP,
		To:   email,
		Name: name,
		OTP:  otp,
	}

	return nil

}

func (svc *UserService) VerifyUserEmail(ctx context.Context, inputOTP string, inputEmail string) (models.User, error) {
	if inputEmail == "" || inputOTP == "" {
		return models.User{}, errors.New("OTP and Email Required")
	}
	key := "pending:" + strings.TrimSpace(strings.ToLower(inputEmail))

	user, err := svc.repo.Verify(ctx, key)

	if err == redis.Nil {
		return models.User{}, errors.New("OTP expired or email not found. Please request a new code.")
	} else if err != nil {
		return models.User{}, fmt.Errorf("redis error: %v", err)
	}

	fmt.Printf("DEBUG stored OTP: [%s], input OTP: [%s]\n", user.OTP, inputOTP)

	if strings.TrimSpace(inputOTP) != strings.TrimSpace(user.OTP) {
		return models.User{}, errors.New("Invalid OTP")
	}

	if inputEmail != user.Email {
		return models.User{}, errors.New("Invalid Email")
	}

	newUser := models.User{
		Phone:        user.Phone,
		Email:        &user.Email,
		PasswordHash: user.HashPassword,
		Role:         models.RoleRider,
		IsVerified:   true,
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Name:         user.Name,
	}

	// create user

	u, err := svc.repo.Register(newUser)

	if err != nil {
		return models.User{}, fmt.Errorf("Error registering user: %v", err)
	}

	// Delete the staged driver from Redis
	if err := svc.repo.DeleteStagedUser(ctx, inputEmail); err != nil {
		return models.User{}, fmt.Errorf("Error deleting staged driver: %v", err)
	}

	workers.EmailQueue <- jobs.EmailJob{
		Type: jobs.EmailTypeWelcome,
		To:   user.Email,
		Name: user.Name,
	}

	return u, nil
}

func (svc *UserService) Login(inputPassword string, inputEmail string) (models.AuthResult, error) {

	u, err := svc.repo.FindUserEmail(inputEmail)

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(inputPassword)); err != nil {
		return models.AuthResult{}, errors.New("Email/Password incorrect")
	}

	// generate jwt token

	token, err := auth.CreateToken(svc.jwtSecret, u.ID.String(), string(u.Role))

	if err != nil {
		return models.AuthResult{}, err
	}

	return models.AuthResult{
		User:  u,
		Token: token,
	}, nil
}
