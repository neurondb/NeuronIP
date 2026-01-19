package users

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* NotificationService provides notification management functionality */
type NotificationService struct {
	pool *pgxpool.Pool
}

/* NewNotificationService creates a new notification service */
func NewNotificationService(pool *pgxpool.Pool) *NotificationService {
	return &NotificationService{
		pool: pool,
	}
}

/* CreateNotification creates a new notification */
func (s *NotificationService) CreateNotification(ctx context.Context, userID uuid.UUID, notificationType, title, message string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	query := `INSERT INTO neuronip.user_notifications (user_id, type, title, message, metadata) 
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := s.pool.Exec(ctx, query, userID, notificationType, title, message, metadata)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

/* GetUserNotifications retrieves notifications for a user */
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit int, unreadOnly bool) ([]db.UserNotification, error) {
	query := `SELECT id, user_id, type, title, message, read, metadata, created_at 
	          FROM neuronip.user_notifications WHERE user_id = $1`
	
	args := []interface{}{userID}
	if unreadOnly {
		query += " AND read = false"
	}
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	var notifications []db.UserNotification
	for rows.Next() {
		var notification db.UserNotification
		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.Type,
			&notification.Title, &notification.Message, &notification.Read,
			&notification.Metadata, &notification.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, notification)
	}
	return notifications, nil
}

/* MarkNotificationRead marks a notification as read */
func (s *NotificationService) MarkNotificationRead(ctx context.Context, notificationID uuid.UUID) error {
	query := `UPDATE neuronip.user_notifications SET read = true WHERE id = $1`
	_, err := s.pool.Exec(ctx, query, notificationID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

/* MarkAllNotificationsRead marks all notifications as read for a user */
func (s *NotificationService) MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE neuronip.user_notifications SET read = true WHERE user_id = $1 AND read = false`
	_, err := s.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}
