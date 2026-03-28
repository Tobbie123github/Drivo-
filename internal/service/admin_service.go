package service

import (
	"context"
	"drivo/internal/models"
	"drivo/internal/repository"
	"fmt"

	"github.com/google/uuid"
)

type AdminService struct {
	adminRepo  *repository.AdminRepo
	driverRepo *repository.DriverRepo
}

func NewAdminService(adminRepo *repository.AdminRepo, driverRepo *repository.DriverRepo) *AdminService {
	return &AdminService{adminRepo: adminRepo, driverRepo: driverRepo}
}

// ---------------------------DRIVERS------------------------//

func (s *AdminService) VerifyDriverIdentity(ctx context.Context, driverID uuid.UUID) error {
	return s.adminRepo.VerifyDriverIdentity(ctx, driverID)
}

func (s *AdminService) VerifyDriverVehicle(ctx context.Context, driverID uuid.UUID) error {
	return s.adminRepo.VerifyDriverVehicle(ctx, driverID)
}

func (s *AdminService) VerifyDriverLicense(ctx context.Context, driverID uuid.UUID) error {
	return s.adminRepo.VerifyDriverLicense(ctx, driverID)
}

func (s *AdminService) GetAllDrivers(ctx context.Context, status string) ([]models.Driver, error) {
	return s.adminRepo.GetAllDrivers(ctx, status)
}

func (s *AdminService) ApproveDriver(ctx context.Context, driverID uuid.UUID) (models.Driver, error) {

	drivers, err := s.adminRepo.GetAllDrivers(ctx, "")
	if err != nil {
		return models.Driver{}, err
	}

	var driver models.Driver
	for _, d := range drivers {
		if d.ID == driverID {
			driver = d
			break
		}
	}

	if !driver.IsIdentityVerified {
		return models.Driver{}, fmt.Errorf("cannot approve driver: identity not verified")
	}

	if !driver.IsVehicleVerified {
		return models.Driver{}, fmt.Errorf("cannot approve driver: vehicle not verified")
	}
	if !driver.LicenseVerified {
		return models.Driver{}, fmt.Errorf("cannot approve driver: license not verified")
	}

	if err := s.adminRepo.UpdateDriverStatus(ctx, driverID, models.DriverActive); err != nil {
		return models.Driver{}, err
	}


	d, err := s.driverRepo.GetUserByDriverID(driverID)

	if err != nil {
		return models.Driver{}, nil 
	}

	return d,  nil
}

func (s *AdminService) RejectDriver(ctx context.Context, driverID uuid.UUID) error {
	return s.adminRepo.UpdateDriverStatus(ctx, driverID, models.DriverBanned)
}

func (s *AdminService) SuspendDriver(ctx context.Context, driverID uuid.UUID) error {
	return s.adminRepo.UpdateDriverStatus(ctx, driverID, models.DriverSuspended)
}

func (s *AdminService) BanDriver(ctx context.Context, driverID uuid.UUID) error {
	return s.adminRepo.UpdateDriverStatus(ctx, driverID, models.DriverBanned)
}

func (s *AdminService) GetDriverByID(ctx context.Context, driverID uuid.UUID) (models.Driver, error) {
	drivers, err := s.adminRepo.GetAllDrivers(ctx, "")
	if err != nil {
		return models.Driver{}, err
	}

	for _, d := range drivers {
		if d.ID == driverID {
			return d, nil
		}
	}

	return models.Driver{}, fmt.Errorf("driver not found")
}

// ------------------RIDERS-----------------------
func (s *AdminService) GetAllRiders(ctx context.Context) ([]models.User, error) {
	return s.adminRepo.GetAllRiders(ctx)
}

func (s *AdminService) GetAllRides(ctx context.Context, status string) ([]models.Ride, error) {
	return s.adminRepo.GetAllRides(ctx, status)
}

func (s *AdminService) GetDashboardStats(ctx context.Context) (repository.DashboardStats, error) {
	return s.adminRepo.GetDashboardStats(ctx)
}
