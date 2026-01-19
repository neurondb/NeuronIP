package support

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neurondb/NeuronIP/api/internal/agent"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* Service provides customer support memory functionality */
type Service struct {
	queries      *db.Queries
	pool         *pgxpool.Pool
	agentClient  *agent.Client
	neurondbClient *neurondb.Client
}

/* NewService creates a new support service */
func NewService(queries *db.Queries, pool *pgxpool.Pool, agentClient *agent.Client, neurondbClient *neurondb.Client) *Service {
	return &Service{
		queries:       queries,
		pool:          pool,
		agentClient:   agentClient,
		neurondbClient: neurondbClient,
	}
}

/* TicketRequest represents a ticket creation request */
type TicketRequest struct {
	CustomerID    string
	CustomerEmail *string
	Subject       string
	Priority      string
	Message       string
	Metadata      map[string]interface{}
}

/* ConversationMessage represents a conversation message */
type ConversationMessage struct {
	ID         uuid.UUID              `json:"id"`
	TicketID   uuid.UUID              `json:"ticket_id"`
	MessageText string                `json:"message_text"`
	SenderType string                 `json:"sender_type"`
	SenderID   *string                `json:"sender_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

/* SimilarCase represents a similar resolved case */
type SimilarCase struct {
	TicketID     uuid.UUID              `json:"ticket_id"`
	TicketNumber string                 `json:"ticket_number"`
	Subject      string                 `json:"subject"`
	Similarity   float64                `json:"similarity"`
	Conversations []ConversationMessage `json:"conversations,omitempty"`
}

/* MemoryEntry represents a support memory entry */
type MemoryEntry struct {
	ID           uuid.UUID              `json:"id"`
	CustomerID   string                 `json:"customer_id"`
	MemoryType   string                 `json:"memory_type"`
	MemoryContent string                `json:"memory_content"`
	ImportanceScore float64             `json:"importance_score"`
	LastAccessedAt *time.Time           `json:"last_accessed_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

/* CreateTicket creates a new support ticket */
func (s *Service) CreateTicket(ctx context.Context, req TicketRequest) (*db.SupportTicket, error) {
	ticketNumber := fmt.Sprintf("TKT-%d", time.Now().Unix())
	
	// Create ticket
	ticketID := uuid.New()
	now := time.Now()
	metadataJSON, _ := json.Marshal(req.Metadata)

	query := `
		INSERT INTO neuronip.support_tickets 
		(id, ticket_number, customer_id, customer_email, subject, status, priority, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, ticket_number, customer_id, customer_email, subject, status, priority, 
		          assigned_agent_id, metadata, created_at, updated_at, resolved_at`

	var ticket db.SupportTicket
	var customerEmail sql.NullString
	var assignedAgentID *uuid.UUID
	var resolvedAt sql.NullTime

	err := s.pool.QueryRow(ctx, query,
		ticketID, ticketNumber, req.CustomerID, req.CustomerEmail, req.Subject,
		"open", req.Priority, metadataJSON, now, now,
	).Scan(
		&ticket.ID, &ticket.TicketNumber, &ticket.CustomerID, &customerEmail,
		&ticket.Subject, &ticket.Status, &ticket.Priority, &assignedAgentID,
		&ticket.Metadata, &ticket.CreatedAt, &ticket.UpdatedAt, &resolvedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	if customerEmail.Valid {
		ticket.CustomerEmail = &customerEmail.String
	}
	ticket.AssignedAgentID = assignedAgentID
	if resolvedAt.Valid {
		ticket.ResolvedAt = &resolvedAt.Time
	}

	// Add initial message as conversation
	messageText := req.Message
	if messageText == "" {
		messageText = req.Subject
	}
	
	err = s.AddConversation(ctx, ticket.ID, messageText, "customer", &req.CustomerID, nil)
	if err != nil {
		// Log error but don't fail ticket creation
	}

	return &ticket, nil
}

/* GetTicket retrieves a support ticket with conversations */
func (s *Service) GetTicket(ctx context.Context, id uuid.UUID) (*db.SupportTicket, []ConversationMessage, error) {
	ticket, err := s.queries.GetSupportTicketByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// Get conversations
	conversations, err := s.GetConversations(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	return ticket, conversations, nil
}

/* ListTickets retrieves a list of support tickets */
func (s *Service) ListTickets(ctx context.Context, status string, customerID string, limit int) ([]db.SupportTicket, error) {
	if limit <= 0 {
		limit = 100
	}

	var whereClauses []string
	var args []interface{}
	argIndex := 1

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}
	if customerID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("customer_id = $%d", argIndex))
		args = append(args, customerID)
		argIndex++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, ticket_number, customer_id, customer_email, subject, status, priority,
		       assigned_agent_id, metadata, created_at, updated_at, resolved_at
		FROM neuronip.support_tickets
		%s
		ORDER BY created_at DESC
		LIMIT $%d`, whereClause, argIndex)
	
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tickets: %w", err)
	}
	defer rows.Close()

	var tickets []db.SupportTicket
	for rows.Next() {
		var ticket db.SupportTicket
		var customerEmail sql.NullString
		var assignedAgentID *uuid.UUID
		var resolvedAt sql.NullTime
		var metadataJSON json.RawMessage

		err := rows.Scan(&ticket.ID, &ticket.TicketNumber, &ticket.CustomerID, &customerEmail,
			&ticket.Subject, &ticket.Status, &ticket.Priority, &assignedAgentID,
			&metadataJSON, &ticket.CreatedAt, &ticket.UpdatedAt, &resolvedAt)
		if err != nil {
			continue
		}

		if customerEmail.Valid {
			ticket.CustomerEmail = &customerEmail.String
		}
		ticket.AssignedAgentID = assignedAgentID
		if resolvedAt.Valid {
			ticket.ResolvedAt = &resolvedAt.Time
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &ticket.Metadata)
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

/* GetConversations retrieves conversations for a ticket */
func (s *Service) GetConversations(ctx context.Context, ticketID uuid.UUID) ([]ConversationMessage, error) {
	query := `
		SELECT id, ticket_id, message_text, sender_type, sender_id, metadata, created_at
		FROM neuronip.support_conversations
		WHERE ticket_id = $1
		ORDER BY created_at ASC`

	rows, err := s.pool.Query(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	defer rows.Close()

	var conversations []ConversationMessage
	for rows.Next() {
		var conv ConversationMessage
		var senderID sql.NullString
		var metadataJSON json.RawMessage

		err := rows.Scan(&conv.ID, &conv.TicketID, &conv.MessageText, &conv.SenderType,
			&senderID, &metadataJSON, &conv.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}

		if senderID.Valid {
			conv.SenderID = &senderID.String
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &conv.Metadata)
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

/* AddConversation adds a message to ticket conversation */
func (s *Service) AddConversation(ctx context.Context, ticketID uuid.UUID, messageText string, senderType string, senderID *string, metadata map[string]interface{}) error {
	convID := uuid.New()
	metadataJSON, _ := json.Marshal(metadata)

	// Generate embedding for the message
	embedding, err := s.neurondbClient.GenerateEmbedding(ctx, messageText, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		// Log error but continue without embedding
		embedding = ""
	}

	query := `
		INSERT INTO neuronip.support_conversations 
		(id, ticket_id, message_text, sender_type, sender_id, embedding, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6::vector, $7, $8)`

	_, err = s.pool.Exec(ctx, query, convID, ticketID, messageText, senderType, senderID, embedding, metadataJSON, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add conversation: %w", err)
	}

	return nil
}

/* GetSimilarCases finds similar resolved cases */
func (s *Service) GetSimilarCases(ctx context.Context, ticketID uuid.UUID, limit int) ([]SimilarCase, error) {
	if limit <= 0 {
		limit = 10
	}

	// Get ticket conversations
	conversations, err := s.GetConversations(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket conversations: %w", err)
	}

	if len(conversations) == 0 {
		return []SimilarCase{}, nil
	}

	// Combine all messages for similarity search
	var combinedText string
	for _, conv := range conversations {
		combinedText += conv.MessageText + " "
	}

	// Generate embedding for combined text
	queryEmbedding, err := s.neurondbClient.GenerateEmbedding(ctx, combinedText, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Search for similar conversations from resolved tickets
	searchQuery := `
		SELECT DISTINCT
			st.id as ticket_id,
			st.ticket_number,
			st.subject,
			1 - (sc.embedding <=> $1::vector) as similarity
		FROM neuronip.support_tickets st
		JOIN neuronip.support_conversations sc ON sc.ticket_id = st.id
		WHERE st.status = 'resolved'
			AND st.id != $2
			AND sc.embedding IS NOT NULL
			AND 1 - (sc.embedding <=> $1::vector) >= 0.5
		ORDER BY sc.embedding <=> $1::vector
		LIMIT $3`

	rows, err := s.pool.Query(ctx, searchQuery, queryEmbedding, ticketID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar cases: %w", err)
	}
	defer rows.Close()

	var similarCases []SimilarCase
	for rows.Next() {
		var sc SimilarCase
		err := rows.Scan(&sc.TicketID, &sc.TicketNumber, &sc.Subject, &sc.Similarity)
		if err != nil {
			continue
		}
		similarCases = append(similarCases, sc)
	}

	return similarCases, nil
}

/* GenerateReply generates a context-aware reply */
func (s *Service) GenerateReply(ctx context.Context, ticketID uuid.UUID) (string, []SimilarCase, error) {
	// Get ticket and conversations
	ticket, conversations, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Get similar cases
	similarCases, err := s.GetSimilarCases(ctx, ticketID, 5)
	if err != nil {
		// Log error but continue
	}

	// Build context for AI generation
	context := []map[string]interface{}{
		{
			"ticket_id":   ticket.ID.String(),
			"subject":     ticket.Subject,
			"customer_id": ticket.CustomerID,
		},
	}

	// Add conversation context
	for _, conv := range conversations {
		context = append(context, map[string]interface{}{
			"sender":  conv.SenderType,
			"message": conv.MessageText,
		})
	}

	// Add similar cases context
	for _, similar := range similarCases {
		context = append(context, map[string]interface{}{
			"similar_ticket": similar.TicketNumber,
			"similarity":     similar.Similarity,
		})
	}

	// Generate prompt
	prompt := fmt.Sprintf("Generate a helpful reply for support ticket %s. Subject: %s", ticket.TicketNumber, ticket.Subject)

	// Generate reply using NeuronAgent
	reply, err := s.agentClient.GenerateReply(ctx, context, prompt)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate reply: %w", err)
	}

	return reply, similarCases, nil
}

/* GetMemory retrieves customer memory */
func (s *Service) GetMemory(ctx context.Context, customerID string) ([]MemoryEntry, error) {
	query := `
		SELECT id, customer_id, memory_type, memory_content, importance_score,
		       last_accessed_at, created_at, updated_at
		FROM neuronip.support_memory
		WHERE customer_id = $1
		ORDER BY importance_score DESC, updated_at DESC`

	rows, err := s.pool.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}
	defer rows.Close()

	var memories []MemoryEntry
	for rows.Next() {
		var mem MemoryEntry
		var lastAccessedAt sql.NullTime

		err := rows.Scan(&mem.ID, &mem.CustomerID, &mem.MemoryType, &mem.MemoryContent,
			&mem.ImportanceScore, &lastAccessedAt, &mem.CreatedAt, &mem.UpdatedAt)
		if err != nil {
			continue
		}

		if lastAccessedAt.Valid {
			mem.LastAccessedAt = &lastAccessedAt.Time
		}

		memories = append(memories, mem)
	}

	return memories, nil
}

/* CreateMemory creates a new memory entry */
func (s *Service) CreateMemory(ctx context.Context, customerID string, memoryType string, memoryContent string, importanceScore float64) (*MemoryEntry, error) {
	memID := uuid.New()
	now := time.Now()

	// Generate embedding
	embedding, err := s.neurondbClient.GenerateEmbedding(ctx, memoryContent, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		embedding = ""
	}

	query := `
		INSERT INTO neuronip.support_memory 
		(id, customer_id, memory_type, memory_content, embedding, importance_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::vector, $6, $7, $8)
		RETURNING id, customer_id, memory_type, memory_content, importance_score, created_at, updated_at`

	var mem MemoryEntry
	err = s.pool.QueryRow(ctx, query, memID, customerID, memoryType, memoryContent,
		embedding, importanceScore, now, now).Scan(
		&mem.ID, &mem.CustomerID, &mem.MemoryType, &mem.MemoryContent,
		&mem.ImportanceScore, &mem.CreatedAt, &mem.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory: %w", err)
	}

	return &mem, nil
}

/* UpdateMemory updates a memory entry */
func (s *Service) UpdateMemory(ctx context.Context, memoryID uuid.UUID, memoryContent string, importanceScore float64) error {
	// Generate new embedding
	embedding, err := s.neurondbClient.GenerateEmbedding(ctx, memoryContent, "sentence-transformers/all-MiniLM-L6-v2")
	if err != nil {
		embedding = ""
	}

	query := `
		UPDATE neuronip.support_memory 
		SET memory_content = $1, embedding = $2::vector, importance_score = $3, updated_at = $4
		WHERE id = $5`

	_, err = s.pool.Exec(ctx, query, memoryContent, embedding, importanceScore, time.Now(), memoryID)
	if err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}

	return nil
}

/* AgentFeedback represents feedback from an agent on generated replies */
type AgentFeedback struct {
	TicketID      uuid.UUID `json:"ticket_id"`
	OriginalReply string    `json:"original_reply"`
	EditedReply   string    `json:"edited_reply"`
	FeedbackType  string    `json:"feedback_type"` // "correction", "improvement", "approval"
	FeedbackNotes string    `json:"feedback_notes,omitempty"`
	AgentID       string    `json:"agent_id"`
}

/* RecordAgentFeedback records agent feedback and learns from it */
func (s *Service) RecordAgentFeedback(ctx context.Context, feedback AgentFeedback) error {
	// Store feedback as memory for learning
	feedbackContent := fmt.Sprintf("Feedback on ticket %s: %s. Original: %s. Edited: %s. Notes: %s",
		feedback.TicketID.String(), feedback.FeedbackType,
		truncateString(feedback.OriginalReply, 200),
		truncateString(feedback.EditedReply, 200),
		feedback.FeedbackNotes)

	// Determine importance score based on feedback type
	importanceScore := 0.5
	if feedback.FeedbackType == "correction" {
		importanceScore = 0.9 // Corrections are highly important
	} else if feedback.FeedbackType == "improvement" {
		importanceScore = 0.7
	}

	// Get ticket to extract customer ID
	ticket, _, err := s.GetTicket(ctx, feedback.TicketID)
	if err == nil {
		// Store feedback in customer memory
		_, err = s.CreateMemory(ctx, ticket.CustomerID, "feedback", feedbackContent, importanceScore)
		if err != nil {
			// Log but don't fail
		}
	}

	// Also store general pattern memory for improving future replies
	patternContent := fmt.Sprintf("Reply pattern learned from feedback: %s", feedback.FeedbackNotes)
	if feedback.EditedReply != "" && feedback.OriginalReply != "" {
		// Store the pattern of what was changed
		patternContent = fmt.Sprintf("When original reply was: '%s', better reply is: '%s'. Reason: %s",
			truncateString(feedback.OriginalReply, 150),
			truncateString(feedback.EditedReply, 150),
			feedback.FeedbackNotes)
	}

	// Store as general pattern memory (no specific customer)
	generalMemoryID := "general-patterns"
	_, err = s.CreateMemory(ctx, generalMemoryID, "pattern", patternContent, importanceScore)
	if err != nil {
		// Memory might already exist, try to update
		memories, _ := s.GetMemory(ctx, generalMemoryID)
		if len(memories) > 0 {
			// Update existing memory with new pattern
			existingContent := memories[0].MemoryContent + "\n" + patternContent
			s.UpdateMemory(ctx, memories[0].ID, existingContent, importanceScore)
		}
	}

	return nil
}

/* truncateString truncates a string to max length */
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
