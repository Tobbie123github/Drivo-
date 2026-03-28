package repository

import (
	"context"
	"drivo/internal/app"
	"drivo/internal/models"

	"github.com/google/uuid"
)

type ChatRepo struct {
	db *app.App
}

func NewChatRepo(db *app.App) *ChatRepo {
	return &ChatRepo{
		db: db,
	}
}

func (r *ChatRepo) CreateSession(ctx context.Context, rideID, driverID, riderID uuid.UUID) (*models.ChatSession, error) {
    session := &models.ChatSession{
        RideID:   rideID,
        DriverID: driverID,
        RiderID:  riderID,
        IsActive: true,
    }
    err := r.db.DB.WithContext(ctx).Create(session).Error
    return session, err
}


func (r *ChatRepo) GetSessionByRideID(ctx context.Context, rideID uuid.UUID) (*models.ChatSession, error) {
	var session models.ChatSession
	err := r.db.DB.WithContext(ctx).Where("ride_id = ?", rideID).First(&session).Error

	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatRepo) SaveMessage(ctx context.Context, message models.ChatMessage) error {
	return r.db.DB.WithContext(ctx).Create(&message).Error
}


func (r *ChatRepo) GetMessages(ctx context.Context, sessionID uuid.UUID) ([]models.ChatMessage, error) {
    var messages []models.ChatMessage
    err := r.db.DB.WithContext(ctx).
        Where("session_id = ?", sessionID).
        Order("created_at ASC").
        Find(&messages).Error
    return messages, err
}


func (r *ChatRepo) CloseSession(ctx context.Context, rideID uuid.UUID) error {
    return r.db.DB.WithContext(ctx).
        Model(&models.ChatSession{}).
        Where("ride_id = ?", rideID).
        Update("is_active", false).Error
}


func (r *ChatRepo) MarkMessagesRead(ctx context.Context, sessionID uuid.UUID, readerID uuid.UUID) error {
    return r.db.DB.WithContext(ctx).
        Model(&models.ChatMessage{}).
        Where("session_id = ? AND sender_id != ? AND is_read = false", sessionID, readerID).
        Update("is_read", true).Error
}