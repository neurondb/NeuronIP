package support

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

/* BatchAddConversations adds multiple conversations with batch embedding generation */
func (s *Service) BatchAddConversations(ctx context.Context, ticketID uuid.UUID, conversations []struct {
	MessageText string
	SenderType  string
	SenderID    *string
	Metadata    map[string]interface{}
}) error {
	if len(conversations) == 0 {
		return nil
	}

	// Extract all message texts for batch embedding
	messageTexts := make([]string, len(conversations))
	for i, conv := range conversations {
		messageTexts[i] = conv.MessageText
	}

	// Generate embeddings in batch
	modelName := "sentence-transformers/all-MiniLM-L6-v2"
	embeddings, err := s.neurondbClient.BatchGenerateEmbedding(ctx, messageTexts, modelName)
	if err != nil {
		// Fallback to individual embeddings if batch fails
		embeddings = make([]string, len(conversations))
		for i, text := range messageTexts {
			emb, err := s.neurondbClient.GenerateEmbedding(ctx, text, modelName)
			if err != nil {
				embeddings[i] = "" // Continue without embedding
			} else {
				embeddings[i] = emb
			}
		}
	}

	// Insert all conversations
	query := `
		INSERT INTO neuronip.support_conversations 
		(id, ticket_id, message_text, sender_type, sender_id, embedding, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6::vector, $7, $8)`

	now := time.Now()
	for i, conv := range conversations {
		convID := uuid.New()
		metadataJSON, _ := json.Marshal(conv.Metadata)
		embedding := ""
		if i < len(embeddings) {
			embedding = embeddings[i]
		}

		_, err := s.pool.Exec(ctx, query, convID, ticketID, conv.MessageText, conv.SenderType,
			conv.SenderID, embedding, metadataJSON, now)
		if err != nil {
			return fmt.Errorf("failed to add conversation %d: %w", i, err)
		}
	}

	return nil
}

/* GetSimilarCases finds similar resolved cases */
func (s *Service) GetSimilarCases(ctx context.Context, ticketID uuid.UUID, limit int) ([]SimilarCase, error) {
	return s.GetSimilarCasesWithMetric(ctx, ticketID, limit, "cosine")
}

/* GetSimilarCasesWithMetric finds similar resolved cases using specified distance metric */
func (s *Service) GetSimilarCasesWithMetric(ctx context.Context, ticketID uuid.UUID, limit int, distanceMetric string) ([]SimilarCase, error) {
	if limit <= 0 {
		limit = 10
	}
	if distanceMetric == "" {
		distanceMetric = "cosine"
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

	// Use appropriate search method based on distance metric
	var searchQuery string
	switch distanceMetric {
	case "l2":
		// Use L2 distance (lower is better)
		searchQuery = `
			SELECT DISTINCT
				st.id as ticket_id,
				st.ticket_number,
				st.subject,
				1.0 / (1.0 + (sc.embedding <-> $1::vector)) as similarity
			FROM neuronip.support_tickets st
			JOIN neuronip.support_conversations sc ON sc.ticket_id = st.id
			WHERE st.status = 'resolved'
				AND st.id != $2
				AND sc.embedding IS NOT NULL
				AND 1.0 / (1.0 + (sc.embedding <-> $1::vector)) >= 0.5
			ORDER BY sc.embedding <-> $1::vector
			LIMIT $3`
	case "inner_product":
		// Use inner product (higher is better, but need to normalize)
		searchQuery = `
			SELECT DISTINCT
				st.id as ticket_id,
				st.ticket_number,
				st.subject,
				(sc.embedding <#> $1::vector) * -1 as similarity
			FROM neuronip.support_tickets st
			JOIN neuronip.support_conversations sc ON sc.ticket_id = st.id
			WHERE st.status = 'resolved'
				AND st.id != $2
				AND sc.embedding IS NOT NULL
			ORDER BY sc.embedding <#> $1::vector
			LIMIT $3`
	default: // cosine
		searchQuery = `
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
	}

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

/* CompareConversations compares two conversations using vector similarity */
func (s *Service) CompareConversations(ctx context.Context, convID1, convID2 uuid.UUID, metric string) (float64, error) {
	if metric == "" {
		metric = "cosine"
	}

	// Get embeddings for both conversations
	var embedding1, embedding2 string
	query := `SELECT embedding::text FROM neuronip.support_conversations WHERE id = $1`
	
	err := s.pool.QueryRow(ctx, query, convID1).Scan(&embedding1)
	if err != nil {
		return 0, fmt.Errorf("failed to get embedding for conversation 1: %w", err)
	}

	err = s.pool.QueryRow(ctx, query, convID2).Scan(&embedding2)
	if err != nil {
		return 0, fmt.Errorf("failed to get embedding for conversation 2: %w", err)
	}

	// Calculate similarity based on metric
	var similarity float64
	switch metric {
	case "l2":
		similarityQuery := `SELECT 1.0 / (1.0 + ($1::vector <-> $2::vector)) as similarity`
		err = s.pool.QueryRow(ctx, similarityQuery, embedding1, embedding2).Scan(&similarity)
	case "inner_product":
		similarityQuery := `SELECT ($1::vector <#> $2::vector) * -1 as similarity`
		err = s.pool.QueryRow(ctx, similarityQuery, embedding1, embedding2).Scan(&similarity)
	default: // cosine
		similarityQuery := `SELECT 1 - ($1::vector <=> $2::vector) as similarity`
		err = s.pool.QueryRow(ctx, similarityQuery, embedding1, embedding2).Scan(&similarity)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to calculate similarity: %w", err)
	}
	return similarity, nil
}

/* GetSimilarCasesHybrid performs hybrid search for similar support cases */
func (s *Service) GetSimilarCasesHybrid(ctx context.Context, ticketID uuid.UUID, limit int, keywordQuery string) ([]SimilarCase, error) {
	if limit <= 0 {
		limit = 5
	}

	// Get ticket conversations for context
	ticket, conversations, err := s.GetTicket(ctx, ticketID)
	if err != nil {
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

	// Use NeuronDB HybridSearch for better results
	tableName := "neuronip.support_tickets st JOIN neuronip.support_conversations sc ON sc.ticket_id = st.id"
	embeddingColumn := "sc.embedding"
	textColumn := "st.subject || ' ' || COALESCE(sc.message_text, '')"
	weights := map[string]float64{
		"semantic": 0.7,
		"keyword":  0.3,
	}

	// Build WHERE clause for filtering
	// Note: HybridSearch in client may not support complex WHERE clauses directly
	// So we'll use the direct SQL approach with hybrid search logic
	searchQuery := `
		WITH semantic_results AS (
			SELECT DISTINCT
				st.id as ticket_id,
				st.ticket_number,
				st.subject,
				1 - (sc.embedding <=> $1::vector) as semantic_score
			FROM neuronip.support_tickets st
			JOIN neuronip.support_conversations sc ON sc.ticket_id = st.id
			WHERE st.status = 'resolved'
				AND st.id != $2
				AND sc.embedding IS NOT NULL
				AND 1 - (sc.embedding <=> $1::vector) >= 0.5
			ORDER BY sc.embedding <=> $1::vector
			LIMIT $3
		),
		keyword_results AS (
			SELECT DISTINCT
				st.id as ticket_id,
				st.ticket_number,
				st.subject,
				ts_rank(to_tsvector('english', COALESCE(st.subject, '') || ' ' || COALESCE(sc.message_text, '')), 
					plainto_tsquery('english', $4)) as keyword_score
			FROM neuronip.support_tickets st
			JOIN neuronip.support_conversations sc ON sc.ticket_id = st.id
			WHERE st.status = 'resolved'
				AND st.id != $2
				AND to_tsvector('english', COALESCE(st.subject, '') || ' ' || COALESCE(sc.message_text, '')) 
					@@ plainto_tsquery('english', $4)
			ORDER BY keyword_score DESC
			LIMIT $3
		)
		SELECT DISTINCT ON (COALESCE(s.ticket_id, k.ticket_id))
			COALESCE(s.ticket_id, k.ticket_id) as ticket_id,
			COALESCE(s.ticket_number, k.ticket_number) as ticket_number,
			COALESCE(s.subject, k.subject) as subject,
			(COALESCE(s.semantic_score, 0) * 0.7 + COALESCE(k.keyword_score, 0) * 0.3) as similarity
		FROM semantic_results s
		FULL OUTER JOIN keyword_results k ON s.ticket_id = k.ticket_id
		ORDER BY COALESCE(s.ticket_id, k.ticket_id), similarity DESC
		LIMIT $3`

	var rows pgx.Rows
	if keywordQuery != "" {
		rows, err = s.pool.Query(ctx, searchQuery, queryEmbedding, ticketID, limit, keywordQuery)
	} else {
		// Use semantic-only if no keyword query
		rows, err = s.pool.Query(ctx, `
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
			LIMIT $3`, queryEmbedding, ticketID, limit)
	}

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

/* CreateSupportSession creates a NeuronAgent session for a support ticket */
func (s *Service) CreateSupportSession(ctx context.Context, ticketID uuid.UUID, agentID string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	// Get ticket for context
	ticket, _, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Create session configuration
	config := map[string]interface{}{
		"ticket_id":   ticketID.String(),
		"ticket_number": ticket.TicketNumber,
		"subject":     ticket.Subject,
		"customer_id": ticket.CustomerID,
		"priority":    ticket.Priority,
	}

	// Create session via NeuronAgent
	session, err := s.agentClient.CreateSession(ctx, agentID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Store session ID in ticket metadata (if metadata field exists)
	// This would require updating the ticket schema or using a separate mapping table

	return session, nil
}

/* AddMessageToSession adds a message to a NeuronAgent session */
func (s *Service) AddMessageToSession(ctx context.Context, sessionID string, role string, content string, metadata map[string]interface{}) error {
	if s.agentClient == nil {
		return fmt.Errorf("agent client not configured")
	}

	_, err := s.agentClient.CreateMessage(ctx, sessionID, role, content, metadata)
	return err
}

/* ExecuteSessionTask executes a task in a NeuronAgent session */
func (s *Service) ExecuteSessionTask(ctx context.Context, sessionID string, task string, tools []string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	return s.agentClient.ExecuteSession(ctx, sessionID, task, tools)
}

/* GetSessionMessages retrieves messages from a NeuronAgent session */
func (s *Service) GetSessionMessages(ctx context.Context, sessionID string, limit int, offset int) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	return s.agentClient.GetMessages(ctx, sessionID, limit, offset)
}

/* StreamSessionMessages streams messages from a NeuronAgent session */
func (s *Service) StreamSessionMessages(ctx context.Context, sessionID string, handler func(map[string]interface{}) error) error {
	if s.agentClient == nil {
		return fmt.Errorf("agent client not configured")
	}

	return s.agentClient.StreamMessages(ctx, sessionID, handler)
}

/* UpdateSupportSession updates a support session configuration */
func (s *Service) UpdateSupportSession(ctx context.Context, sessionID string, updates map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	session, err := s.agentClient.UpdateSession(ctx, sessionID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

/* ListSupportSessions lists all support sessions for a ticket or agent */
func (s *Service) ListSupportSessions(ctx context.Context, agentID string, filters map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	sessions, err := s.agentClient.ListSessions(ctx, agentID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}

/* CreateSupportSnapshot creates a snapshot of support session state for replay */
func (s *Service) CreateSupportSnapshot(ctx context.Context, sessionID string, ticketID uuid.UUID, userMessage string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	// Get ticket for context
	ticket, conversations, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}

	// Build execution state
	executionState := map[string]interface{}{
		"ticket_id":    ticketID.String(),
		"ticket_number": ticket.TicketNumber,
		"subject":      ticket.Subject,
		"customer_id":  ticket.CustomerID,
		"conversations": len(conversations),
	}

	// Create snapshot
	snapshot, err := s.agentClient.CreateSnapshot(ctx, sessionID, "", userMessage, executionState)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return snapshot, nil
}

/* ReplaySupportSession replays a support session from a snapshot */
func (s *Service) ReplaySupportSession(ctx context.Context, snapshotID string, options map[string]interface{}) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	result, err := s.agentClient.ReplaySession(ctx, snapshotID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to replay session: %w", err)
	}

	return result, nil
}

/* GetSupportSnapshot retrieves a support session snapshot */
func (s *Service) GetSupportSnapshot(ctx context.Context, snapshotID string) (map[string]interface{}, error) {
	if s.agentClient == nil {
		return nil, fmt.Errorf("agent client not configured")
	}

	snapshot, err := s.agentClient.GetSnapshot(ctx, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return snapshot, nil
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

/* BatchCreateMemories creates multiple memory entries with batch embedding generation */
func (s *Service) BatchCreateMemories(ctx context.Context, memories []struct {
	CustomerID      string
	MemoryType      string
	MemoryContent   string
	ImportanceScore float64
}) ([]MemoryEntry, error) {
	if len(memories) == 0 {
		return []MemoryEntry{}, nil
	}

	// Extract all memory contents for batch embedding
	memoryContents := make([]string, len(memories))
	for i, mem := range memories {
		memoryContents[i] = mem.MemoryContent
	}

	// Generate embeddings in batch
	modelName := "sentence-transformers/all-MiniLM-L6-v2"
	embeddings, err := s.neurondbClient.BatchGenerateEmbedding(ctx, memoryContents, modelName)
	if err != nil {
		// Fallback to individual embeddings if batch fails
		embeddings = make([]string, len(memories))
		for i, content := range memoryContents {
			emb, err := s.neurondbClient.GenerateEmbedding(ctx, content, modelName)
			if err != nil {
				embeddings[i] = "" // Continue without embedding
			} else {
				embeddings[i] = emb
			}
		}
	}

	// Insert all memories
	query := `
		INSERT INTO neuronip.support_memory 
		(id, customer_id, memory_type, memory_content, embedding, importance_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5::vector, $6, $7, $8)
		RETURNING id, customer_id, memory_type, memory_content, importance_score, created_at, updated_at`

	now := time.Now()
	var createdMemories []MemoryEntry

	for i, mem := range memories {
		memID := uuid.New()
		embedding := ""
		if i < len(embeddings) {
			embedding = embeddings[i]
		}

		var createdMem MemoryEntry
		err := s.pool.QueryRow(ctx, query, memID, mem.CustomerID, mem.MemoryType, mem.MemoryContent,
			embedding, mem.ImportanceScore, now, now).Scan(
			&createdMem.ID, &createdMem.CustomerID, &createdMem.MemoryType, &createdMem.MemoryContent,
			&createdMem.ImportanceScore, &createdMem.CreatedAt, &createdMem.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create memory %d: %w", i, err)
		}

		createdMemories = append(createdMemories, createdMem)
	}

	return createdMemories, nil
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
