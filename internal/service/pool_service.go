package service

import (
	"context"
	"drivo/internal/models"
	"drivo/internal/repository"
	"drivo/internal/ws"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/google/uuid"
)

type PoolService struct {
	poolRepo   *repository.PoolRepo
	rideRepo   *repository.RideRepo
	riderHub   *ws.RiderHub
	driverHub  *ws.Hub
	rideSvc    *RideService
	driverRepo *repository.DriverRepo
}

func NewPoolService(
	poolRepo *repository.PoolRepo,
	rideRepo *repository.RideRepo,
	riderHub *ws.RiderHub,
	driverHub *ws.Hub,
	rideSvc *RideService,
	driverRepo *repository.DriverRepo,

) *PoolService {
	return &PoolService{
		poolRepo:   poolRepo,
		rideRepo:   rideRepo,
		riderHub:   riderHub,
		driverHub:  driverHub,
		rideSvc:    rideSvc,
		driverRepo: driverRepo,
	}
}

type JoinPoolInput struct {
	RiderID     uuid.UUID
	PickupLat   float64
	PickupLng   float64
	DropoffLat  float64
	DropoffLng  float64
	PickupAddr  string
	DropoffAddr string
}
type FindPoolInput struct {
	RiderID    uuid.UUID
	PickupLat  float64
	PickupLng  float64
	DropoffLat float64
	DropoffLng float64
}

type DriverCreatePoolInput struct {
	DriverID    uuid.UUID
	PickupLat   float64
	PickupLng   float64
	DropoffLat  float64
	DropoffLng  float64
	PickupAddr  string
	DropoffAddr string
}

type PoolCheckResult struct {
	HasPool       bool              `json:"has_pool"`
	Pool          *models.PoolGroup `json:"pool,omitempty"`
	EstimatedFare float64           `json:"estimated_fare"`
	SoloFare      float64           `json:"solo_fare"`
	Savings       float64           `json:"savings"`
	RidersInPool  int               `json:"riders_in_pool"`
}

func (s *PoolService) recalculateFares(ctx context.Context, poolID uuid.UUID, newFarePerHead float64) {
	s.rideRepo.UpdatePoolRidesFare(ctx, poolID, newFarePerHead)
}

func (s *PoolService) notifyDriver(driverID uuid.UUID, msg ws.Message) {
	bytes, _ := json.Marshal(msg)
	s.driverHub.SendToDriver(driverID, bytes)
}

func (s *PoolService) notifyRider(riderID uuid.UUID, msg ws.Message) {
	bytes, _ := json.Marshal(msg)
	s.riderHub.SendToRider(riderID, bytes)
}

func (s *PoolService) notifyPoolRiders(ctx context.Context, poolID uuid.UUID, msg ws.Message) {
	rides, _ := s.rideRepo.GetRidesByPoolID(ctx, poolID)
	bytes, _ := json.Marshal(msg)
	for _, ride := range rides {

		sent := s.riderHub.SendToRider(ride.RiderID, bytes)
		fmt.Printf("[notifyAllPoolRiders] rider %s notified: %v\n", ride.RiderID, sent)

	}
}

func calculatePoolFare(pickupLat, pickupLng, dropoffLat, dropoffLng, multiplier float64) float64 {
	const R = 6371
	dLat := (dropoffLat - pickupLat) * math.Pi / 180
	dLng := (dropoffLng - pickupLng) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(pickupLat*math.Pi/180)*math.Cos(dropoffLat*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	dist := R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	estimatedMinutes := (dist / 30.0) * 60
	fare := 500.0 + (dist * 150.0) + (estimatedMinutes * 20.0)
	fare = math.Round(fare/50) * 50
	return fare * multiplier
}

func (s *PoolService) CreatePool(ctx context.Context, driverID uuid.UUID, input DriverCreatePoolInput) (*models.PoolGroup, error) {

	soloFare := calculatePoolFare(input.PickupLat, input.PickupLng, input.DropoffLat, input.DropoffLng, 1.0)
	initialFare := soloFare * 0.7

	pool := &models.PoolGroup{
		Status:      models.PoolStatusOpen,
		DriverID:    &driverID,
		MaxRiders:   3,
		CurrentSize: 0,
		BaseFare:    soloFare,
		FarePerHead: initialFare,
		OriginLat:   input.PickupLat,
		OriginLng:   input.PickupLng,
		DestLat:     input.DropoffLat,
		DestLng:     input.DropoffLng,
	}
	if err := s.poolRepo.CreatePool(ctx, pool); err != nil {
		return nil, err
	}

	onlineRiderIDs := s.riderHub.GetOnlineRiderIDs()

	nearbyRiderIDs, err := s.driverRepo.GetNearbyRiderIDs(ctx, input.PickupLat, input.PickupLng, 5.0, onlineRiderIDs)

	if err != nil {
		fmt.Printf("createpool:= error finding nearby riders: %v\n", err)
	}

	// broadcast to some riders withing this location

	for _, riderID := range nearbyRiderIDs {
		s.notifyRider(riderID, ws.Message{
			Type: ws.MessageTypePoolAvailable,
			Payload: map[string]interface{}{
				"pool_id":         pool.ID,
				"origin_lat":      pool.OriginLat,
				"origin_lng":      pool.OriginLng,
				"dest_lat":        pool.DestLat,
				"dest_lng":        pool.DestLng,
				"fare_per_head":   pool.FarePerHead,
				"current_size":    pool.CurrentSize,
				"max_riders":      pool.MaxRiders,
				"pickup_address":  input.PickupAddr,
				"dropoff_address": input.DropoffAddr,
				"message":         "A pool ride is available near you!",
			},
		})
	}

	return pool, nil

}

func (s *PoolService) FindAvailablePools(ctx context.Context, input FindPoolInput) (*PoolCheckResult, error) {

	soloFare := calculatePoolFare(input.PickupLat, input.PickupLng, input.DropoffLat, input.DropoffLng, 1.0)
	pool, err := s.poolRepo.FindCompatiblePool(ctx, input.PickupLat, input.PickupLng, input.DropoffLat, input.DropoffLng)

	if err != nil {
		return nil, err
	}

	if pool == nil {
		return &PoolCheckResult{
			HasPool:       false,
			EstimatedFare: soloFare,
			SoloFare:      soloFare,
			Savings:       0,
			RidersInPool:  0,
		}, nil
	}

	newSize := pool.CurrentSize + 1

	farePerHead := soloFare / float64(newSize) * 0.05

	return &PoolCheckResult{
		HasPool:       true,
		Pool:          pool,
		EstimatedFare: farePerHead,
		SoloFare:      soloFare,
		Savings:       soloFare - farePerHead,
		RidersInPool:  pool.CurrentSize,
	}, nil

}

func (s *PoolService) JoinPool(ctx context.Context, poolID uuid.UUID, input JoinPoolInput) (*models.Ride, error) {
	pool, err := s.poolRepo.GetPoolByID(ctx, poolID)
	if err != nil {
		return nil, fmt.Errorf("pool not found")
	}

	if pool.Status != models.PoolStatusOpen || pool.CurrentSize >= pool.MaxRiders {
		return nil, fmt.Errorf("pool no longer available")
	}

	rides, err := s.rideRepo.GetRidesByPoolID(ctx, pool.ID)

	for _, ride := range rides {

		if ride.RiderID == input.RiderID {
			return nil, fmt.Errorf("Already in the pool")
		}

	}

	comPool, err := s.poolRepo.FindCompatiblePool(ctx, input.PickupLat, input.PickupLng, input.DropoffLat, input.DropoffLng)

	if err != nil {
		return nil, err
	}

	if comPool == nil {
		return nil, fmt.Errorf("Not within location")
	}

	newSize := pool.CurrentSize + 1
	totalFare := calculatePoolFare(pool.OriginLat, pool.OriginLng, pool.DestLat, pool.DestLng, 1.0)
	farePerHead := totalFare / float64(newSize) * 0.85

	ride := &models.Ride{
		RiderID:        input.RiderID,
		PickupLat:      input.PickupLat,
		PickupLng:      input.PickupLng,
		DropoffLat:     input.DropoffLat,
		DropoffLng:     input.DropoffLng,
		PickupAddress:  input.PickupAddr,
		DropoffAddress: input.DropoffAddr,
		EstimatedFare:  farePerHead,
		PoolFarePaid:   farePerHead,
		RideMode:       "pool",
		PoolGroupID:    &pool.ID,
		Status:         models.RideStatusAccepted,
		DriverID:       pool.DriverID,
	}

	if err = s.poolRepo.CreatePoolRide(ctx, ride); err != nil {
		return nil, err
	}

	if err := s.poolRepo.IncrementPoolSize(ctx, poolID); err != nil {
		return nil, err
	}

	if newSize >= pool.MaxRiders {
		pool.Status = models.PoolStatusFull
		s.poolRepo.UpdatePool(ctx, pool)
	}

	go s.recalculateFares(ctx, pool.ID, farePerHead)

	driver, err := s.driverRepo.GetDriverByID(*pool.DriverID)

	if err != nil {
		log.Printf("failed to get driver: %v", err)
		return ride, nil
	}

	fmt.Printf("driver to send is : %v", driver.UserID)

	if pool.DriverID != nil {
		s.notifyDriver(driver.UserID, ws.Message{
			Type: ws.MessageTypePoolRiderJoined,
			Payload: map[string]interface{}{
				"pool_id":        pool.ID,
				"riders_count":   newSize,
				"new_rider":      pool.DriverID,
				"pickup_address": input.PickupAddr,
			},
		})
	}

	s.notifyPoolRiders(ctx, pool.ID, ws.Message{
		Type: ws.MessageTypePoolUpdated,
		Payload: map[string]interface{}{
			"riders_count": newSize,
			"new_fare":     farePerHead,
			"message":      fmt.Sprintf("Another rider joined! Your fare is now ₦%.0f", farePerHead),
		},
	})

	return ride, nil

}

func (s *PoolService) StartPoolTrip(ctx context.Context, poolID uuid.UUID, driverUserID uuid.UUID) error {

	// get driver ID by user id

	driver, err := s.driverRepo.GetDriverByUserID(driverUserID)

	if err != nil {
		return fmt.Errorf("Error getting driver: %v", err)
	}

	// get pool by pool id

	pool, err := s.poolRepo.GetPoolByID(ctx, poolID)

	if err != nil {
		return fmt.Errorf("Error getting pool by id: %v", err)
	}

	if pool.Status == models.PoolStatusOpen || pool.Status == models.PoolStatusFull {
		pool.Status = models.PoolStatusActive
		s.poolRepo.UpdatePool(ctx, pool)
	}

	// check if the driver is the owner of the pool
	if pool.DriverID == nil || *pool.DriverID != driver.ID {
		return errors.New("you are not assigned to this ride")
	}

	// get rides that has thesame pool id

	rides, err := s.rideRepo.GetRidesByPoolID(ctx, poolID)

	if err != nil {
		return fmt.Errorf("Error getting pool rides: %v", err)
	}

	if len(rides) == 0 {
		return errors.New("no riders in this pool")
	}

	for _, ride := range rides {

		if ride.Status != models.RideStatusAccepted {
			return errors.New("ride is not in accepted state")
		}

		if err := s.rideRepo.UpdateRideStatus(ctx, ride.ID, models.RideStatusOngoing, nil); err != nil {
			return fmt.Errorf("error updating ride status: %v", err)
		}

		s.notifyPoolRiders(ctx, poolID, ws.Message{
			Type: ws.MessageTypePoolRideStarted,
			Payload: map[string]interface{}{
				"pool_id": poolID.String(),
				"message": "Your trip has started.",
			},
		})

	}

	return nil

}

func (s *PoolService) CancelPoolTrip(ctx context.Context, poolID uuid.UUID, driverUserID uuid.UUID) error {

	driver, err := s.driverRepo.GetDriverByUserID(driverUserID)

	if err != nil {
		return fmt.Errorf("Error getting driver: %v", err)
	}

	// get pool by pool id
	pool, err := s.poolRepo.GetPoolByID(ctx, poolID)

	if err != nil {
		return fmt.Errorf("error getting pool: %v", err)
	}

	if pool.Status == models.PoolStatusActive || pool.Status == models.PoolStatusCompleted {
		return fmt.Errorf("An active or completed pool trip cant be cancelled")
	}

	// check if the driver is the owner of the pool
	if pool.DriverID == nil || *pool.DriverID != driver.ID {
		return errors.New("you are not assigned to this ride")
	}

	rides, err := s.rideRepo.GetRidesByPoolID(ctx, poolID)

	if err != nil {
		return fmt.Errorf("error getting pool rides: %v", err)
	}

	for _, ride := range rides {
		if ride.Status == models.RideStatusOngoing || ride.Status == models.RideStatusCompleted {
			return fmt.Errorf("Ongoing pool ride cant be cancelled")
		}

		if err := s.rideRepo.UpdateRideStatus(ctx, ride.ID, models.RideStatusCancelled, nil); err != nil {
			return fmt.Errorf("failed to cancel ride %s: %v", ride.ID, err)
		}
	}

	pool.Status = models.PoolStatusCancelled
	if err := s.poolRepo.UpdatePool(ctx, pool); err != nil {
		return fmt.Errorf("failed to cancel pool: %v", err)
	}

	go s.notifyPoolRiders(ctx, poolID, ws.Message{
		Type: ws.MessageTypePoolCancelled,
		Payload: map[string]interface{}{
			"message": "Driver cancelled the pool trip",
		},
	})

	return nil

}

func (s *PoolService) CompleteTrip(ctx context.Context, poolID uuid.UUID, driverUserID uuid.UUID) error {
	driver, err := s.driverRepo.GetDriverByUserID(driverUserID)

	if err != nil {
		return fmt.Errorf("Error getting driver: %v", err)
	}

	// get pool by pool id
	pool, err := s.poolRepo.GetPoolByID(ctx, poolID)

	if err != nil {
		return fmt.Errorf("error getting pool: %v", err)
	}

	if pool.Status != models.PoolStatusActive {
		return fmt.Errorf("only active trip can be completed")
	}

	// check if the driver is the owner of the pool
	if pool.DriverID == nil || *pool.DriverID != driver.ID {
		return errors.New("you are not assigned to this ride")
	}

	rides, err := s.rideRepo.GetRidesByPoolID(ctx, poolID)

	for _, ride := range rides {
		if ride.Status != models.RideStatusOngoing {
			return fmt.Errorf("Only ongoing pool ride can be completed")
		}

		if err := s.rideRepo.UpdateRideStatus(ctx, ride.ID, models.RideStatusCompleted, nil); err != nil {
			return fmt.Errorf("failed to cancel ride %s: %v", ride.ID, err)
		}
	}

	pool.Status = models.PoolStatusCompleted
	if err := s.poolRepo.UpdatePool(ctx, pool); err != nil {
		return fmt.Errorf("failed to complete pool: %v", err)
	}

	go s.notifyPoolRiders(ctx, poolID, ws.Message{
		Type: ws.MessageTypePoolRideCompleted,
		Payload: map[string]interface{}{
			"message": "Driver completed the pool trip",
		},
	})

	return nil

}
