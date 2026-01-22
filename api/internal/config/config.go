package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

/* Config holds application configuration */
type Config struct {
	Database      DatabaseConfig
	Server        ServerConfig
	Logging       LoggingConfig
	CORS          CORSConfig
	Auth          AuthConfig
	NeuronDB      NeuronDBConfig
	NeuronAgent   NeuronAgentConfig
	NeuronMCP     NeuronMCPConfig
	Observability ObservabilityConfig
	RateLimit     RateLimitConfig
}

/* DatabaseConfig holds database configuration */
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	TenancyMode     string
}

/* ServerConfig holds server configuration */
type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

/* LoggingConfig holds logging configuration */
type LoggingConfig struct {
	Level  string
	Format string
	Output string
}

/* CORSConfig holds CORS configuration */
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

/* AuthConfig holds authentication configuration */
type AuthConfig struct {
	JWTSecret       string
	EnableAPIKeys   bool
	SCIMSecret      string
	Session         SessionConfig
	NeuronAIDemo    DatabaseConfig
}

/* SessionConfig holds session configuration */
type SessionConfig struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CookieDomain    string
	CookieSecure    bool
	CookieSameSite  string
}

/* NeuronDBConfig holds NeuronDB connection configuration */
type NeuronDBConfig struct {
	Host              string
	Port              string
	Database          string
	User              string
	Password          string
	EnableVectorOps   bool
	EnableMLOps       bool
	EnableRAGOps      bool
	EnableMultimodal  bool
	EnableImageEmbeddings bool
	EnableBatchOps    bool
	EnableVectorIndexing bool
	AutoCreateIndexes bool
	DefaultIndexType  string // "hnsw" or "ivf"
}

/* NeuronAgentConfig holds NeuronAgent connection configuration */
type NeuronAgentConfig struct {
	Endpoint         string
	APIKey           string
	EnableSessions   bool
	EnableWorkflows  bool
	EnableEvaluation bool
	EnableReplay     bool
	EnableSpecializations bool
	SessionTimeout   time.Duration
	WorkflowTimeout  time.Duration
}

/* NeuronMCPConfig holds NeuronMCP configuration */
type NeuronMCPConfig struct {
	BinaryPath          string
	ToolCategories      []string
	EnableVectorOps     bool
	EnableMLTools       bool
	EnableRAGTools      bool
	EnablePostgresTools bool
	EnableAnalytics     bool
	EnableReranking     bool
	EnableIndexMgmt     bool
	EnableAutoML        bool
	EnableDataDrift     bool
	EnableQueryOptimization bool
	Timeout             time.Duration
	MaxConcurrentTools  int
	CacheResults        bool
}

/* ObservabilityConfig holds observability configuration */
type ObservabilityConfig struct {
	EnableTracing bool
}

/* RateLimitConfig holds rate limiting configuration */
type RateLimitConfig struct {
	Enabled      bool
	MaxRequests  int
	Window       time.Duration
}

/* Load loads configuration from environment variables */
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "neuronip"),
			Password:        getEnv("DB_PASSWORD", "neuronip"),
			Name:            getEnv("DB_NAME", "neuronip"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnv("SERVER_PORT", "8082"),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods: getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", ""),
			EnableAPIKeys: getEnv("ENABLE_API_KEYS", "true") == "true",
			SCIMSecret:    getEnv("SCIM_SECRET", ""),
			Session: SessionConfig{
				AccessTokenTTL:  getEnvDuration("SESSION_ACCESS_TTL", 15*time.Minute),
				RefreshTokenTTL:  getEnvDuration("SESSION_REFRESH_TTL", 7*24*time.Hour),
				CookieDomain:     getEnv("SESSION_COOKIE_DOMAIN", ""),
				CookieSecure:     getEnv("SESSION_COOKIE_SECURE", "false") == "true",
				CookieSameSite:   getEnv("SESSION_COOKIE_SAME_SITE", "Lax"),
			},
			NeuronAIDemo: DatabaseConfig{
				Host:     getEnv("NEURONAI_DEMO_HOST", "localhost"),
				Port:     getEnv("NEURONAI_DEMO_PORT", "5432"),
				User:     getEnv("NEURONAI_DEMO_USER", "neurondb"),
				Password: getEnv("NEURONAI_DEMO_PASSWORD", "neurondb"),
				Name:     getEnv("NEURONAI_DEMO_DATABASE", "neuronai-demo"),
			},
		},
		Observability: ObservabilityConfig{
			EnableTracing: getEnv("ENABLE_TRACING", "false") == "true",
		},
		RateLimit: RateLimitConfig{
			Enabled:     getEnv("RATE_LIMIT_ENABLED", "true") == "true",
			MaxRequests: getEnvInt("RATE_LIMIT_MAX_REQUESTS", 1000),
			Window:      getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Hour),
		},
		NeuronDB: NeuronDBConfig{
			Host:              getEnv("NEURONDB_HOST", "localhost"),
			Port:              getEnv("NEURONDB_PORT", "5433"),
			Database:          getEnv("NEURONDB_DATABASE", "neurondb"),
			User:              getEnv("NEURONDB_USER", "neurondb"),
			Password:          getEnv("NEURONDB_PASSWORD", "neurondb"),
			EnableVectorOps:   getEnv("NEURONDB_ENABLE_VECTOR_OPS", "true") == "true",
			EnableMLOps:       getEnv("NEURONDB_ENABLE_ML_OPS", "true") == "true",
			EnableRAGOps:       getEnv("NEURONDB_ENABLE_RAG_OPS", "true") == "true",
			EnableMultimodal:  getEnv("NEURONDB_ENABLE_MULTIMODAL", "true") == "true",
			EnableImageEmbeddings: getEnv("NEURONDB_ENABLE_IMAGE_EMBEDDINGS", "true") == "true",
			EnableBatchOps:    getEnv("NEURONDB_ENABLE_BATCH_OPS", "true") == "true",
			EnableVectorIndexing: getEnv("NEURONDB_ENABLE_VECTOR_INDEXING", "true") == "true",
			AutoCreateIndexes: getEnv("NEURONDB_AUTO_CREATE_INDEXES", "true") == "true",
			DefaultIndexType:  getEnv("NEURONDB_DEFAULT_INDEX_TYPE", "hnsw"),
		},
		NeuronAgent: NeuronAgentConfig{
			Endpoint:         getEnv("NEURONAGENT_ENDPOINT", "http://localhost:8080"),
			APIKey:           getEnv("NEURONAGENT_API_KEY", ""),
			EnableSessions:   getEnv("NEURONAGENT_ENABLE_SESSIONS", "true") == "true",
			EnableWorkflows:  getEnv("NEURONAGENT_ENABLE_WORKFLOWS", "true") == "true",
			EnableEvaluation: getEnv("NEURONAGENT_ENABLE_EVALUATION", "true") == "true",
			EnableReplay:     getEnv("NEURONAGENT_ENABLE_REPLAY", "true") == "true",
			EnableSpecializations: getEnv("NEURONAGENT_ENABLE_SPECIALIZATIONS", "true") == "true",
			SessionTimeout:   getEnvDuration("NEURONAGENT_SESSION_TIMEOUT", 30*time.Minute),
			WorkflowTimeout:  getEnvDuration("NEURONAGENT_WORKFLOW_TIMEOUT", 1*time.Hour),
		},
		NeuronMCP: NeuronMCPConfig{
			BinaryPath:          getEnv("NEURONMCP_BINARY_PATH", "/usr/local/bin/neurondb-mcp"),
			ToolCategories:      getEnvSlice("NEURONMCP_TOOL_CATEGORIES", []string{"vector", "embedding", "rag", "ml", "analytics", "postgresql"}),
			EnableVectorOps:     getEnv("NEURONMCP_ENABLE_VECTOR_OPS", "true") == "true",
			EnableMLTools:       getEnv("NEURONMCP_ENABLE_ML_TOOLS", "true") == "true",
			EnableRAGTools:      getEnv("NEURONMCP_ENABLE_RAG_TOOLS", "true") == "true",
			EnablePostgresTools: getEnv("NEURONMCP_ENABLE_POSTGRES_TOOLS", "true") == "true",
			EnableAnalytics:     getEnv("NEURONMCP_ENABLE_ANALYTICS", "true") == "true",
			EnableReranking:     getEnv("NEURONMCP_ENABLE_RERANKING", "true") == "true",
			EnableIndexMgmt:     getEnv("NEURONMCP_ENABLE_INDEX_MGMT", "true") == "true",
			EnableAutoML:       getEnv("NEURONMCP_ENABLE_AUTOML", "true") == "true",
			EnableDataDrift:     getEnv("NEURONMCP_ENABLE_DATA_DRIFT", "true") == "true",
			EnableQueryOptimization: getEnv("NEURONMCP_ENABLE_QUERY_OPTIMIZATION", "true") == "true",
			Timeout:             getEnvDuration("NEURONMCP_TIMEOUT", 30*time.Second),
			MaxConcurrentTools:  getEnvInt("NEURONMCP_MAX_CONCURRENT_TOOLS", 10),
			CacheResults:        getEnv("NEURONMCP_CACHE_RESULTS", "true") == "true",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := []string{}
		for _, part := range splitString(value, ",") {
			parts = append(parts, trimSpace(part))
		}
		if len(parts) > 0 {
			return parts
		}
	}
	return defaultValue
}

func splitString(s, sep string) []string {
	parts := []string{}
	current := ""
	for _, char := range s {
		if string(char) == sep {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

/* DSN returns the database connection string */
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name)
}
