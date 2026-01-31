package testutil

import (
	"time"

	"github.com/google/uuid"
)

// TestFixtures contains common test data structures

type TestUser struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Name      string
	CreatedAt time.Time
}

type TestAgent struct {
	ID          uuid.UUID
	Name        string
	Status      string
	Description string
	CreatedAt   time.Time
}

type TestWorkflow struct {
	ID          uuid.UUID
	Name        string
	Status      string
	Description string
	CreatedAt   time.Time
}

type TestModel struct {
	ID          uuid.UUID
	Name        string
	Type        string
	Status      string
	Description string
}

type TestDataSource struct {
	ID             uuid.UUID
	Name           string
	Type           string
	Status         string
	ConnectionInfo map[string]interface{}
}

type TestMetric struct {
	ID         uuid.UUID
	Name       string
	Category   string
	Status     string
	Definition string
}

// GetTestUser returns a test user fixture
func GetTestUser() TestUser {
	return TestUser{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Password:  "testpassword123",
		Name:      "Test User",
		CreatedAt: time.Now(),
	}
}

// GetTestAgent returns a test agent fixture
func GetTestAgent() TestAgent {
	return TestAgent{
		ID:          uuid.New(),
		Name:        "Test Agent",
		Status:      "active",
		Description: "A test agent for testing",
		CreatedAt:   time.Now(),
	}
}

// GetTestWorkflow returns a test workflow fixture
func GetTestWorkflow() TestWorkflow {
	return TestWorkflow{
		ID:          uuid.New(),
		Name:        "Test Workflow",
		Status:      "draft",
		Description: "A test workflow",
		CreatedAt:   time.Now(),
	}
}

// GetTestModel returns a test model fixture
func GetTestModel() TestModel {
	return TestModel{
		ID:          uuid.New(),
		Name:        "Test Model",
		Type:        "llm",
		Status:      "active",
		Description: "A test model",
	}
}

// GetTestDataSource returns a test data source fixture
func GetTestDataSource() TestDataSource {
	return TestDataSource{
		ID:   uuid.New(),
		Name: "Test Database",
		Type: "postgresql",
		Status: "connected",
		ConnectionInfo: map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"database": "testdb",
		},
	}
}

// GetTestMetric returns a test metric fixture
func GetTestMetric() TestMetric {
	return TestMetric{
		ID:         uuid.New(),
		Name:       "Test Metric",
		Category:   "financial",
		Status:     "draft",
		Definition: "A test metric definition",
	}
}
