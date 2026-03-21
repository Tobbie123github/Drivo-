package server

import (
	"drivo/internal/app"
	"drivo/internal/config"
	"drivo/internal/handler"
	"drivo/internal/middleware"
	"drivo/internal/repository"
	"drivo/internal/service"
	"drivo/internal/workers"
	"drivo/internal/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(a *app.App, cfg config.Config) *gin.Engine {

	r := gin.New()

	hub := ws.NewHub()
	go hub.Run()

	riderHub := ws.NewRiderHub()
	go riderHub.Run()

	r.Use(gin.Logger())

	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
			"Upgrade",
			"Keep-Alive",
			"Connection",
			"Sec-WebSocket-Key",
			"Sec-WebSocket-Version",
		},
		AllowCredentials: true,
	}))

	r.GET("/health", health)

	userRepo := repository.NewUserRepo(a)
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	userHandler := handler.NewUserHandler(userService)

	driverRepo := repository.NewDriverRepo(a)
	driverService := service.NewDriverService(driverRepo, cfg.JWTSecret)
	driverHandler := handler.NewDriverHandler(driverService)

	rideRepo := repository.NewRideRepo(a)
	rideService := service.NewRideService(rideRepo, driverRepo, hub, riderHub)
	rideHandler := handler.NewRideHandler(rideService)

	ratingRepo := repository.NewRatingRepo(a)
	ratingService := service.NewRatingService(ratingRepo, rideRepo, driverRepo)
	ratingHandler := handler.NewRatingHandler(ratingService)

	adminRepo := repository.NewAdminRepo(a)
	adminService := service.NewAdminService(adminRepo, driverRepo)
	adminHandler := handler.NewAdminHandler(adminService)

	mailSvc := service.NewMailService(a)
	workers.StartEmailWorkers(mailSvc)

	wsHandler := handler.NewWSHandler(hub, riderHub, driverService, rideService)

	// User Auth
	r.POST("/auth/user/register", userHandler.PreRegisterUser)
	r.POST("/auth/user/verify", userHandler.RegisterUser)
	r.POST("/auth/user/login", userHandler.LoginUser)

	// Driver Auth
	r.POST("/auth/driver/register", driverHandler.PreRegisterDriver)
	r.POST("/auth/driver/verify", driverHandler.RegisterDriver)
	r.POST("/auth/driver/login", driverHandler.LoginDriver)

	authenticated := r.Group("")

	authenticated.Use(middleware.AuthRequired(cfg.JWTSecret))

	
	rideGroup := authenticated.Group("/ride")
	{

		rideGroup.POST("/request", rideHandler.RequestRide)
		rideGroup.POST("/cancel", rideHandler.CancelRide)
		rideGroup.GET("/history", rideHandler.GetRiderHistory)

		rideGroup.POST("/driver/cancel", middleware.RequireOnboardingComplete(a.DB), middleware.RequireActiveDriver(a.DB), rideHandler.DriverCancelRide)
		rideGroup.GET("/driver/history", middleware.RequireOnboardingComplete(a.DB), middleware.RequireActiveDriver(a.DB), rideHandler.GetDriverHistory)
	}

	authenticated.GET("/ws/driver",middleware.RequireOnboardingComplete(a.DB), middleware.RequireActiveDriver(a.DB), wsHandler.DriverConnect)
	authenticated.GET("/ws/rider", wsHandler.RiderConnect)

	authenticated.POST("/rating/driver", ratingHandler.RateDriver)

	driver := authenticated.Group("/driver")
	driver.Use(middleware.RequireDriver())
	{
		driver.PUT("/profile", driverHandler.UpdateDriverProfile)
		driver.PUT("/license", driverHandler.UpdateDriverLicense)
		driver.POST("/vehicle", driverHandler.InsertVehicle)
		driver.POST("/documents", driverHandler.DriverProofUpload)
		driver.POST("/onboarding/complete", driverHandler.AgreeTerms)
		driver.GET("/profile", driverHandler.GetDriver)
	}

	admin := authenticated.Group("/admin")
	admin.Use(middleware.RequireAdmin())
	{

		admin.GET("/stats", adminHandler.GetDashboardStats)

		admin.GET("/drivers", adminHandler.GetAllDrivers) // ?status=pending
		admin.PUT("/drivers/:id/approve", adminHandler.ApproveDriver)
		admin.PUT("/drivers/:id/reject", adminHandler.RejectDriver)
		admin.PUT("/drivers/:id/suspend", adminHandler.SuspendDriver)
		admin.PUT("/drivers/:id/ban", adminHandler.BanDriver)

		admin.PUT("/drivers/:id/verify-identity", adminHandler.VerifyDriverIdentity)
		admin.PUT("/drivers/:id/verify-vehicle", adminHandler.VerifyDriverVehicle)
		admin.PUT("/drivers/:id/verify-license", adminHandler.VerifyDriverLicense)

		admin.GET("/riders", adminHandler.GetAllRiders)

		admin.GET("/rides", adminHandler.GetAllRides) // ?status=completed
	}

	return r
}
