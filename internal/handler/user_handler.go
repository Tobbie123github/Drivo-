package handler

import (
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

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

func (h *UserHandler) PreRegisterUser(c *gin.Context) {

	var input models.UserRegisterInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}

	if err := h.svc.PreRegister(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Please verify your account",
	})

}

func (h *UserHandler) RegisterUser(c *gin.Context) {

	var input models.UserVerifyEmail

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

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registration Successful, Proceed to login",
		"data":    u,
	})

}

func (h *UserHandler) LoginUser(c *gin.Context) {

	var input models.UserLoginInput

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

func (h *UserHandler) RequestPasswordReset(c *gin.Context) {

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

	resetLink := fmt.Sprintf("https://drivo-nine.vercel.app/reset-password?token=%s", token)

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

func (h *UserHandler) PasswordReset(c *gin.Context) {

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

func (h *UserHandler) UpdateFCMToken(c *gin.Context) {
	userID, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		FCMToken string `json:"fcm_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fcm_token is required"})
		return
	}

	uid, _ := uuid.Parse(userID)
	if err := h.svc.UpdateFCMToken(uid, input.FCMToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "FCM token updated"})
}
