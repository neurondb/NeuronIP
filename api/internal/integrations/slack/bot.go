package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/neurondb/NeuronIP/api/internal/agent"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* SlackBotService provides Slack bot functionality */
type SlackBotService struct {
	agentClient   *agent.Client
	neurondbClient *neurondb.Client
	workspaceToken string
}

/* NewSlackBotService creates a new Slack bot service */
func NewSlackBotService(agentClient *agent.Client, neurondbClient *neurondb.Client, workspaceToken string) *SlackBotService {
	return &SlackBotService{
		agentClient:   agentClient,
		neurondbClient: neurondbClient,
		workspaceToken: workspaceToken,
	}
}

/* SlackCommand represents a Slack slash command */
type SlackCommand struct {
	Token       string `json:"token"`
	TeamID      string `json:"team_id"`
	TeamDomain  string `json:"team_domain"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Command     string `json:"command"`
	Text        string `json:"text"`
	ResponseURL string `json:"response_url"`
}

/* HandleSlashCommand handles Slack slash commands */
func (s *SlackBotService) HandleSlashCommand(ctx context.Context, cmd SlackCommand) (string, error) {
	// Verify token
	if cmd.Token != s.workspaceToken {
		return "", fmt.Errorf("invalid token")
	}

	// Parse command
	parts := strings.Fields(cmd.Text)
	if len(parts) == 0 {
		return "Usage: /neuronip <question>", nil
	}

	question := strings.Join(parts, " ")

	// Execute query using agent
	result, err := s.agentClient.ExecuteAgent(ctx, "default", question, nil, nil)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	// Format response for Slack
	answer, _ := result["response"].(string)
	if answer == "" {
		answer, _ = result["answer"].(string)
	}
	if answer == "" {
		answer = "I couldn't generate a response. Please try again."
	}
	
	response := fmt.Sprintf("*Question:* %s\n*Answer:* %s", question, answer)
	
	// Add citations if available
	if citations, ok := result["citations"].([]interface{}); ok && len(citations) > 0 {
		response += "\n*Sources:*\n"
		for i, citation := range citations {
			response += fmt.Sprintf("%d. %v\n", i+1, citation)
		}
	}

	return response, nil
}

/* SendMessage sends a message to a Slack channel */
func (s *SlackBotService) SendMessage(ctx context.Context, channelID string, message string) error {
	// In production, would use Slack Web API
	// For now, return nil (would implement actual Slack API call)
	return nil
}

/* HandleHTTPRequest handles HTTP requests from Slack */
func (s *SlackBotService) HandleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmd SlackCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	response, err := s.HandleSlashCommand(r.Context(), cmd)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"response_type": "in_channel",
		"text":          response,
	})
}
