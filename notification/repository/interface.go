package repository

import (
	"context"

	"notification-service/models"
)

// NotificationRepo defines the interface for notification data access.
type NotificationRepo interface {
	Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error)
	CreateFromEvent(ctx context.Context, userID int64, notifType, channel, title, content string, metadata map[string]interface{}) (*models.Notification, error)
	GetByID(ctx context.Context, id int64) (*models.Notification, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) (*models.NotificationListResponse, error)
	ListAll(ctx context.Context, limit, offset int) (*models.NotificationListResponse, error)
	MarkAsRead(ctx context.Context, id int64) (*models.Notification, error)
	MarkAsSent(ctx context.Context, id int64) (*models.Notification, error)
	MarkAllAsReadForUser(ctx context.Context, userID int64) error
	Delete(ctx context.Context, id int64) error
}
