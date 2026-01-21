package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/neurondb/NeuronIP/api/internal/db"
)

/* SCIMService provides SCIM 2.0 user provisioning functionality */
type SCIMService struct {
	queries *db.Queries
	secret  string // SCIM bearer token
}

/* NewSCIMService creates a new SCIM service */
func NewSCIMService(queries *db.Queries, secret string) *SCIMService {
	return &SCIMService{
		queries: queries,
		secret:  secret,
	}
}

/* ValidateSCIMToken validates SCIM bearer token */
func (s *SCIMService) ValidateSCIMToken(token string) bool {
	return token == s.secret
}

/* SCIMUser represents a SCIM 2.0 user resource */
type SCIMUser struct {
	Schemas    []string               `json:"schemas"`
	ID         string                 `json:"id,omitempty"`
	ExternalID string                 `json:"externalId,omitempty"`
	UserName   string                 `json:"userName"`
	Name       SCIMUserName           `json:"name,omitempty"`
	Emails     []SCIMUserEmail        `json:"emails"`
	Active     bool                   `json:"active"`
	Meta       SCIMUserMeta           `json:"meta,omitempty"`
	Groups     []string               `json:"groups,omitempty"`
	Extensions map[string]interface{} `json:"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User,omitempty"`
}

/* SCIMUserName represents SCIM user name */
type SCIMUserName struct {
	Formatted  string `json:"formatted,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	GivenName  string `json:"givenName,omitempty"`
	MiddleName string `json:"middleName,omitempty"`
}

/* SCIMUserEmail represents SCIM user email */
type SCIMUserEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

/* SCIMUserMeta represents SCIM user metadata */
type SCIMUserMeta struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created,omitempty"`
	LastModified time.Time `json:"lastModified,omitempty"`
	Location     string    `json:"location,omitempty"`
	Version      string    `json:"version,omitempty"`
}

/* SCIMListResponse represents SCIM list response */
type SCIMListResponse struct {
	TotalResults int        `json:"totalResults"`
	ItemsPerPage int        `json:"itemsPerPage"`
	StartIndex   int        `json:"startIndex"`
	Schemas      []string   `json:"schemas"`
	Resources    []SCIMUser `json:"Resources"`
}

/* CreateUser creates a user via SCIM */
func (s *SCIMService) CreateUser(ctx context.Context, scimUser SCIMUser) (*SCIMUser, error) {
	// Extract email (primary email or first email)
	email := ""
	for _, e := range scimUser.Emails {
		if e.Primary || email == "" {
			email = e.Value
		}
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Extract name
	name := scimUser.UserName
	if scimUser.Name.Formatted != "" {
		name = scimUser.Name.Formatted
	} else if scimUser.Name.GivenName != "" {
		name = scimUser.Name.GivenName
		if scimUser.Name.FamilyName != "" {
			name += " " + scimUser.Name.FamilyName
		}
	}

	// Determine role from groups or extensions
	role := "analyst" // Default
	if scimUser.Groups != nil && len(scimUser.Groups) > 0 {
		// Map groups to roles (simplified)
		for _, group := range scimUser.Groups {
			if group == "admin" || group == "administrators" {
				role = "admin"
				break
			}
		}
	}

	// Create user in database
	user, err := s.queries.CreateUser(ctx, email, &name, nil, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Build SCIM response
	result := SCIMUser{
		Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		ID:       user.ID.String(),
		UserName: scimUser.UserName,
		Name: SCIMUserName{
			Formatted: name,
		},
		Emails: []SCIMUserEmail{
			{
				Value:   email,
				Primary: true,
			},
		},
		Active: scimUser.Active,
		Meta: SCIMUserMeta{
			ResourceType: "User",
			Created:      user.CreatedAt,
			LastModified: user.UpdatedAt,
			Location:     fmt.Sprintf("/Users/%s", user.ID.String()),
			Version:      fmt.Sprintf("W/\"%s\"", user.UpdatedAt.Format(time.RFC3339)),
		},
	}

	return &result, nil
}

/* GetUser retrieves a user via SCIM */
func (s *SCIMService) GetUser(ctx context.Context, userID string) (*SCIMUser, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	name := ""
	if user.Name != nil {
		name = *user.Name
	}

	result := SCIMUser{
		Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		ID:       user.ID.String(),
		UserName: user.Email,
		Name: SCIMUserName{
			Formatted: name,
		},
		Emails: []SCIMUserEmail{
			{
				Value:   user.Email,
				Primary: true,
			},
		},
		Active: true, // Assume active if user exists
		Meta: SCIMUserMeta{
			ResourceType: "User",
			Created:      user.CreatedAt,
			LastModified: user.UpdatedAt,
			Location:     fmt.Sprintf("/Users/%s", user.ID.String()),
			Version:      fmt.Sprintf("W/\"%s\"", user.UpdatedAt.Format(time.RFC3339)),
		},
	}

	return &result, nil
}

/* UpdateUser updates a user via SCIM */
func (s *SCIMService) UpdateUser(ctx context.Context, userID string, scimUser SCIMUser) (*SCIMUser, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update email if provided
	if len(scimUser.Emails) > 0 {
		newEmail := scimUser.Emails[0].Value
		if newEmail != user.Email {
			// Update email - requires direct SQL since UpdateUser doesn't support email updates
			query := `UPDATE neuronip.users SET email = $1, updated_at = NOW() WHERE id = $2`
			_, err := s.queries.DB.Exec(ctx, query, newEmail, id)
			if err != nil {
				return nil, fmt.Errorf("failed to update email: %w", err)
			}
		}
	}

	// Update name if provided
	if scimUser.Name.Formatted != "" {
		name := scimUser.Name.Formatted
		err := s.queries.UpdateUser(ctx, id, &name, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
	}

	// Update active status (in SCIM, active=false means deactivated)
	if !scimUser.Active {
		// Mark user as inactive - could use a status field or soft delete
		// For now, update role to indicate inactive
		query := `UPDATE neuronip.users SET role = 'inactive', updated_at = NOW() WHERE id = $1`
		_, err := s.queries.DB.Exec(ctx, query, id)
		if err != nil {
			return nil, fmt.Errorf("failed to deactivate user: %w", err)
		}
	} else if user.Role == "inactive" {
		// Reactivate user
		query := `UPDATE neuronip.users SET role = 'analyst', updated_at = NOW() WHERE id = $1`
		_, err := s.queries.DB.Exec(ctx, query, id)
		if err != nil {
			return nil, fmt.Errorf("failed to reactivate user: %w", err)
		}
	}

	// Get updated user
	updatedUser, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	name := ""
	if updatedUser.Name != nil {
		name = *updatedUser.Name
	}

	result := SCIMUser{
		Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		ID:       updatedUser.ID.String(),
		UserName: updatedUser.Email,
		Name: SCIMUserName{
			Formatted: name,
		},
		Emails: []SCIMUserEmail{
			{
				Value:   updatedUser.Email,
				Primary: true,
			},
		},
		Active: scimUser.Active,
		Meta: SCIMUserMeta{
			ResourceType: "User",
			Created:      updatedUser.CreatedAt,
			LastModified: updatedUser.UpdatedAt,
			Location:     fmt.Sprintf("/Users/%s", updatedUser.ID.String()),
			Version:      fmt.Sprintf("W/\"%s\"", updatedUser.UpdatedAt.Format(time.RFC3339)),
		},
	}

	return &result, nil
}

/* DeleteUser deletes a user via SCIM */
func (s *SCIMService) DeleteUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	// Delete user - soft delete by marking as inactive, or hard delete
	// For safety, we'll use soft delete by updating role
	query := `UPDATE neuronip.users SET role = 'deleted', updated_at = NOW() WHERE id = $1`
	_, err = s.queries.DB.Exec(ctx, query, id)
	if err != nil {
		// If soft delete fails or hard delete preferred, use direct delete
		deleteQuery := `DELETE FROM neuronip.users WHERE id = $1`
		_, err = s.queries.DB.Exec(ctx, deleteQuery, id)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
	}

	return nil
}

/* ListUsers lists users via SCIM with pagination */
func (s *SCIMService) ListUsers(ctx context.Context, startIndex, count int, filter string) (*SCIMListResponse, error) {
	// Parse filter (simplified - full SCIM filter parsing would be more complex)
	// Build query with filtering and pagination
	query := `SELECT id, email, email_verified, password_hash, name, avatar_url, role, 
	          two_factor_enabled, two_factor_secret, preferences, last_login_at, created_at, updated_at 
	          FROM neuronip.users WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	// Simple filter parsing - supports "email eq 'value'" or "userName eq 'value'"
	if filter != "" {
		emailRe := regexp.MustCompile(`(email|userName)\s+eq\s+['"]([^'"]+)['"]`)
		matches := emailRe.FindStringSubmatch(filter)
		if len(matches) > 2 {
			email := matches[2]
			query += fmt.Sprintf(" AND email = $%d", argIndex)
			args = append(args, email)
			argIndex++
		}
	}
	
	// Exclude deleted users
	query += " AND role != 'deleted'"
	
	// Order and paginate
	query += " ORDER BY created_at DESC"
	if count > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, count)
		argIndex++
		if startIndex > 1 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, startIndex-1)
		}
	} else {
		query += " LIMIT 100" // Default limit
	}
	
	rows, err := s.queries.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()
	
	users := []*db.User{}

	// Scan users from database
	for rows.Next() {
		user := &db.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.EmailVerified, &user.PasswordHash,
			&user.Name, &user.AvatarURL, &user.Role, &user.TwoFactorEnabled,
			&user.TwoFactorSecret, &user.Preferences, &user.LastLoginAt,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	
	// Get total count for pagination
	countQuery := `SELECT COUNT(*) FROM neuronip.users WHERE role != 'deleted'`
	var totalResults int
	if filter != "" {
		emailRe := regexp.MustCompile(`(email|userName)\s+eq\s+['"]([^'"]+)['"]`)
		matches := emailRe.FindStringSubmatch(filter)
		if len(matches) > 2 {
			countQuery += fmt.Sprintf(" AND email = '%s'", matches[2])
		}
	}
	s.queries.DB.QueryRow(ctx, countQuery).Scan(&totalResults)

	// Convert to SCIM format
	resources := make([]SCIMUser, 0, len(users))
	for _, user := range users {
		name := ""
		if user.Name != nil {
			name = *user.Name
		}

		scimUser := SCIMUser{
			Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
			ID:       user.ID.String(),
			UserName: user.Email,
			Name: SCIMUserName{
				Formatted: name,
			},
			Emails: []SCIMUserEmail{
				{
					Value:   user.Email,
					Primary: true,
				},
			},
			Active: true,
			Meta: SCIMUserMeta{
				ResourceType: "User",
				Created:      user.CreatedAt,
				LastModified: user.UpdatedAt,
				Location:     fmt.Sprintf("/Users/%s", user.ID.String()),
			},
		}
		resources = append(resources, scimUser)
	}

	// Pagination is already handled in SQL query, but adjust if needed
	if startIndex > 0 && startIndex <= totalResults {
		// Resources are already correctly ordered and paginated from query
	}
	if count > 0 && count < len(resources) {
		resources = resources[:count]
	}

	return &SCIMListResponse{
		TotalResults: totalResults,
		ItemsPerPage: count,
		StartIndex:   startIndex,
		Schemas:      []string{"urn:ietf:params:scim:schemas:core:2.0:ListResponse"},
		Resources:    resources,
	}, nil
}

/* HandleSCIMRequest handles SCIM HTTP requests */
func (s *SCIMService) HandleSCIMRequest(w http.ResponseWriter, r *http.Request) {
	// Validate bearer token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !s.ValidateSCIMToken(authHeader) {
		w.Header().Set("Content-Type", "application/scim+json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"detail": "Unauthorized",
		})
		return
	}

	ctx := r.Context()
	path := r.URL.Path

	switch r.Method {
	case http.MethodGet:
		if path == "/Users" {
			// List users
			startIndex, _ := strconv.Atoi(r.URL.Query().Get("startIndex"))
			count, _ := strconv.Atoi(r.URL.Query().Get("count"))
			filter := r.URL.Query().Get("filter")

			if startIndex == 0 {
				startIndex = 1
			}
			if count == 0 {
				count = 100
			}

			response, err := s.ListUsers(ctx, startIndex, count, filter)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
				return
			}

			w.Header().Set("Content-Type", "application/scim+json")
			json.NewEncoder(w).Encode(response)
		} else {
			// Get single user
			userID := path[len("/Users/"):]
			user, err := s.GetUser(ctx, userID)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"detail": "User not found"})
				return
			}

			w.Header().Set("Content-Type", "application/scim+json")
			json.NewEncoder(w).Encode(user)
		}

	case http.MethodPost:
		// Create user
		var scimUser SCIMUser
		if err := json.NewDecoder(r.Body).Decode(&scimUser); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"detail": "Invalid request body"})
			return
		}

		user, err := s.CreateUser(ctx, scimUser)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"detail": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/scim+json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)

	case http.MethodPut:
		// Update user
		userID := path[len("/Users/"):]
		var scimUser SCIMUser
		if err := json.NewDecoder(r.Body).Decode(&scimUser); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"detail": "Invalid request body"})
			return
		}

		user, err := s.UpdateUser(ctx, userID, scimUser)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"detail": "User not found"})
			return
		}

		w.Header().Set("Content-Type", "application/scim+json")
		json.NewEncoder(w).Encode(user)

	case http.MethodDelete:
		// Delete user
		userID := path[len("/Users/"):]
		if err := s.DeleteUser(ctx, userID); err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"detail": "User not found"})
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
