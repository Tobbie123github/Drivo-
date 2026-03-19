package service

import (
	"context"
	"crypto/rand"
	"drivo/internal/auth"
	"drivo/internal/jobs"
	"drivo/internal/models"
	"drivo/internal/repository"
	"drivo/internal/workers"
	"drivo/pkg/utils"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type DriverService struct {
	repo      *repository.DriverRepo
	jwtSecret string
}

func NewDriverService(repo *repository.DriverRepo, jwtSecret string) *DriverService {
	return &DriverService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func generateToken(length int) (string, error) {

	max := big.NewInt(100000)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", length, n), nil
}

func (svc *DriverService) PreRegister(ctx context.Context, input models.DriverRegisterInput) error {

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

	otp, _ := generateToken(5)

	u := models.PendingDriver{
		Email:        email,
		HashPassword: string(hashPass),
		Phone:        phone,
		OTP:          otp,
		IsVerified:   false,
		IsActive:     false,
		Name:         name,
	}

	if err := svc.repo.StageDriver(ctx, email, otp, u); err != nil {
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

func (svc *DriverService) VerifyUserEmail(ctx context.Context, inputOTP string, inputEmail string) (models.User, error) {
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
		Role:         models.RoleDriver,
		IsVerified:   false,
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Name:         user.Name,
	}

	// create user

	u, err := svc.repo.Register(newUser)

	if err != nil {
		return models.User{}, err
	}

	d := models.Driver{
		UserID: u.ID,
		Status: models.DriverPending,
	}

	_, err = svc.repo.Driver(d)

	if err != nil {
		return models.User{}, fmt.Errorf("Error registering driver: %v", err)
	}

	// Delete the staged driver from Redis
	if err := svc.repo.DeleteStagedDriver(ctx, inputEmail); err != nil {
		return models.User{}, fmt.Errorf("Error deleting staged driver: %v", err)
	}

	workers.EmailQueue <- jobs.EmailJob{
		Type: jobs.EmailTypeDriverWelcome,
		To:   user.Email,
		Name: user.Name,
	}

	return u, nil
}

func (svc *DriverService) Login(inputPassword string, inputEmail string) (models.AuthResult, error) {

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

func (svc *DriverService) CompleteProfile(ctx context.Context, userId uuid.UUID, input models.DriverProfileInput) error {

	address := strings.TrimSpace(input.Address)
	city := strings.TrimSpace(input.City)
	dob := strings.TrimSpace(input.DOB)
	country := strings.TrimSpace(input.Country)
	fullname := strings.TrimSpace(input.FullName)
	gender := strings.TrimSpace(input.Gender)
	state := strings.TrimSpace(input.State)

	if address == "" || city == "" || dob == "" || country == "" || fullname == "" || gender == "" || state == "" {
		return errors.New("All fields are required")
	}

	if input.Avatar == nil {
		return errors.New("File should not be empty")
	}

	cloud, err := utils.NewCloudinary()

	if err != nil {
		return fmt.Errorf("Issue with the cloud utils: %v", err)
	}

	var avatarUrl string
	var PublicID string

	file, err := input.Avatar.Open()

	if err != nil {
		return fmt.Errorf("failed to get file header: %v", err)
	}

	defer file.Close()

	uploadRes, err := cloud.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "avatar-image",
		UniqueFilename: api.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("Error getting result: %v", err)
	}

	avatarUrl = uploadRes.SecureURL
	PublicID = uploadRes.PublicID

	user := models.User{
		Name:      fullname,
		AvatarURL: &avatarUrl,
		PublicID:  &PublicID,
		UpdatedAt: time.Now().UTC(),
	}

	//update driver table

	dateofbirth, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return errors.New("invalid date format, use YYYY-MM-DD")
	}

	driver := models.Driver{
		DOB:            &dateofbirth,
		Gender:         &gender,
		City:           &city,
		State:          &state,
		Country:        &country,
		OnboardingStep: 2,
		Address:        &address,
	}

	if err := svc.repo.Update(user, driver, userId); err != nil {
		return err
	}

	return nil

}

func (svc *DriverService) UpdateLicence(ctx context.Context, userId uuid.UUID, input models.DriverLicence) error {

	licenseExp := strings.TrimSpace(input.LicenseExpiry)
	licenseNum := strings.TrimSpace(input.LicenseNumber)

	if licenseExp == "" || licenseNum == "" {
		return errors.New("All fields are important")
	}

	if input.LicenseImage == nil {
		return errors.New("File should not be empty")
	}

	cloud, err := utils.NewCloudinary()

	if err != nil {
		return fmt.Errorf("Issue with the cloud utils: %v", err)
	}

	var licenseUrl string

	file, err := input.LicenseImage.Open()

	if err != nil {
		return fmt.Errorf("failed to get file header: %v", err)
	}

	defer file.Close()

	uploadRes, err := cloud.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "avatar-image",
		UniqueFilename: api.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("Error getting result: %v", err)
	}

	licenseUrl = uploadRes.SecureURL

	driver := models.Driver{
		LicenseNumber:   licenseNum,
		LicenseExpiry:   licenseExp,
		LicenseVerified: false,
		LicenseImage:    &licenseUrl,
		OnboardingStep:  3,
	}

	if err := svc.repo.UpdateDriverLicense(driver, userId); err != nil {
		return err
	}

	return nil

}

func (svc *DriverService) Vehicle(ctx context.Context, userId uuid.UUID, input models.VehicleInput) error {

	make := strings.TrimSpace(input.Make)
	model := strings.TrimSpace(input.Model)
	year := input.Year
	color := strings.TrimSpace(input.Color)
	plateNum := strings.TrimSpace(input.PlateNumber)
	category := strings.TrimSpace(input.Category)
	seats := input.Seats

	driver, err := svc.repo.GetDriver(userId)
	if err != nil {
		return err
	}

	if make == "" || model == "" || color == "" || plateNum == "" || category == "" || year == 0 || seats == 0 {
		return errors.New("All fields are important")
	}

	if input.VehicleImage == nil {
		return errors.New("File should not be empty")
	}

	cloud, err := utils.NewCloudinary()

	if err != nil {
		return fmt.Errorf("Issue with the cloud utils: %v", err)
	}

	var vehicleUrl string

	file, err := input.VehicleImage.Open()

	if err != nil {
		return fmt.Errorf("failed to get file header: %v", err)
	}

	defer file.Close()

	uploadRes, err := cloud.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "vehicle-image",
		UniqueFilename: api.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("Error getting result: %v", err)
	}

	vehicleUrl = uploadRes.SecureURL

	v := models.Vehicle{
		DriverID:     driver.ID,
		Make:         make,
		Model:        model,
		Year:         year,
		Color:        color,
		PlateNumber:  plateNum,
		Category:     category,
		Seats:        seats,
		VehicleImage: &vehicleUrl,
	}

	if err := svc.repo.AddVehicle(v); err != nil {
		return err
	}

	driver = models.Driver{
		OnboardingStep: 4,
	}

	if err := svc.repo.UpdateDriverLicense(driver, userId); err != nil {
		return err
	}

	return nil

}

func (svc *DriverService) ProofofProfile(ctx context.Context, userId uuid.UUID, input models.DocumentUploadInput) error {

	if input.NationalIdImage == nil || input.SelfieImage == nil || input.ProofOfAddress == nil {
		return errors.New("All files are required")
	}

	cloud, err := utils.NewCloudinary()

	if err != nil {
		return fmt.Errorf("Issue with the cloud utils: %v", err)
	}

	// Upload National ID Image

	var nationalIdUrl string

	file, err := input.NationalIdImage.Open()

	if err != nil {
		return fmt.Errorf("failed to get file header: %v", err)
	}

	defer file.Close()

	uploadRes, err := cloud.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "driver-documents",
		UniqueFilename: api.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("Error getting result: %v", err)
	}

	nationalIdUrl = uploadRes.SecureURL

	// Upload Selfie Image

	var selfieUrl string

	file1, err := input.SelfieImage.Open()

	if err != nil {
		return fmt.Errorf("failed to get file header: %v", err)
	}

	defer file.Close()

	uploadRes1, err := cloud.Upload.Upload(ctx, file1, uploader.UploadParams{
		Folder:         "driver-documents",
		UniqueFilename: api.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("Error getting result: %v", err)
	}

	selfieUrl = uploadRes1.SecureURL

	// Upload Proof of Address

	var proofOfAddressUrl string

	file2, err := input.ProofOfAddress.Open()

	if err != nil {
		return fmt.Errorf("failed to get file header: %v", err)
	}

	defer file.Close()

	uploadRes2, err := cloud.Upload.Upload(ctx, file2, uploader.UploadParams{
		Folder:         "driver-documents",
		UniqueFilename: api.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("Error getting result: %v", err)
	}

	proofOfAddressUrl = uploadRes2.SecureURL

	driver := models.Driver{
		NationalIdImage: &nationalIdUrl,
		SelfieImage:     &selfieUrl,
		ProofOfAddress:  &proofOfAddressUrl,
		OnboardingStep:  5,
	}

	if err := svc.repo.UpdateDriverLicense(driver, userId); err != nil {
		return err
	}

	return nil

}

func (svc *DriverService) AgreeTerms(userId uuid.UUID) error {

	driver := models.Driver{
		AgreeTerms:            true,
		OnboardingStep:        6,
		IsOnboardingCompleted: true,
		Status:                "pending",
	}

	return svc.repo.UpdateDriver(driver, userId)

}

func (svc *DriverService) GetDriverProfile(userId uuid.UUID) (models.Driver, error) {

	driver, err := svc.repo.GetDriver(userId)

	if err != nil {
		return models.Driver{}, err
	}

	return driver, nil

}

func (svc *DriverService) OnlineStatus(input models.StatusUpdate, userId uuid.UUID) error {
	return svc.repo.UpdateOnlineStatus(userId, input.IsOnline)
}

func (svc *DriverService) UpdateLocation(ctx context.Context, driverID uuid.UUID, lat float64, lng float64) error {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return errors.New("invalid coordinates")
	}

	if err := svc.repo.SaveLocationToRedis(ctx, driverID, lat, lng); err != nil {
		return fmt.Errorf("failed to cache location: %v", err)
	}

	return nil
}
