package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* PreparedQueries manages prepared statements for frequently executed queries */
type PreparedQueries struct {
	pool *pgxpool.Pool
}

/* NewPreparedQueries creates a new prepared queries manager */
func NewPreparedQueries(pool *pgxpool.Pool) *PreparedQueries {
	return &PreparedQueries{pool: pool}
}

/* QueryOptions provides options for query execution */
type QueryOptions struct {
	Limit    int           // Maximum number of rows to return
	Offset   int           // Offset for pagination
	Timeout  time.Duration // Query timeout
	OrderBy  string        // ORDER BY clause
	OrderDir string        // ASC or DESC
}

/* DefaultQueryOptions returns default query options */
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		Limit:    100,
		Offset:   0,
		Timeout:  30 * time.Second,
		OrderBy:  "created_at",
		OrderDir: "DESC",
	}
}

/* ExecuteQueryWithOptions executes a query with options and returns rows */
func (pq *PreparedQueries) ExecuteQueryWithOptions(ctx context.Context, query string, args []interface{}, opts QueryOptions) (pgx.Rows, error) {
	// Apply timeout
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Add pagination if limit is set
	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, opts.Limit)
		if opts.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", len(args)+1)
			args = append(args, opts.Offset)
		}
	}

	// Add ordering if specified
	if opts.OrderBy != "" {
		orderDir := opts.OrderDir
		if orderDir != "ASC" && orderDir != "DESC" {
			orderDir = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", opts.OrderBy, orderDir)
	}

	return pq.pool.Query(ctx, query, args...)
}

/* ExecuteQueryWithPagination executes a query with pagination support */
func (pq *PreparedQueries) ExecuteQueryWithPagination(ctx context.Context, query string, args []interface{}, limit, offset int) (pgx.Rows, error) {
	opts := DefaultQueryOptions()
	opts.Limit = limit
	opts.Offset = offset
	return pq.ExecuteQueryWithOptions(ctx, query, args, opts)
}

/* CountQuery executes a COUNT query */
func (pq *PreparedQueries) CountQuery(ctx context.Context, baseQuery string, whereClause string, args []interface{}) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM (%s %s) AS count_query", baseQuery, whereClause)

	var count int
	err := pq.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count rows: %w", err)
	}
	return count, nil
}

/* PaginatedResult represents a paginated query result */
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
	NextOffset *int        `json:"next_offset,omitempty"`
}

/* NewPaginatedResult creates a new paginated result */
func NewPaginatedResult(data interface{}, total, limit, offset int) *PaginatedResult {
	result := &PaginatedResult{
		Data:    data,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: offset+limit < total,
	}
	if result.HasMore {
		nextOffset := offset + limit
		result.NextOffset = &nextOffset
	}
	return result
}
