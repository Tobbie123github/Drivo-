package handler

import (
	"context"
	"drivo/internal/jobs"
	"drivo/internal/middleware"
	"drivo/internal/models"
	"drivo/internal/service"
	"drivo/internal/workers"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DriverHandler struct {
	svc     *service.DriverService
	rideSvc *service.RideService
}

func NewDriverHandler(svc *service.DriverService, rideSvc *service.RideService) *DriverHandler {
	return &DriverHandler{
		svc:     svc,
		rideSvc: rideSvc,
	}
}

func (h *DriverHandler) PreRegisterDriver(c *gin.Context) {

	var input models.DriverRegisterInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	email, name, otp, err := h.svc.PreRegister(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	workers.EmailQueue <- jobs.EmailJob{
		Type: jobs.EmailTypeOTP,
		To:   email,
		Name: name,
		OTP:  otp,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Please verify your account",
	})

}

func (h *DriverHandler) RegisterDriver(c *gin.Context) {

	var input models.DriverVerifyEmail

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	u, err := h.svc.VerifyUserEmail(c.Request.Context(), input.OTP, input.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	workers.EmailQueue <- jobs.EmailJob{
		Type: jobs.EmailTypeDriverWelcome,
		To:   *u.Email,
		Name: u.Name,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration Successful, Proceed to login",
		"data":    u,
	})

}

func (h *DriverHandler) LoginDriver(c *gin.Context) {

	var input models.DriverLoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	res, err := h.svc.Login(input.Password, input.Email)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": res,
	})

}

func (h *DriverHandler) UpdateDriverProfile(c *gin.Context) {

	userID, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Error grtting user id",
		})

		return
	}

	userId, _ := uuid.Parse(userID)

	var input models.DriverProfileInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if err := h.svc.CompleteProfile(c.Request.Context(), userId, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated",
	})

}

func (h *DriverHandler) UpdateDriverLicense(c *gin.Context) {

	userID, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Error grtting user id",
		})

		return
	}

	userId, _ := uuid.Parse(userID)

	var input models.DriverLicence

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if err := h.svc.UpdateLicence(c.Request.Context(), userId, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "License updated",
	})

}

func (h *DriverHandler) InsertVehicle(c *gin.Context) {

	userID, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "UnAuthorized User",
		})
		return
	}

	var input models.VehicleInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	userId, _ := uuid.Parse(userID)

	if err := h.svc.Vehicle(c.Request.Context(), userId, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vehicle added",
	})
}

func (h *DriverHandler) DriverProofUpload(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "UnAuthorized User",
		})
		return
	}

	var input models.DocumentUploadInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	userId, _ := uuid.Parse(userID)

	if err := h.svc.ProofofProfile(c.Request.Context(), userId, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Documents uploaded, pending verification",
	})

}

func (h *DriverHandler) AgreeTerms(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "UnAuthorized User",
		})
		return
	}

	var input struct {
		AgreeTerms bool `json:"agree_terms"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if input.AgreeTerms != true {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You must agree to the terms and conditions",
		})
		return
	}

	userId, _ := uuid.Parse(userID)

	if err := h.svc.AgreeTerms(userId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Onboarding Complete, pending account verification",
	})

}

func (h *DriverHandler) GetDriver(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "UnAuthorized User",
		})
		return
	}

	userId, _ := uuid.Parse(userID)

	driver, err := h.svc.GetDriverProfile(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"driver": driver,
	})

}

func (h *DriverHandler) UpdateLocation(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, _ := userID.(string)
	driverUserID, _ := uuid.Parse(uid)

	var req struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	ctx := context.Background()
	if err := h.svc.UpdateLocation(ctx, driverUserID, req.Latitude, req.Longitude); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.rideSvc.PushLocationToRider(ctx, driverUserID, req.Latitude, req.Longitude)

	c.JSON(http.StatusOK, gin.H{"message": "location updated"})
}

func (h *DriverHandler) RequestPasswordReset(c *gin.Context) {

	var input models.UserPasswordResetRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	token, err := h.svc.RequestResetPassword(input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	resetLink := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)

	// send email to user with otp
	workers.EmailQueue <- jobs.EmailJob{
		Type:      jobs.EmailTypePasswordReset,
		To:        input.Email,
		ResetLink: resetLink,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset OTP sent",
		"token":   token,
		"email":   input.Email,
	})

}

func (h *DriverHandler) PasswordReset(c *gin.Context) {

	var input models.PasswordResetRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if err := h.svc.ResetPassword(input.Token, input.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password Reset Successfully",
	})
}
