package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* AnalyticsService provides user analytics functionality */
type AnalyticsService struct {
	pool *pgxpool.Pool
}

/* NewAnalyticsService creates a new analytics service */
func NewAnalyticsService(pool *pgxpool.Pool) *AnalyticsService {
	return &AnalyticsService{
		pool: pool,
	}
}

/* GetUserActivity retrieves activity logs for a user */
func (s *AnalyticsService) GetUserActivity(ctx context.Context, userID uuid.UUID, limit int, activityType *string) ([]db.UserActivityLog, error) {
	query := `SELECT id, user_id, activity_type, resource_type, resource_id, metadata, ip_address, created_at 
	          FROM neuronip.user_activity_logs WHERE user_id = $1`
	
	args := []interface{}{userID}
	if activityType != nil {
		query += " AND activity_type = $2"
		args = append(args, *activityType)
	}
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity logs: %w", err)
	}
	defer rows.Close()

	var logs []db.UserActivityLog
	for rows.Next() {
		var log db.UserActivityLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.ActivityType, &log.ResourceType,
			&log.ResourceID, &log.Metadata, &log.IPAddress, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, nil
}

/* LogUserActivity logs a user activity */
func (s *AnalyticsService) LogUserActivity(ctx context.Context, userID uuid.UUID, activityType string, resourceType *string, resourceID *uuid.UUID, metadata map[string]interface{}, ipAddress *string) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	query := `INSERT INTO neuronip.user_activity_logs (user_id, activity_type, resource_type, resource_id, metadata, ip_address) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := s.pool.Exec(ctx, query, userID, activityType, resourceType, resourceID, metadata, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to log activity: %w", err)
	}
	return nil
}
