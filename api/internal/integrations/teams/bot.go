package teams

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/neurondb/NeuronIP/api/internal/agent"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
)

/* TeamsBotService provides Microsoft Teams bot functionality */
type TeamsBotService struct {
	agentClient   *agent.Client
	neurondbClient *neurondb.Client
	appID         string
	appPassword   string
}

/* NewTeamsBotService creates a new Teams bot service */
func NewTeamsBotService(agentClient *agent.Client, neurondbClient *neurondb.Client, appID string, appPassword string) *TeamsBotService {
	return &TeamsBotService{
		agentClient:   agentClient,
		neurondbClient: neurondbClient,
		appID:         appID,
		appPassword:   appPassword,
	}
}

/* TeamsMessage represents a Teams message */
type TeamsMessage struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	From    struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"from"`
	ChannelID string `json:"channelId"`
}

/* HandleMessage handles Teams messages */
func (s *TeamsBotService) HandleMessage(ctx context.Context, msg TeamsMessage) (string, error) {
	if msg.Type != "message" {
		return "", nil
	}

	question := strings.TrimSpace(msg.Text)
	if question == "" {
		return "Please ask a question", nil
	}

	// Execute query using agent
	result, err := s.agentClient.ExecuteAgent(ctx, "default", question, nil, nil)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	// Format response for Teams
	answer, _ := result["response"].(string)
	if answer == "" {
		answer, _ = result["answer"].(string)
	}
	if answer == "" {
		answer = "I couldn't generate a response. Please try again."
	}
	
	response := fmt.Sprintf("**Question:** %s\n\n**Answer:** %s", question, answer)
	
	// Add citations if available
	if citations, ok := result["citations"].([]interface{}); ok && len(citations) > 0 {
		response += "\n\n**Sources:**\n"
		for i, citation := range citations {
			response += fmt.Sprintf("%d. %v\n", i+1, citation)
		}
	}

	return response, nil
}

/* HandleHTTPRequest handles HTTP requests from Teams */
func (s *TeamsBotService) HandleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg TeamsMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	response, err := s.HandleMessage(r.Context(), msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"type": "message",
		"text": response,
	})
}
