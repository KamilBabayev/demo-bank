package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"notification-service/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrInvalidInput         = errors.New("invalid input")
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	query := `
		INSERT INTO notifications (user_id, type, channel, title, content, metadata, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
		RETURNING id, user_id, type, channel, title, content, metadata, status,
		          read_at, sent_at, created_at, updated_at
	`

	notification := &models.Notification{}
	err := r.db.QueryRow(
		ctx, query,
		req.UserID, req.Type, req.Channel, req.Title, req.Content, req.Metadata,
	).Scan(
		&notification.ID, &notification.UserID, &notification.Type, &notification.Channel,
		&notification.Title, &notification.Content, &notification.Metadata, &notification.Status,
		&notification.ReadAt, &notification.SentAt, &notification.CreatedAt, &notification.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

// CreateFromEvent creates a notification from an event (simpler interface)
func (r *NotificationRepository) CreateFromEvent(ctx context.Context, userID int64, notifType, channel, title, content string, metadata map[string]interface{}) (*models.Notification, error) {
	var metadataJSON *json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		raw := json.RawMessage(data)
		metadataJSON = &raw
	}

	req := &models.CreateNotificationRequest{
		UserID:   userID,
		Type:     notifType,
		Channel:  channel,
		Title:    title,
		Content:  content,
		Metadata: metadataJSON,
	}

	return r.Create(ctx, req)
}

// GetByID retrieves a notification by ID
func (r *NotificationRepository) GetByID(ctx context.Context, id int64) (*models.Notification, error) {
	query := `
		SELECT id, user_id, type, channel, title, content, metadata, status,
		       read_at, sent_at, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	notification := &models.Notification{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&notification.ID, &notification.UserID, &notification.Type, &notification.Channel,
		&notification.Title, &notification.Content, &notification.Metadata, &notification.Status,
		&notification.ReadAt, &notification.SentAt, &notification.CreatedAt, &notification.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return notification, nil
}

// ListByUserID retrieves all notifications for a user
func (r *NotificationRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) (*models.NotificationListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	var unread int64
	unreadQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND status != 'read'`
	if err := r.db.QueryRow(ctx, unreadQuery, userID).Scan(&unread); err != nil {
		return nil, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	query := `
		SELECT id, user_id, type, channel, title, content, metadata, status,
		       read_at, sent_at, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	notifications := []models.Notification{}
	for rows.Next() {
		var notification models.Notification
		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.Type, &notification.Channel,
			&notification.Title, &notification.Content, &notification.Metadata, &notification.Status,
			&notification.ReadAt, &notification.SentAt, &notification.CreatedAt, &notification.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return &models.NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		Unread:        unread,
	}, nil
}

// ListAll retrieves all notifications (admin only)
func (r *NotificationRepository) ListAll(ctx context.Context, limit, offset int) (*models.NotificationListResponse, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM notifications`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	query := `
		SELECT id, user_id, type, channel, title, content, metadata, status,
		       read_at, sent_at, created_at, updated_at
		FROM notifications
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	notifications := []models.Notification{}
	for rows.Next() {
		var notification models.Notification
		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.Type, &notification.Channel,
			&notification.Title, &notification.Content, &notification.Metadata, &notification.Status,
			&notification.ReadAt, &notification.SentAt, &notification.CreatedAt, &notification.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return &models.NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		Unread:        0, // Not relevant for admin view
	}, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id int64) (*models.Notification, error) {
	query := `
		UPDATE notifications
		SET status = 'read', read_at = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, type, channel, title, content, metadata, status,
		          read_at, sent_at, created_at, updated_at
	`

	notification := &models.Notification{}
	err := r.db.QueryRow(ctx, query, time.Now(), id).Scan(
		&notification.ID, &notification.UserID, &notification.Type, &notification.Channel,
		&notification.Title, &notification.Content, &notification.Metadata, &notification.Status,
		&notification.ReadAt, &notification.SentAt, &notification.CreatedAt, &notification.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return notification, nil
}

// MarkAsSent marks a notification as sent
func (r *NotificationRepository) MarkAsSent(ctx context.Context, id int64) (*models.Notification, error) {
	query := `
		UPDATE notifications
		SET status = 'sent', sent_at = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, type, channel, title, content, metadata, status,
		          read_at, sent_at, created_at, updated_at
	`

	notification := &models.Notification{}
	err := r.db.QueryRow(ctx, query, time.Now(), id).Scan(
		&notification.ID, &notification.UserID, &notification.Type, &notification.Channel,
		&notification.Title, &notification.Content, &notification.Metadata, &notification.Status,
		&notification.ReadAt, &notification.SentAt, &notification.CreatedAt, &notification.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to mark notification as sent: %w", err)
	}

	return notification, nil
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotificationNotFound
	}

	return nil
}

// MarkAllAsReadForUser marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsReadForUser(ctx context.Context, userID int64) error {
	query := `
		UPDATE notifications
		SET status = 'read', read_at = $1, updated_at = NOW()
		WHERE user_id = $2 AND status != 'read'
	`

	_, err := r.db.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}
