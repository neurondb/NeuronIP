package collaboration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

/* ThreadService provides discussion thread functionality */
type ThreadService struct {
	pool *pgxpool.Pool
}

/* NewThreadService creates a new thread service */
func NewThreadService(pool *pgxpool.Pool) *ThreadService {
	return &ThreadService{pool: pool}
}

/* Thread represents a discussion thread */
type Thread struct {
	ID           uuid.UUID              `json:"id"`
	ResourceType string                 `json:"resource_type"` // "dashboard", "query", "metric", "document"
	ResourceID   uuid.UUID              `json:"resource_id"`
	Title        string                 `json:"title"`
	InitialPost  ThreadPost             `json:"initial_post"`
	Posts        []ThreadPost           `json:"posts,omitempty"`
	Status       string                 `json:"status"` // "open", "resolved", "archived"
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

/* ThreadPost represents a post in a thread */
type ThreadPost struct {
	ID          uuid.UUID              `json:"id"`
	ThreadID    uuid.UUID              `json:"thread_id"`
	AuthorID    string                 `json:"author_id"`
	Content     string                 `json:"content"`
	ParentPostID *uuid.UUID            `json:"parent_post_id,omitempty"`
	Attachments []string               `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

/* CreateThread creates a new discussion thread */
func (ts *ThreadService) CreateThread(ctx context.Context, thread Thread) (*Thread, error) {
	thread.ID = uuid.New()
	thread.Status = "open"
	thread.CreatedAt = time.Now()
	thread.UpdatedAt = time.Now()
	
	// Create initial post
	thread.InitialPost.ID = uuid.New()
	thread.InitialPost.ThreadID = thread.ID
	thread.InitialPost.CreatedAt = time.Now()
	thread.InitialPost.UpdatedAt = time.Now()
	
	tagsJSON, _ := json.Marshal(thread.Tags)
	metadataJSON, _ := json.Marshal(thread.Metadata)
	
	query := `
		INSERT INTO neuronip.discussion_threads 
		(id, resource_type, resource_id, title, status, tags, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	
	err := ts.pool.QueryRow(ctx, query,
		thread.ID, thread.ResourceType, thread.ResourceID, thread.Title,
		thread.Status, tagsJSON, metadataJSON, thread.CreatedAt, thread.UpdatedAt,
	).Scan(&thread.ID, &thread.CreatedAt, &thread.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}
	
	// Create initial post
	if err := ts.AddPost(ctx, thread.InitialPost); err != nil {
		return nil, fmt.Errorf("failed to create initial post: %w", err)
	}
	
	return &thread, nil
}

/* AddPost adds a post to a thread */
func (ts *ThreadService) AddPost(ctx context.Context, post ThreadPost) error {
	post.ID = uuid.New()
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	
	attachmentsJSON, _ := json.Marshal(post.Attachments)
	metadataJSON, _ := json.Marshal(post.Metadata)
	
	query := `
		INSERT INTO neuronip.discussion_posts 
		(id, thread_id, author_id, content, parent_post_id, attachments, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	_, err := ts.pool.Exec(ctx, query,
		post.ID, post.ThreadID, post.AuthorID, post.Content, post.ParentPostID,
		attachmentsJSON, metadataJSON, post.CreatedAt, post.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to add post: %w", err)
	}
	
	// Update thread updated_at
	ts.pool.Exec(ctx, `UPDATE neuronip.discussion_threads SET updated_at = NOW() WHERE id = $1`, post.ThreadID)
	
	return nil
}

/* GetThread retrieves a thread with all posts */
func (ts *ThreadService) GetThread(ctx context.Context, threadID uuid.UUID) (*Thread, error) {
	// Get thread
	var thread Thread
	var tagsJSON, metadataJSON json.RawMessage
	
	query := `
		SELECT id, resource_type, resource_id, title, status, tags, metadata, created_at, updated_at
		FROM neuronip.discussion_threads
		WHERE id = $1
	`
	
	err := ts.pool.QueryRow(ctx, query, threadID).Scan(
		&thread.ID, &thread.ResourceType, &thread.ResourceID, &thread.Title,
		&thread.Status, &tagsJSON, &metadataJSON, &thread.CreatedAt, &thread.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	
	if tagsJSON != nil {
		json.Unmarshal(tagsJSON, &thread.Tags)
	}
	if metadataJSON != nil {
		json.Unmarshal(metadataJSON, &thread.Metadata)
	}
	
	// Get posts
	postsQuery := `
		SELECT id, thread_id, author_id, content, parent_post_id, attachments, metadata, created_at, updated_at
		FROM neuronip.discussion_posts
		WHERE thread_id = $1
		ORDER BY created_at ASC
	`
	
	rows, err := ts.pool.Query(ctx, postsQuery, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var post ThreadPost
		var parentPostID *uuid.UUID
		var attachmentsJSON, metadataJSON json.RawMessage
		
		err := rows.Scan(
			&post.ID, &post.ThreadID, &post.AuthorID, &post.Content, &parentPostID,
			&attachmentsJSON, &metadataJSON, &post.CreatedAt, &post.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		post.ParentPostID = parentPostID
		if attachmentsJSON != nil {
			json.Unmarshal(attachmentsJSON, &post.Attachments)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &post.Metadata)
		}
		
		if len(thread.Posts) == 0 {
			thread.InitialPost = post
		}
		thread.Posts = append(thread.Posts, post)
	}
	
	return &thread, nil
}

/* GetThreadsForResource retrieves threads for a resource */
func (ts *ThreadService) GetThreadsForResource(ctx context.Context, resourceType string, resourceID uuid.UUID, limit int) ([]Thread, error) {
	if limit <= 0 {
		limit = 50
	}
	
	query := `
		SELECT id, resource_type, resource_id, title, status, tags, metadata, created_at, updated_at
		FROM neuronip.discussion_threads
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY updated_at DESC
		LIMIT $3
	`
	
	rows, err := ts.pool.Query(ctx, query, resourceType, resourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get threads: %w", err)
	}
	defer rows.Close()
	
	var threads []Thread
	for rows.Next() {
		var thread Thread
		var tagsJSON, metadataJSON json.RawMessage
		
		err := rows.Scan(
			&thread.ID, &thread.ResourceType, &thread.ResourceID, &thread.Title,
			&thread.Status, &tagsJSON, &metadataJSON, &thread.CreatedAt, &thread.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		if tagsJSON != nil {
			json.Unmarshal(tagsJSON, &thread.Tags)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &thread.Metadata)
		}
		
		threads = append(threads, thread)
	}
	
	return threads, nil
}

/* ResolveThread marks a thread as resolved */
func (ts *ThreadService) ResolveThread(ctx context.Context, threadID uuid.UUID) error {
	query := `
		UPDATE neuronip.discussion_threads
		SET status = 'resolved', updated_at = NOW()
		WHERE id = $1
	`
	
	_, err := ts.pool.Exec(ctx, query, threadID)
	return err
}
