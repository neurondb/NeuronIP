package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/neurondb/NeuronIP/api/internal/agent"
	"github.com/neurondb/NeuronIP/api/internal/agents"
	"github.com/neurondb/NeuronIP/api/internal/ai"
	"github.com/neurondb/NeuronIP/api/internal/rag"
	"github.com/neurondb/NeuronIP/api/internal/alerts"
	"github.com/neurondb/NeuronIP/api/internal/analytics"
	"github.com/neurondb/NeuronIP/api/internal/audit"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/backup"
	"github.com/neurondb/NeuronIP/api/internal/billing"
	"github.com/neurondb/NeuronIP/api/internal/catalog"
	"github.com/neurondb/NeuronIP/api/internal/classification"
	"github.com/neurondb/NeuronIP/api/internal/comments"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/config"
	"github.com/neurondb/NeuronIP/api/internal/connectors"
	"github.com/neurondb/NeuronIP/api/internal/dataquality"
	"github.com/neurondb/NeuronIP/api/internal/datasources"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/handlers"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
	ingestionconnectors "github.com/neurondb/NeuronIP/api/internal/ingestion/connectors"
	"github.com/neurondb/NeuronIP/api/internal/integrations"
	"github.com/neurondb/NeuronIP/api/internal/knowledgegraph"
	"github.com/neurondb/NeuronIP/api/internal/lineage"
	"github.com/neurondb/NeuronIP/api/internal/logging"
	"github.com/neurondb/NeuronIP/api/internal/masking"
	"github.com/neurondb/NeuronIP/api/internal/mcp"
	"github.com/neurondb/NeuronIP/api/internal/metrics"
	"github.com/neurondb/NeuronIP/api/internal/middleware"
	"github.com/neurondb/NeuronIP/api/internal/models"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
	"github.com/neurondb/NeuronIP/api/internal/observability"
	"github.com/neurondb/NeuronIP/api/internal/ownership"
	"github.com/neurondb/NeuronIP/api/internal/policy"
	"github.com/neurondb/NeuronIP/api/internal/profiling"
	"github.com/neurondb/NeuronIP/api/internal/semantic"
	"github.com/neurondb/NeuronIP/api/internal/support"
	"github.com/neurondb/NeuronIP/api/internal/tenancy"
	"github.com/neurondb/NeuronIP/api/internal/tracing"
	"github.com/neurondb/NeuronIP/api/internal/versioning"
	"github.com/neurondb/NeuronIP/api/internal/warehouse"
	"github.com/neurondb/NeuronIP/api/internal/webhooks"
	"github.com/neurondb/NeuronIP/api/internal/workflows"
	"github.com/neurondb/NeuronIP/api/internal/session"
	"github.com/neurondb/NeuronIP/api/internal/execution"
	"github.com/neurondb/NeuronIP/api/internal/collaboration"
	slackbot "github.com/neurondb/NeuronIP/api/internal/integrations/slack"
	teamsbot "github.com/neurondb/NeuronIP/api/internal/integrations/teams"
	bibot "github.com/neurondb/NeuronIP/api/internal/integrations/bi"
)

var (
	version   = "dev"
	buildDate = "unknown"
	gitCommit = "unknown"
)

func main() {
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		showHelp    = flag.Bool("help", false, "Show help message")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "NeuronIP API Server - Enterprise Intelligence Platform\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("neuronip-api version %s\n", version)
		fmt.Printf("Build date: %s\n", buildDate)
		fmt.Printf("Git commit: %s\n", gitCommit)
		os.Exit(0)
	}

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging
	logging.InitLogger(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output)
	logger := logging.DefaultLogger
	if logger == nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger\n")
		os.Exit(1)
	}

	logger.Info("Starting NeuronIP API server",
		"version", version,
		"build_date", buildDate,
		"git_commit", gitCommit,
	)

	// Create database connection pools (multi-database support)
	ctx := context.Background()
	multiPool, err := db.NewMultiPool(ctx, *cfg)
	if err != nil {
		logger.Error("Failed to create database pools", "error", err)
		os.Exit(1)
	}
	defer multiPool.Close()

	// Get default neuronip pool for backward compatibility
	pool, err := multiPool.GetPool("neuronip")
	if err != nil {
		logger.Error("Failed to get neuronip pool", "error", err)
		os.Exit(1)
	}

	// Create Pool wrapper for backward compatibility
	poolWrapper := &db.Pool{Pool: pool}

	logger.Info("Database pools created successfully",
		"max_conns", cfg.Database.MaxOpenConns,
		"min_conns", cfg.Database.MaxIdleConns,
		"databases", []string{"neuronip", "neuronai-demo"},
	)

	// Initialize database queries (uses default pool, but queries can be context-aware)
	queries := db.NewQueries(pool)

	// Initialize NeuronDB client
	neurondbClient := neurondb.NewClient(pool)

	// Create router
	router := mux.NewRouter()

	// Apply global middleware (order matters)
	router.Use(middleware.Recovery)                   // Recover from panics
	router.Use(middleware.RequestID)                  // Add request ID
	router.Use(middleware.SecurityHeaders)            // Security headers
	router.Use(middleware.HTTPLogging)                // Request/response logging
	router.Use(middleware.CORS(middleware.CORSConfig{ // CORS
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		AllowedMethods: cfg.CORS.AllowedMethods,
		AllowedHeaders: cfg.CORS.AllowedHeaders,
	}))

	// Initialize MCP client (optional)
	var mcpClient *mcp.Client
	if cfg.NeuronMCP.BinaryPath != "" {
		mcpClient = mcp.NewClient(cfg.NeuronMCP.BinaryPath)
		logger.Info("MCP client initialized", "binary_path", cfg.NeuronMCP.BinaryPath)
	}

	// Health check endpoint (no auth required)
	var healthHandler *handlers.HealthHandler
	if mcpClient != nil {
		healthHandler = handlers.NewHealthHandlerWithMCP(pool, mcpClient)
	} else {
		healthHandler = handlers.NewHealthHandler(pool)
	}
	router.Handle("/health", healthHandler).Methods("GET")
	router.Handle("/api/v1/health", healthHandler).Methods("GET")

	// Metrics endpoint (no auth required) - Prometheus metrics
	router.Handle("/metrics", metrics.Handler()).Methods("GET")

	// API routes (require auth)
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	
	// Apply session middleware first (supports cookie-based auth)
	// Note: sessionManager is initialized later, so we'll add this after initialization
	// For now, we'll add it after sessionManager is created

	// Apply rate limiting to API routes (after auth)
	if cfg.RateLimit.Enabled {
		rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.MaxRequests, cfg.RateLimit.Window)
		apiRouter.Use(middleware.RateLimit(rateLimiter))
	}

	// Initialize agent client
	agentClient := agent.NewClient(cfg.NeuronAgent.Endpoint, cfg.NeuronAgent.APIKey)

	// Initialize services (use pool directly - services can be enhanced later to use context-aware pools)
	semanticService := semantic.NewService(queries, pool, neurondbClient, mcpClient)
	approvalService := semantic.NewApprovalService(pool)
	ownershipService := semantic.NewMetricOwnershipService(pool)
	lineageService := semantic.NewLineageService(pool)
	semanticHandler := handlers.NewSemanticHandler(semanticService, approvalService, ownershipService, lineageService)

	// Initialize pipeline service
	pipelineService := semantic.NewPipelineService(pool)
	pipelineHandler := handlers.NewPipelineHandler(pipelineService)

	// Initialize warehouse service
	warehouseService := warehouse.NewService(pool, agentClient, neurondbClient, mcpClient)
	warehouseHandler := handlers.NewWarehouseHandler(warehouseService)

	// Initialize saved search service
	savedSearchService := warehouse.NewSavedSearchService(pool)
	savedSearchHandler := handlers.NewSavedSearchHandler(savedSearchService, warehouseService)

	// Initialize governance service
	governanceService := warehouse.NewGovernanceService(pool)
	governanceHandler := handlers.NewGovernanceHandler(governanceService)

	// Initialize cache service
	cacheService := warehouse.NewCacheService(pool)
	cacheHandler := handlers.NewCacheHandler(cacheService)

	// Initialize workflow service
	workflowService := workflows.NewService(pool, agentClient, neurondbClient, mcpClient)
	workflowHandler := handlers.NewWorkflowHandler(workflowService)

	// Initialize compliance services
	complianceService := compliance.NewService(pool.Pool, neurondbClient)
	anomalyService := compliance.NewAnomalyService(pool.Pool, neurondbClient)
	policyService := compliance.NewPolicyService(pool.Pool, neurondbClient)
	complianceHandler := handlers.NewComplianceHandler(complianceService, anomalyService, policyService)

	// Initialize analytics service
	analyticsService := analytics.NewService(pool.Pool, neurondbClient, mcpClient)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	// Initialize models service
	modelsService := models.NewService(pool.Pool, neurondbClient, mcpClient)
	modelsHandler := handlers.NewModelHandler(modelsService)

	// Initialize integrations services
	integrationsService := integrations.NewIntegrationsService(pool.Pool)
	helpdeskService := integrations.NewHelpdeskService(pool.Pool)
	webhookService := integrations.NewWebhookService(pool.Pool)
	integrationHandler := handlers.NewIntegrationHandler(integrationsService, helpdeskService, webhookService)

	// Initialize alerts service
	alertsService := alerts.NewService(pool.Pool, anomalyService)
	alertsHandler := handlers.NewAlertsHandler(alertsService)

	// Initialize support service
	supportService := support.NewService(queries, pool.Pool, agentClient, neurondbClient)
	supportHandler := handlers.NewSupportHandler(supportService)

	// Initialize knowledge graph service
	knowledgeGraphService := knowledgegraph.NewService(pool.Pool, neurondbClient)
	knowledgeGraphHandler := handlers.NewKnowledgeGraphHandler(knowledgeGraphService)

	// Initialize data sources service
	dataSourceService := datasources.NewDataSourceService(pool.Pool)
	dataSourceHandler := handlers.NewDataSourceHandler(dataSourceService)

	// Initialize business metrics service (semantic layer)
	businessMetricsService := metrics.NewMetricsService(pool.Pool)

	// Initialize metrics service (metric catalog)
	metricsCatalogService := catalog.NewMetricsService(pool.Pool)
	businessMetricsHandler := handlers.NewMetricsHandler(businessMetricsService, metricsCatalogService)

	// Initialize agents service
	agentsService := agents.NewAgentsService(pool.Pool, agentClient)
	agentsHandler := handlers.NewAgentsHandler(agentsService)

	// Initialize observability service
	observabilityService := observability.NewObservabilityService(pool.Pool)
	observabilityHandler := handlers.NewObservabilityHandler(pool.Pool)
	
	// Initialize model governance handler
	modelGovernanceHandler := handlers.NewModelGovernanceHandler(pool.Pool)
	
	// Initialize collaboration handler
	collaborationHandler := handlers.NewCollaborationHandler(pool.Pool)
	
	// Initialize execution services
	replicaService := execution.NewReplicaService(pool.Pool)
	shardService := execution.NewShardService(pool.Pool)
	jobQueueService := execution.NewJobQueueService(pool.Pool)
	_ = replicaService // Used for read routing
	_ = shardService   // Used for sharding
	_ = jobQueueService // Used for job queue
	
	// Initialize quota service
	quotaService := tenancy.NewQuotaService(pool.Pool)
	_ = quotaService // Used for resource limits
	
	// Initialize row security service
	rowSecurityService := auth.NewRowSecurityService(queries)
	
	// Initialize policy-aware service
	policyAwareService := semantic.NewPolicyAwareService(pool.Pool, rowSecurityService)
	_ = policyAwareService // Used for policy-aware retrieval
	
	// Initialize Slack bot service
	slackToken := os.Getenv("SLACK_BOT_TOKEN")
	slackBotService := slackbot.NewSlackBotService(agentClient, neurondbClient, slackToken)
	
	// Initialize Teams bot service
	teamsAppID := os.Getenv("TEAMS_APP_ID")
	teamsAppPassword := os.Getenv("TEAMS_APP_PASSWORD")
	teamsBotService := teamsbot.NewTeamsBotService(agentClient, neurondbClient, teamsAppID, teamsAppPassword)
	
	// Initialize BI export service
	biExportService := bibot.NewBIExportService(warehouseService)

	// Initialize metrics collector
	metricsCollector := metrics.NewMetricsCollector(pool.Pool)
	metricsEnhancedHandler := handlers.NewMetricsEnhancedHandler(metricsCollector)

	// Initialize tracing service
	tracerService := tracing.NewTracerService(cfg.Observability.EnableTracing)
	router.Use(middleware.Tracing(tracerService))

	// Apply timeout middleware to API routes
	timeoutConfig := middleware.DefaultTimeoutConfig()
	apiRouter.Use(middleware.TimeoutByRoute(timeoutConfig))

	// Initialize lineage service
	lineageService := lineage.NewLineageService(pool.Pool)
	lineageHandler := handlers.NewLineageHandler(lineageService)

	// Audit service is initialized above with new audit package

	// Initialize billing service
	billingService := billing.NewBillingService(pool.Pool)
	billingHandler := handlers.NewBillingHandler(billingService)

	// Initialize versioning service
	versioningService := versioning.NewVersioningService(pool.Pool)
	versioningHandler := handlers.NewVersioningHandler(versioningService)

	// Initialize catalog service with NeuronDB client for multimodal embeddings
	catalogService := catalog.NewCatalogServiceWithNeuronDB(pool.Pool, neurondbClient)
	catalogHandler := handlers.NewCatalogHandler(catalogService)

	// Initialize semantic definitions service
	// Note: semantic search handler uses semantic.Service, not catalog.SemanticService
	_ = catalog.NewSemanticService(pool.Pool) // Reserved for future use

	// Initialize unified AI service for orchestration
	unifiedAIService := ai.NewUnifiedAIService(neurondbClient, mcpClient, agentClient)
	unifiedAIHandler := handlers.NewUnifiedAIHandler(unifiedAIService)

	// Initialize unified RAG service
	unifiedRAGService := rag.NewUnifiedRAGService(neurondbClient, mcpClient, agentClient)
	unifiedRAGHandler := handlers.NewUnifiedRAGHandler(unifiedRAGService)

	// Initialize ingestion service
	ingestionService := ingestion.NewService(pool.Pool, mcpClient)
	// Register connector factories to avoid import cycles
	ingestionService.RegisterConnectorFactory("zendesk", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewZendeskConnector()
	})
	ingestionService.RegisterConnectorFactory("salesforce", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewSalesforceConnector()
	})
	ingestionService.RegisterConnectorFactory("slack", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewSlackConnector()
	})
	ingestionService.RegisterConnectorFactory("teams", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewTeamsConnector()
	})
	ingestionService.RegisterConnectorFactory("mysql", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewMySQLConnector()
	})
	ingestionService.RegisterConnectorFactory("sqlserver", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewSQLServerConnector()
	})
	ingestionService.RegisterConnectorFactory("snowflake", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewSnowflakeConnector()
	})
	ingestionService.RegisterConnectorFactory("bigquery", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewBigQueryConnector()
	})
	ingestionService.RegisterConnectorFactory("redshift", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewRedshiftConnector()
	})
	ingestionService.RegisterConnectorFactory("mongodb", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewMongoDBConnector()
	})
	ingestionService.RegisterConnectorFactory("oracle", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewOracleConnector()
	})
	ingestionService.RegisterConnectorFactory("databricks", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewDatabricksConnector()
	})
	ingestionService.RegisterConnectorFactory("elasticsearch", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewElasticsearchConnector()
	})
	ingestionService.RegisterConnectorFactory("s3", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewS3Connector()
	})
	// Phase 3.1: Additional connectors
	ingestionService.RegisterConnectorFactory("azuresql", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewAzureSQLConnector()
	})
	ingestionService.RegisterConnectorFactory("azuresynapse", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewAzureSynapseConnector()
	})
	ingestionService.RegisterConnectorFactory("teradata", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewTeradataConnector()
	})
	ingestionService.RegisterConnectorFactory("presto", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewPrestoConnector()
	})
	ingestionService.RegisterConnectorFactory("trino", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewTrinoConnector()
	})
	ingestionService.RegisterConnectorFactory("hive", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewHiveConnector()
	})
	ingestionService.RegisterConnectorFactory("cassandra", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewCassandraConnector()
	})
	ingestionService.RegisterConnectorFactory("dynamodb", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewDynamoDBConnector()
	})
	ingestionService.RegisterConnectorFactory("redis", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewRedisConnector()
	})
	// Phase 3.1: Additional BI and ETL connectors
	ingestionService.RegisterConnectorFactory("kafka", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewKafkaConnector()
	})
	ingestionService.RegisterConnectorFactory("splunk", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewSplunkConnector()
	})
	ingestionService.RegisterConnectorFactory("tableau", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewTableauConnector()
	})
	ingestionService.RegisterConnectorFactory("powerbi", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewPowerBIConnector()
	})
	ingestionService.RegisterConnectorFactory("looker", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewLookerConnector()
	})
	ingestionService.RegisterConnectorFactory("dbt", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewDbtConnector()
	})
	ingestionService.RegisterConnectorFactory("airflow", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewAirflowConnector()
	})
	ingestionService.RegisterConnectorFactory("fivetran", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewFivetranConnector()
	})
	ingestionService.RegisterConnectorFactory("stitch", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewStitchConnector()
	})
	ingestionService.RegisterConnectorFactory("segment", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewSegmentConnector()
	})
	ingestionService.RegisterConnectorFactory("hubspot", func(ct string) ingestion.Connector {
		return ingestionconnectors.NewHubSpotConnector()
	})
	ingestionHandler := handlers.NewIngestionHandler(ingestionService)

	// Initialize audit service
	auditService := audit.NewAuditService(pool.Pool)
	auditHandler := handlers.NewAuditHandler(auditService)

	// Initialize session manager
	sessionManager := session.NewManager(
		pool.Pool,
		cfg.Auth.Session.AccessTokenTTL,
		cfg.Auth.Session.RefreshTokenTTL,
		cfg.Auth.Session.CookieDomain,
		cfg.Auth.Session.CookieSecure,
		cfg.Auth.Session.CookieSameSite,
	)

	// Start session cleanup service (runs every hour)
	var cleanupService *session.CleanupService
	cleanupService = session.NewCleanupService(pool.Pool, 1*time.Hour)
	cleanupService.Start(ctx)

	// Apply session middleware to API routes (before API key middleware)
	apiRouter.Use(sessionManager.SessionMiddleware())

	// Apply database middleware to inject correct database pool based on session
	apiRouter.Use(middleware.DatabaseMiddleware(multiPool))

	// Initialize enhanced auth services
	authService := auth.NewAuthService(queries, cfg.Auth.JWTSecret, sessionManager)
	oidcService := auth.NewOIDCService(queries)
	scimService := auth.NewSCIMService(queries, cfg.Auth.SCIMSecret)
	sessionService := auth.NewSessionService(queries)
	twoFactorService := auth.NewTwoFactorService(queries)
	authEnhancedHandler := handlers.NewAuthEnhancedHandler(
		authService,
		oidcService,
		scimService,
		sessionService,
		twoFactorService,
		sessionManager,
	)

	// Initialize RBAC services
	rbacService := auth.NewRBACService(queries)
	workspaceService := auth.NewWorkspaceService(queries)
	rbacHandler := handlers.NewRBACHandler(rbacService, workspaceService)

	// Initialize API key service
	apiKeyService := auth.NewAPIKeyService(queries)
	apiKeyEnhancedHandler := handlers.NewAPIKeyEnhancedHandler(apiKeyService, queries)

	// Initialize tenancy service (if needed)
	// tenancyMode := tenancy.TenancyMode(cfg.Database.TenancyMode)
	// if tenancyMode == "" {
	// 	tenancyMode = tenancy.TenancyModeSchema
	// }
	// tenancyService := tenancy.NewTenancyService(pool.Pool, tenancyMode)

	// Initialize policy engine
	policyEngine := policy.NewPolicyEngine(pool.Pool)
	policyHandler := handlers.NewPolicyHandler(policyEngine)

	// Initialize webhook handler (webhookService already initialized above)
	webhookHandler := handlers.NewWebhookHandler(webhookService)

	// Semantic search routes
	apiRouter.HandleFunc("/semantic/search", semanticHandler.Search).Methods("POST")
	apiRouter.HandleFunc("/semantic/rag", semanticHandler.RAG).Methods("POST")
	apiRouter.HandleFunc("/semantic/documents", semanticHandler.CreateDocument).Methods("POST")
	apiRouter.HandleFunc("/semantic/documents/{id}", semanticHandler.UpdateDocument).Methods("PUT")
	apiRouter.HandleFunc("/semantic/collections/{id}", semanticHandler.GetCollection).Methods("GET")

	// Unified AI routes
	apiRouter.HandleFunc("/ai/embedding", unifiedAIHandler.GenerateEmbedding).Methods("POST")
	apiRouter.HandleFunc("/ai/workflow", unifiedAIHandler.ExecuteWorkflow).Methods("POST")
	apiRouter.HandleFunc("/ai/register-tools", unifiedAIHandler.RegisterTools).Methods("POST")

	// Unified RAG routes
	apiRouter.HandleFunc("/rag/query", unifiedRAGHandler.PerformRAG).Methods("POST")
	apiRouter.HandleFunc("/rag/query/stream", unifiedRAGHandler.PerformRAGStream).Methods("POST")
	apiRouter.HandleFunc("/rag/status", unifiedRAGHandler.GetRAGStatus).Methods("GET")

	// Pipeline routes
	apiRouter.HandleFunc("/semantic/pipelines", pipelineHandler.CreatePipeline).Methods("POST")
	apiRouter.HandleFunc("/semantic/pipelines/{id}", pipelineHandler.GetPipeline).Methods("GET")
	apiRouter.HandleFunc("/semantic/pipelines/{id}/versions", pipelineHandler.ListPipelineVersions).Methods("GET")
	apiRouter.HandleFunc("/semantic/pipelines/{id}/replay", pipelineHandler.ReplayPipeline).Methods("POST")
	apiRouter.HandleFunc("/semantic/pipelines/{id}/activate", pipelineHandler.ActivatePipeline).Methods("POST")

	// Warehouse routes
	apiRouter.HandleFunc("/warehouse/query", warehouseHandler.Query).Methods("POST")
	apiRouter.HandleFunc("/warehouse/queries/{id}", warehouseHandler.GetQuery).Methods("GET")
	apiRouter.HandleFunc("/warehouse/queries/history", warehouseHandler.GetQueryHistory).Methods("GET")
	apiRouter.HandleFunc("/warehouse/optimize", warehouseHandler.GetQueryOptimization).Methods("POST")
	apiRouter.HandleFunc("/warehouse/schemas", warehouseHandler.ListSchemas).Methods("GET")
	apiRouter.HandleFunc("/warehouse/schemas", warehouseHandler.CreateSchema).Methods("POST")
	apiRouter.HandleFunc("/warehouse/schemas/{id}", warehouseHandler.GetSchema).Methods("GET")

	// Saved searches routes
	apiRouter.HandleFunc("/warehouse/saved-searches", savedSearchHandler.ListSavedSearches).Methods("GET")
	apiRouter.HandleFunc("/warehouse/saved-searches", savedSearchHandler.CreateSavedSearch).Methods("POST")
	apiRouter.HandleFunc("/warehouse/saved-searches/{id}", savedSearchHandler.GetSavedSearch).Methods("GET")
	apiRouter.HandleFunc("/warehouse/saved-searches/{id}", savedSearchHandler.UpdateSavedSearch).Methods("PUT")
	apiRouter.HandleFunc("/warehouse/saved-searches/{id}", savedSearchHandler.DeleteSavedSearch).Methods("DELETE")
	apiRouter.HandleFunc("/warehouse/saved-searches/{id}/execute", savedSearchHandler.ExecuteSavedSearch).Methods("POST")

	// Query governance routes
	apiRouter.HandleFunc("/warehouse/governance/validate", governanceHandler.ValidateQuery).Methods("POST")
	apiRouter.HandleFunc("/warehouse/governance/sanitize", governanceHandler.SanitizeQuery).Methods("POST")

	// Cache routes
	apiRouter.HandleFunc("/warehouse/cache", cacheHandler.GetCachedResult).Methods("GET")
	apiRouter.HandleFunc("/warehouse/cache/invalidate", cacheHandler.InvalidateCache).Methods("POST")
	apiRouter.HandleFunc("/warehouse/cache/stats", cacheHandler.GetCacheStats).Methods("GET")

	// Workflow routes
	apiRouter.HandleFunc("/workflows", workflowHandler.ListWorkflows).Methods("GET")
	apiRouter.HandleFunc("/workflows", workflowHandler.CreateWorkflow).Methods("POST")
	apiRouter.HandleFunc("/workflows/{id}", workflowHandler.GetWorkflow).Methods("GET")
	apiRouter.HandleFunc("/workflows/{id}", workflowHandler.UpdateWorkflow).Methods("PUT")
	apiRouter.HandleFunc("/workflows/{id}", workflowHandler.DeleteWorkflow).Methods("DELETE")
	apiRouter.HandleFunc("/workflows/{id}/execute", workflowHandler.ExecuteWorkflow).Methods("POST")
	apiRouter.HandleFunc("/workflows/{id}/versions", workflowHandler.CreateWorkflowVersion).Methods("POST")
	apiRouter.HandleFunc("/workflows/{id}/versions", workflowHandler.GetWorkflowVersions).Methods("GET")
	apiRouter.HandleFunc("/workflows/{id}/versions/{version_id}", workflowHandler.GetWorkflowVersion).Methods("GET")
	apiRouter.HandleFunc("/workflows/{id}/schedule", workflowHandler.ScheduleWorkflow).Methods("POST")
	apiRouter.HandleFunc("/workflows/{id}/schedules", workflowHandler.GetScheduledWorkflows).Methods("GET")
	apiRouter.HandleFunc("/workflows/{id}/schedules/{schedule_id}/cancel", workflowHandler.CancelScheduledWorkflow).Methods("POST")
	apiRouter.HandleFunc("/workflows/{id}/monitoring", workflowHandler.GetWorkflowMonitoring).Methods("GET")
	apiRouter.HandleFunc("/workflows/executions/{id}/status", workflowHandler.GetWorkflowExecutionStatus).Methods("GET")
	apiRouter.HandleFunc("/workflows/executions/{id}/recover", workflowHandler.RecoverWorkflowExecution).Methods("POST")
	apiRouter.HandleFunc("/workflows/executions/{id}/logs", workflowHandler.GetWorkflowExecutionLogs).Methods("GET")
	apiRouter.HandleFunc("/workflows/executions/{id}/metrics", workflowHandler.GetWorkflowExecutionMetrics).Methods("GET")
	apiRouter.HandleFunc("/workflows/executions/{id}/decisions", workflowHandler.GetWorkflowExecutionDecisions).Methods("GET")

	// Compliance routes
	apiRouter.HandleFunc("/compliance/check", complianceHandler.CheckCompliance).Methods("POST")
	apiRouter.HandleFunc("/compliance/anomalies", complianceHandler.GetAnomalyDetections).Methods("GET")
	apiRouter.HandleFunc("/compliance/policies", complianceHandler.ListPolicies).Methods("GET")
	apiRouter.HandleFunc("/compliance/policies", complianceHandler.CreatePolicy).Methods("POST")
	apiRouter.HandleFunc("/compliance/policies/{id}", complianceHandler.GetPolicy).Methods("GET")
	apiRouter.HandleFunc("/compliance/policies/{id}", complianceHandler.UpdatePolicy).Methods("PUT")
	apiRouter.HandleFunc("/compliance/policies/{id}", complianceHandler.DeletePolicy).Methods("DELETE")
	apiRouter.HandleFunc("/compliance/report", complianceHandler.GetComplianceReport).Methods("GET")

	// Analytics routes
	apiRouter.HandleFunc("/analytics/search", analyticsHandler.GetSearchAnalytics).Methods("GET")
	apiRouter.HandleFunc("/analytics/warehouse", analyticsHandler.GetWarehouseAnalytics).Methods("GET")
	apiRouter.HandleFunc("/analytics/workflows", analyticsHandler.GetWorkflowAnalytics).Methods("GET")
	apiRouter.HandleFunc("/analytics/compliance", analyticsHandler.GetComplianceAnalytics).Methods("GET")
	apiRouter.HandleFunc("/analytics/retrieval-quality", analyticsHandler.GetRetrievalQuality).Methods("GET")

	// Models routes
	apiRouter.HandleFunc("/models", modelsHandler.RegisterModel).Methods("POST")
	apiRouter.HandleFunc("/models/{id}", modelsHandler.GetModel).Methods("GET")
	apiRouter.HandleFunc("/models/{id}/infer", modelsHandler.InferModel).Methods("POST")

	// Integration routes
	apiRouter.HandleFunc("/integrations", integrationHandler.ListIntegrations).Methods("GET")
	apiRouter.HandleFunc("/integrations", integrationHandler.CreateIntegration).Methods("POST")
	apiRouter.HandleFunc("/integrations/{id}", integrationHandler.GetIntegration).Methods("GET")
	apiRouter.HandleFunc("/integrations/{id}", integrationHandler.UpdateIntegration).Methods("PUT")
	apiRouter.HandleFunc("/integrations/{id}", integrationHandler.DeleteIntegration).Methods("DELETE")
	apiRouter.HandleFunc("/integrations/{id}/test", integrationHandler.TestIntegration).Methods("POST")
	apiRouter.HandleFunc("/integrations/health", integrationHandler.GetIntegrationHealth).Methods("GET")
	apiRouter.HandleFunc("/integrations/helpdesk/sync", integrationHandler.SyncHelpdesk).Methods("POST")

	// Alerts routes
	apiRouter.HandleFunc("/alerts/check", alertsHandler.CheckAlerts).Methods("POST")
	apiRouter.HandleFunc("/alerts", alertsHandler.GetAlerts).Methods("GET")
	apiRouter.HandleFunc("/alerts/{id}/resolve", alertsHandler.ResolveAlert).Methods("POST")
	apiRouter.HandleFunc("/alerts/rules", alertsHandler.CreateAlertRule).Methods("POST")
	apiRouter.HandleFunc("/alerts/rules/{id}", alertsHandler.UpdateAlertRule).Methods("PUT")
	apiRouter.HandleFunc("/alerts/rules/{id}", alertsHandler.DeleteAlertRule).Methods("DELETE")

	// Support routes
	apiRouter.HandleFunc("/support/tickets", supportHandler.CreateTicket).Methods("POST")
	apiRouter.HandleFunc("/support/tickets", supportHandler.ListTickets).Methods("GET")
	apiRouter.HandleFunc("/support/tickets/{id}", supportHandler.GetTicket).Methods("GET")
	apiRouter.HandleFunc("/support/tickets/{id}/conversations", supportHandler.AddConversation).Methods("POST")
	apiRouter.HandleFunc("/support/tickets/{id}/conversations", supportHandler.GetConversations).Methods("GET")
	apiRouter.HandleFunc("/support/tickets/{id}/similar-cases", supportHandler.GetSimilarCases).Methods("GET")

	// Knowledge graph routes
	apiRouter.HandleFunc("/knowledge-graph/entities/extract", knowledgeGraphHandler.ExtractEntities).Methods("POST")
	apiRouter.HandleFunc("/knowledge-graph/entities/{id}", knowledgeGraphHandler.GetEntity).Methods("GET")
	apiRouter.HandleFunc("/knowledge-graph/entities/{id}/links", knowledgeGraphHandler.GetEntityLinks).Methods("GET")
	apiRouter.HandleFunc("/knowledge-graph/entities/search", knowledgeGraphHandler.SearchEntities).Methods("POST")
	apiRouter.HandleFunc("/knowledge-graph/entities/link", knowledgeGraphHandler.LinkEntities).Methods("POST")
	apiRouter.HandleFunc("/knowledge-graph/traverse", knowledgeGraphHandler.TraverseGraph).Methods("POST")
	apiRouter.HandleFunc("/knowledge-graph/entity-types", knowledgeGraphHandler.CreateEntityType).Methods("POST")
	apiRouter.HandleFunc("/knowledge-graph/glossary", knowledgeGraphHandler.CreateGlossaryTerm).Methods("POST")
	apiRouter.HandleFunc("/knowledge-graph/glossary/{id}", knowledgeGraphHandler.GetGlossaryTerm).Methods("GET")
	apiRouter.HandleFunc("/knowledge-graph/glossary/search", knowledgeGraphHandler.SearchGlossary).Methods("POST")

	// Data sources routes
	apiRouter.HandleFunc("/data-sources", dataSourceHandler.ListDataSources).Methods("GET")
	apiRouter.HandleFunc("/data-sources", dataSourceHandler.CreateDataSource).Methods("POST")
	apiRouter.HandleFunc("/data-sources/{id}", dataSourceHandler.GetDataSource).Methods("GET")
	apiRouter.HandleFunc("/data-sources/{id}", dataSourceHandler.UpdateDataSource).Methods("PUT")
	apiRouter.HandleFunc("/data-sources/{id}", dataSourceHandler.DeleteDataSource).Methods("DELETE")
	apiRouter.HandleFunc("/data-sources/{id}/sync", dataSourceHandler.TriggerSync).Methods("POST")
	apiRouter.HandleFunc("/data-sources/{id}/status", dataSourceHandler.GetSyncStatus).Methods("GET")

	// Business metrics routes (semantic layer)
	apiRouter.HandleFunc("/metrics", businessMetricsHandler.ListMetrics).Methods("GET")
	apiRouter.HandleFunc("/metrics", businessMetricsHandler.CreateMetric).Methods("POST")
	apiRouter.HandleFunc("/metrics/{id}", businessMetricsHandler.GetMetric).Methods("GET")
	apiRouter.HandleFunc("/metrics/{id}", businessMetricsHandler.UpdateMetric).Methods("PUT")
	apiRouter.HandleFunc("/metrics/{id}", businessMetricsHandler.DeleteMetric).Methods("DELETE")
	apiRouter.HandleFunc("/metrics/search", businessMetricsHandler.SearchMetrics).Methods("POST")
	apiRouter.HandleFunc("/metrics/discover", businessMetricsHandler.DiscoverMetrics).Methods("POST")
	apiRouter.HandleFunc("/metrics/{id}/calculate", businessMetricsHandler.CalculateMetric).Methods("POST")
	apiRouter.HandleFunc("/metrics/{id}/lineage", businessMetricsHandler.GetMetricLineage).Methods("GET")
	apiRouter.HandleFunc("/metrics/{id}/lineage", businessMetricsHandler.AddMetricLineage).Methods("POST")
	apiRouter.HandleFunc("/metrics/{id}/approve", businessMetricsHandler.ApproveMetric).Methods("POST")

	// Agents routes
	apiRouter.HandleFunc("/agents", agentsHandler.ListAgents).Methods("GET")
	apiRouter.HandleFunc("/agents", agentsHandler.CreateAgent).Methods("POST")
	apiRouter.HandleFunc("/agents/{id}", agentsHandler.GetAgent).Methods("GET")
	apiRouter.HandleFunc("/agents/{id}", agentsHandler.UpdateAgent).Methods("PUT")
	apiRouter.HandleFunc("/agents/{id}", agentsHandler.DeleteAgent).Methods("DELETE")
	apiRouter.HandleFunc("/agents/{id}/performance", agentsHandler.GetPerformance).Methods("GET")
	apiRouter.HandleFunc("/agents/{id}/deploy", agentsHandler.DeployAgent).Methods("POST")

	// Observability routes
	apiRouter.HandleFunc("/observability/queries/performance", observabilityHandler.GetQueryPerformance).Methods("GET")
	apiRouter.HandleFunc("/observability/logs", observabilityHandler.GetSystemLogs).Methods("GET")
	apiRouter.HandleFunc("/observability/logs/stream", observabilityHandler.GetLogStream).Methods("GET")
	apiRouter.HandleFunc("/observability/metrics", observabilityHandler.GetSystemMetrics).Methods("GET")
	apiRouter.HandleFunc("/observability/realtime", observabilityHandler.GetRealTimeMetrics).Methods("GET")
	apiRouter.HandleFunc("/observability/benchmark", observabilityHandler.GetPerformanceBenchmark).Methods("GET")
	apiRouter.HandleFunc("/observability/cost/breakdown", observabilityHandler.GetCostBreakdown).Methods("GET")
	apiRouter.HandleFunc("/observability/agent-logs", observabilityHandler.GetAgentLogs).Methods("GET")
	apiRouter.HandleFunc("/observability/workflow-logs", observabilityHandler.GetWorkflowLogs).Methods("GET")

	// Enhanced metrics routes
	apiRouter.HandleFunc("/observability/metrics/latency", metricsEnhancedHandler.GetLatencyMetrics).Methods("GET")
	apiRouter.HandleFunc("/observability/metrics/error-rate", metricsEnhancedHandler.GetErrorRate).Methods("GET")
	apiRouter.HandleFunc("/observability/metrics/token-usage", metricsEnhancedHandler.GetTokenUsage).Methods("GET")
	apiRouter.HandleFunc("/observability/metrics/embedding-cost", metricsEnhancedHandler.GetEmbeddingCost).Methods("GET")

	// Lineage routes
	apiRouter.HandleFunc("/lineage/{resource_type}/{resource_id}", lineageHandler.GetLineage).Methods("GET")
	apiRouter.HandleFunc("/lineage/track", lineageHandler.TrackTransformation).Methods("POST")
	apiRouter.HandleFunc("/lineage/impact/{resource_id}", lineageHandler.GetImpactAnalysis).Methods("GET")
	apiRouter.HandleFunc("/lineage/graph", lineageHandler.GetFullGraph).Methods("GET")

	// Audit routes
	apiRouter.HandleFunc("/audit/events", auditHandler.GetAuditEvents).Methods("GET")
	apiRouter.HandleFunc("/audit/activity", auditHandler.GetActivityTimeline).Methods("GET")
	apiRouter.HandleFunc("/audit/compliance-trail", auditHandler.GetComplianceTrail).Methods("GET")
	apiRouter.HandleFunc("/audit/search", auditHandler.SearchAuditEvents).Methods("POST")

	// Billing routes
	apiRouter.HandleFunc("/billing/usage", billingHandler.GetUsageMetrics).Methods("GET")
	apiRouter.HandleFunc("/billing/metrics", billingHandler.GetDetailedMetrics).Methods("GET")
	apiRouter.HandleFunc("/billing/dashboard", billingHandler.GetDashboardData).Methods("GET")
	apiRouter.HandleFunc("/billing/track", billingHandler.TrackUsage).Methods("POST")

	// Versioning routes
	apiRouter.HandleFunc("/versions/{resource_type}/{resource_id}", versioningHandler.ListVersions).Methods("GET")
	apiRouter.HandleFunc("/versions/create", versioningHandler.CreateVersion).Methods("POST")
	apiRouter.HandleFunc("/versions/{id}", versioningHandler.GetVersion).Methods("GET")
	apiRouter.HandleFunc("/versions/{id}/rollback", versioningHandler.RollbackVersion).Methods("POST")
	apiRouter.HandleFunc("/versions/{id}/history", versioningHandler.GetVersionHistory).Methods("GET")

	// Catalog routes
	apiRouter.HandleFunc("/catalog/datasets", catalogHandler.ListDatasets).Methods("GET")
	apiRouter.HandleFunc("/catalog/datasets/{id}", catalogHandler.GetDataset).Methods("GET")
	apiRouter.HandleFunc("/catalog/search", catalogHandler.SearchDatasets).Methods("GET")
	apiRouter.HandleFunc("/catalog/owners", catalogHandler.ListOwners).Methods("GET")
	apiRouter.HandleFunc("/catalog/discover", catalogHandler.DiscoverDatasets).Methods("POST")

	// Metric catalog routes (using business metrics handler which has both services)
	apiRouter.HandleFunc("/catalog/metrics", businessMetricsHandler.ListMetrics).Methods("GET")
	apiRouter.HandleFunc("/catalog/metrics", businessMetricsHandler.CreateMetric).Methods("POST")
	apiRouter.HandleFunc("/catalog/metrics/{id}", businessMetricsHandler.GetMetric).Methods("GET")
	apiRouter.HandleFunc("/catalog/metrics/{id}/lineage", businessMetricsHandler.GetMetricLineage).Methods("GET")

	// Ingestion routes
	apiRouter.HandleFunc("/ingestion/jobs", ingestionHandler.CreateJob).Methods("POST")
	apiRouter.HandleFunc("/ingestion/jobs", ingestionHandler.ListJobs).Methods("GET")
	apiRouter.HandleFunc("/ingestion/jobs/{id}", ingestionHandler.GetJob).Methods("GET")
	apiRouter.HandleFunc("/ingestion/jobs/{id}/execute", ingestionHandler.ExecuteJob).Methods("POST")

	// Policy routes
	apiRouter.HandleFunc("/policies", policyHandler.CreatePolicy).Methods("POST")
	apiRouter.HandleFunc("/policies/{id}", policyHandler.GetPolicy).Methods("GET")
	apiRouter.HandleFunc("/policies/{id}/evaluate", policyHandler.EvaluatePolicy).Methods("POST")

	// Webhook routes
	apiRouter.HandleFunc("/webhooks", webhookHandler.CreateWebhook).Methods("POST")
	apiRouter.HandleFunc("/webhooks", webhookHandler.ListWebhooks).Methods("GET")
	apiRouter.HandleFunc("/webhooks/{id}", webhookHandler.GetWebhook).Methods("GET")
	apiRouter.HandleFunc("/webhooks/{id}/trigger", webhookHandler.TriggerWebhook).Methods("POST")

	// Enhanced authentication routes
	// Auth routes (login, register, me, logout, refresh)
	apiRouter.HandleFunc("/auth/login", authEnhancedHandler.Login).Methods("POST")
	apiRouter.HandleFunc("/auth/register", authEnhancedHandler.Register).Methods("POST")
	apiRouter.HandleFunc("/auth/me", authEnhancedHandler.GetCurrentUser).Methods("GET")
	apiRouter.HandleFunc("/auth/logout", authEnhancedHandler.Logout).Methods("POST")
	apiRouter.HandleFunc("/auth/refresh", authEnhancedHandler.RefreshToken).Methods("POST")

	// OIDC routes
	apiRouter.HandleFunc("/auth/oidc/{provider}/initiate", authEnhancedHandler.InitiateOIDC).Methods("POST")
	apiRouter.HandleFunc("/auth/oidc/{provider}/callback", authEnhancedHandler.HandleOIDCCallback).Methods("GET")
	apiRouter.HandleFunc("/auth/scim/{path:.*}", authEnhancedHandler.HandleSCIM).Methods("GET", "POST", "PUT", "DELETE")
	apiRouter.HandleFunc("/auth/2fa/generate", authEnhancedHandler.GenerateTOTPSecret).Methods("POST")
	apiRouter.HandleFunc("/auth/sessions", authEnhancedHandler.GetUserSessions).Methods("GET")
	apiRouter.HandleFunc("/auth/sessions/{id}", authEnhancedHandler.RevokeSession).Methods("DELETE")

	// RBAC routes
	apiRouter.HandleFunc("/rbac/workspaces", rbacHandler.CreateWorkspace).Methods("POST")
	apiRouter.HandleFunc("/rbac/workspaces", rbacHandler.ListWorkspaces).Methods("GET")
	apiRouter.HandleFunc("/rbac/permissions/check", rbacHandler.CheckPermission).Methods("POST")

	// Enhanced API key routes
	apiRouter.HandleFunc("/api-keys", apiKeyEnhancedHandler.CreateAPIKey).Methods("POST")
	apiRouter.HandleFunc("/api-keys/{id}/rotate", apiKeyEnhancedHandler.RotateAPIKey).Methods("POST")
	apiRouter.HandleFunc("/api-keys/{id}/usage", apiKeyEnhancedHandler.GetUsageAnalytics).Methods("GET")
	apiRouter.HandleFunc("/api-keys/{id}/revoke", apiKeyEnhancedHandler.RevokeAPIKey).Methods("POST")

	// Initialize SSO service
	ssoConfig := &auth.SSOConfig{
		BaseURL:           cfg.Server.Host + ":" + cfg.Server.Port,
		CallbackPath:      "/api/v1/sso/callback",
		SessionTimeout:    24 * time.Hour,
		EnableAutoMapping: true,
	}
	ssoService := auth.NewSSOService(pool.Pool, ssoConfig)
	ssoHandler := handlers.NewSSOHandler(ssoService)

	// Initialize comments service
	commentsService := comments.NewService(pool.Pool)
	commentsHandler := handlers.NewCommentsHandler(commentsService)

	// Initialize ownership service
	ownershipService := ownership.NewService(pool.Pool)
	ownershipHandler := handlers.NewOwnershipHandler(ownershipService)

	// Initialize webhooks service (new implementation)
	webhooksService := webhooks.NewService(pool.Pool)
	webhooksHandler := handlers.NewWebhooksHandler(webhooksService)

	// Initialize connector framework service
	connectorService := connectors.NewConnectorService(pool.Pool)
	connectorHandler := handlers.NewConnectorHandler(connectorService)

	// Initialize data quality service
	// Initialize data quality service with MCP and Agent clients for ML-powered analysis
	dataQualityService := dataquality.NewServiceWithMCPAndAgent(pool.Pool, neurondbClient, mcpClient, agentClient)
	dataQualityHandler := handlers.NewDataQualityHandler(dataQualityService)

	// Initialize profiling service
	profilingService := profiling.NewService(pool.Pool, neurondbClient, mcpClient)
	profilingHandler := handlers.NewProfilingHandler(profilingService)

	// Initialize classification service
	classificationService := classification.NewService(pool.Pool)
	classificationHandler := handlers.NewClassificationHandler(classificationService)

	// Initialize column lineage service
	columnLineageService := lineage.NewColumnLineageService(pool.Pool)
	columnLineageHandler := handlers.NewColumnLineageHandler(columnLineageService)

	// Initialize region service (Phase 2.1)
	regionService := tenancy.NewRegionService(pool.Pool)
	regionHandler := handlers.NewRegionHandler(regionService)

	// Initialize backup service (Phase 2.2)
	backupConfig := backup.BackupConfig{
		BackupDir:     "/var/backups/neuronip",
		RetentionDays: 30,
		Compress:      true,
	}
	backupService := backup.NewBackupService(pool.Pool, backupConfig)
	backupHandler := handlers.NewBackupHandler(backupService)

	// Initialize column and row security services (Phase 2.3 - already implemented)
	// These services are available for integration with query handlers
	_ = auth.NewColumnSecurityService(queries)
	_ = auth.NewRowSecurityService(queries)

	// Initialize DSAR service (Phase 2.4)
	dsarService := compliance.NewDSARService(pool.Pool)
	dsarHandler := handlers.NewDSARHandler(dsarService)

	// Initialize PIA service (Phase 2.4)
	piaService := compliance.NewPIAService(pool.Pool)
	piaHandler := handlers.NewPIAHandler(piaService)

	// Initialize consent service (Phase 2.4)
	consentService := compliance.NewConsentService(pool.Pool)
	consentHandler := handlers.NewConsentHandler(consentService)

	// Initialize masking service (Phase 2.5)
	maskingService := masking.NewMaskingService(pool.Pool)
	maskingHandler := handlers.NewMaskingHandler(maskingService)

	// SSO routes
	apiRouter.HandleFunc("/sso/providers", ssoHandler.CreateProvider).Methods("POST")
	apiRouter.HandleFunc("/sso/providers", ssoHandler.ListProviders).Methods("GET")
	apiRouter.HandleFunc("/sso/providers/{id}", ssoHandler.GetProvider).Methods("GET")
	apiRouter.HandleFunc("/sso/providers/{id}/initiate", ssoHandler.InitiateSSO).Methods("GET")
	apiRouter.HandleFunc("/sso/callback", ssoHandler.SSOCallback).Methods("GET", "POST")
	apiRouter.HandleFunc("/sso/validate", ssoHandler.ValidateSession).Methods("POST")

	// Comments routes
	apiRouter.HandleFunc("/comments", commentsHandler.CreateComment).Methods("POST")
	apiRouter.HandleFunc("/comments/{id}", commentsHandler.GetComment).Methods("GET")
	apiRouter.HandleFunc("/comments/{resource_type}/{resource_id}", commentsHandler.ListComments).Methods("GET")
	apiRouter.HandleFunc("/comments/{id}/resolve", commentsHandler.ResolveComment).Methods("POST")
	apiRouter.HandleFunc("/comments/{id}", commentsHandler.DeleteComment).Methods("DELETE")

	// Ownership routes
	apiRouter.HandleFunc("/ownership", ownershipHandler.AssignOwnership).Methods("POST")
	apiRouter.HandleFunc("/ownership/{resource_type}/{resource_id}", ownershipHandler.GetOwnership).Methods("GET")
	apiRouter.HandleFunc("/ownership/by-owner", ownershipHandler.ListOwnershipByOwner).Methods("GET")
	apiRouter.HandleFunc("/ownership/{resource_type}/{resource_id}", ownershipHandler.RemoveOwnership).Methods("DELETE")

	// Webhooks routes (new implementation)
	apiRouter.HandleFunc("/webhooks", webhooksHandler.CreateWebhook).Methods("POST")
	apiRouter.HandleFunc("/webhooks", webhooksHandler.ListWebhooks).Methods("GET")
	apiRouter.HandleFunc("/webhooks/{id}", webhooksHandler.GetWebhook).Methods("GET")
	apiRouter.HandleFunc("/webhooks/{id}/trigger", webhooksHandler.TriggerWebhook).Methods("POST")

	// Connector framework routes
	apiRouter.HandleFunc("/connectors", connectorHandler.CreateConnector).Methods("POST")
	apiRouter.HandleFunc("/connectors", connectorHandler.ListConnectors).Methods("GET")
	apiRouter.HandleFunc("/connectors/{id}", connectorHandler.GetConnector).Methods("GET")
	apiRouter.HandleFunc("/connectors/{id}/sync", connectorHandler.SyncConnector).Methods("POST")

	// Data quality routes
	apiRouter.HandleFunc("/data-quality/rules", dataQualityHandler.CreateRule).Methods("POST")
	apiRouter.HandleFunc("/data-quality/rules/{id}", dataQualityHandler.GetRule).Methods("GET")
	apiRouter.HandleFunc("/data-quality/rules/{id}/execute", dataQualityHandler.ExecuteRule).Methods("POST")
	apiRouter.HandleFunc("/data-quality/dashboard", dataQualityHandler.GetDashboard).Methods("GET")
	apiRouter.HandleFunc("/data-quality/trends", dataQualityHandler.GetTrends).Methods("GET")

	// Data profiling routes
	apiRouter.HandleFunc("/profiling/connectors/{connector_id}/schemas/{schema_name}/tables/{table_name}", profilingHandler.ProfileTable).Methods("POST")
	apiRouter.HandleFunc("/profiling/connectors/{connector_id}/schemas/{schema_name}/tables/{table_name}/columns/{column_name}", profilingHandler.ProfileColumn).Methods("POST")

	// Classification routes
	apiRouter.HandleFunc("/classification/connectors/{connector_id}/schemas/{schema_name}/tables/{table_name}/columns/{column_name}", classificationHandler.ClassifyColumn).Methods("POST")
	apiRouter.HandleFunc("/classification/connectors/{id}/classify", classificationHandler.ClassifyConnector).Methods("POST")
	apiRouter.HandleFunc("/classification/rules", classificationHandler.CreateClassificationRule).Methods("POST")

	// Column lineage routes
	apiRouter.HandleFunc("/lineage/columns/{connector_id}/{schema_name}/{table_name}/{column_name}", columnLineageHandler.GetColumnLineage).Methods("GET")
	apiRouter.HandleFunc("/lineage/columns/track", columnLineageHandler.TrackColumnLineage).Methods("POST")
	apiRouter.HandleFunc("/lineage/columns/nodes", columnLineageHandler.CreateColumnNode).Methods("POST")

	// Region routes (Phase 2.1)
	apiRouter.HandleFunc("/regions", regionHandler.CreateRegion).Methods("POST")
	apiRouter.HandleFunc("/regions", regionHandler.ListRegions).Methods("GET")
	apiRouter.HandleFunc("/regions/{id}", regionHandler.GetRegion).Methods("GET")
	apiRouter.HandleFunc("/regions/{id}/health", regionHandler.CheckRegionHealth).Methods("GET")
	apiRouter.HandleFunc("/regions/{id}/failover", regionHandler.FailoverToRegion).Methods("POST")

	// Backup routes (Phase 2.2)
	apiRouter.HandleFunc("/backups", backupHandler.CreateBackup).Methods("POST")
	apiRouter.HandleFunc("/backups", backupHandler.ListBackups).Methods("GET")
	apiRouter.HandleFunc("/backups/{id}/restore", backupHandler.RestoreBackup).Methods("POST")

	// DSAR routes (Phase 2.4)
	apiRouter.HandleFunc("/dsar/requests", dsarHandler.CreateDSARRequest).Methods("POST")
	apiRouter.HandleFunc("/dsar/requests", dsarHandler.ListDSARRequests).Methods("GET")
	apiRouter.HandleFunc("/dsar/requests/{id}", dsarHandler.GetDSARRequest).Methods("GET")
	apiRouter.HandleFunc("/dsar/requests/{id}/complete", dsarHandler.CompleteDSARRequest).Methods("POST")

	// PIA routes (Phase 2.4)
	apiRouter.HandleFunc("/pia/requests", piaHandler.CreatePIARequest).Methods("POST")
	apiRouter.HandleFunc("/pia/requests/{id}", piaHandler.GetPIARequest).Methods("GET")
	apiRouter.HandleFunc("/pia/requests/{id}/submit", piaHandler.SubmitPIARequest).Methods("POST")
	apiRouter.HandleFunc("/pia/requests/{id}/review", piaHandler.ReviewPIARequest).Methods("POST")

	// Consent routes (Phase 2.4)
	apiRouter.HandleFunc("/consent", consentHandler.RecordConsent).Methods("POST")
	apiRouter.HandleFunc("/consent/withdraw", consentHandler.WithdrawConsent).Methods("POST")
	apiRouter.HandleFunc("/consent/{subject_id}", consentHandler.CheckConsent).Methods("GET")
	apiRouter.HandleFunc("/consent/subject/{subject_id}", consentHandler.GetSubjectConsents).Methods("GET")

	// Masking routes (Phase 2.5)
	apiRouter.HandleFunc("/masking/policies", maskingHandler.CreateMaskingPolicy).Methods("POST")
	apiRouter.HandleFunc("/masking/policies", maskingHandler.GetMaskingPolicy).Methods("GET")
	apiRouter.HandleFunc("/masking/apply", maskingHandler.ApplyMasking).Methods("POST")

	// Enterprise Feature Routes - Semantic Layer Approval
	apiRouter.HandleFunc("/metrics/approvals/queue", semanticHandler.GetApprovalQueue).Methods("GET")
	apiRouter.HandleFunc("/metrics/{id}/approvals", semanticHandler.GetMetricApprovals).Methods("GET")
	apiRouter.HandleFunc("/metrics/{id}/approvals", semanticHandler.CreateMetricApproval).Methods("POST")
	apiRouter.HandleFunc("/metrics/approvals/{id}/approve", semanticHandler.ApproveMetric).Methods("POST")
	apiRouter.HandleFunc("/metrics/approvals/{id}/reject", semanticHandler.RejectMetric).Methods("POST")
	apiRouter.HandleFunc("/metrics/approvals/{id}/request-changes", semanticHandler.RequestChanges).Methods("POST")
	apiRouter.HandleFunc("/metrics/{id}/owners", semanticHandler.GetMetricOwners).Methods("GET")
	apiRouter.HandleFunc("/metrics/{id}/owners", semanticHandler.AddMetricOwner).Methods("POST")
	apiRouter.HandleFunc("/metrics/{id}/owners/{owner_id}", semanticHandler.RemoveMetricOwner).Methods("DELETE")

	// Enterprise Feature Routes - Ingestion Status
	apiRouter.HandleFunc("/ingestion/data-sources/{id}/status", ingestionHandler.GetIngestionStatus).Methods("GET")
	apiRouter.HandleFunc("/ingestion/failures", ingestionHandler.GetIngestionFailures).Methods("GET")
	apiRouter.HandleFunc("/ingestion/jobs/{id}/retry", ingestionHandler.RetryIngestionJob).Methods("POST")

	// Enterprise Feature Routes - Model & Prompt Governance
	apiRouter.HandleFunc("/models", modelGovernanceHandler.ListModels).Methods("GET")
	apiRouter.HandleFunc("/models/{id}", modelGovernanceHandler.GetModel).Methods("GET")
	apiRouter.HandleFunc("/models/{id}/versions", modelGovernanceHandler.GetModelVersions).Methods("GET")
	apiRouter.HandleFunc("/models/{id}/approve", modelGovernanceHandler.ApproveModel).Methods("POST")
	apiRouter.HandleFunc("/models/{name}/rollback", modelGovernanceHandler.RollbackModel).Methods("POST")
	apiRouter.HandleFunc("/prompts", modelGovernanceHandler.ListPrompts).Methods("GET")
	apiRouter.HandleFunc("/prompts/{id}", modelGovernanceHandler.GetPrompt).Methods("GET")
	apiRouter.HandleFunc("/prompts/{name}/versions", modelGovernanceHandler.GetPromptVersions).Methods("GET")
	apiRouter.HandleFunc("/prompts/{id}/approve", modelGovernanceHandler.ApprovePrompt).Methods("POST")
	apiRouter.HandleFunc("/prompts/{name}/rollback", modelGovernanceHandler.RollbackPrompt).Methods("POST")

	// Enterprise Feature Routes - AI Observability
	apiRouter.HandleFunc("/observability/agents/{agent_id}/logs", observabilityHandler.GetAgentExecutionLogs).Methods("GET")
	apiRouter.HandleFunc("/observability/retrieval/metrics", observabilityHandler.GetRetrievalMetrics).Methods("GET")
	apiRouter.HandleFunc("/observability/retrieval/stats", observabilityHandler.GetRetrievalStats).Methods("GET")
	apiRouter.HandleFunc("/observability/hallucination/signals", observabilityHandler.GetHallucinationSignals).Methods("GET")
	apiRouter.HandleFunc("/observability/hallucination/stats", observabilityHandler.GetHallucinationStats).Methods("GET")
	apiRouter.HandleFunc("/observability/queries/{id}/cost", observabilityHandler.GetQueryCost).Methods("GET")
	apiRouter.HandleFunc("/observability/agents/runs/{id}/cost", observabilityHandler.GetAgentRunCost).Methods("GET")

	// Enterprise Feature Routes - Knowledge Graph Query
	apiRouter.HandleFunc("/knowledge-graph/query", knowledgeGraphHandler.ExecuteGraphQuery).Methods("POST")

	// Enterprise Feature Routes - Collaboration
	apiRouter.HandleFunc("/collaboration/dashboards", collaborationHandler.CreateSharedDashboard).Methods("POST")
	apiRouter.HandleFunc("/collaboration/dashboards", collaborationHandler.GetSharedDashboards).Methods("GET")
	apiRouter.HandleFunc("/collaboration/dashboards/{id}/comments", collaborationHandler.AddDashboardComment).Methods("POST")
	apiRouter.HandleFunc("/collaboration/dashboards/{id}/comments", collaborationHandler.GetDashboardComments).Methods("GET")
	apiRouter.HandleFunc("/collaboration/answer-cards", collaborationHandler.CreateAnswerCard).Methods("POST")
	apiRouter.HandleFunc("/collaboration/saved-questions", collaborationHandler.SaveQuestion).Methods("POST")

	// Enterprise Feature Routes - Governance (RLS)
	rlsHandler := handlers.NewRLSHandler(queries)
	apiRouter.HandleFunc("/governance/rls/policies", rlsHandler.GetRLSPolicies).Methods("GET")
	apiRouter.HandleFunc("/governance/rls/policies", rlsHandler.CreateRLSPolicy).Methods("POST")
	
	// Enterprise Feature Routes - Resource Quotas
	quotaHandler := handlers.NewQuotaHandler(pool.Pool)
	apiRouter.HandleFunc("/quotas", quotaHandler.SetQuota).Methods("POST")
	apiRouter.HandleFunc("/quotas", quotaHandler.ListQuotas).Methods("GET")
	apiRouter.HandleFunc("/quotas/check", quotaHandler.CheckQuota).Methods("POST")

	// Enterprise Feature Routes - Integrations (Slack/Teams)
	apiRouter.HandleFunc("/integrations/slack/command", slackBotService.HandleHTTPRequest).Methods("POST")
	apiRouter.HandleFunc("/integrations/teams/message", teamsBotService.HandleHTTPRequest).Methods("POST")
	
	// BI Export Handler
	biExportHandler := handlers.NewBIExportHandler(biExportService)
	apiRouter.HandleFunc("/integrations/bi/export", biExportHandler.ExportQuery).Methods("GET")

	// Tenancy routes (if needed)
	// apiRouter.HandleFunc("/tenants", ...).Methods("POST")

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting",
			"host", cfg.Server.Host,
			"port", cfg.Server.Port,
			"read_timeout", cfg.Server.ReadTimeout,
			"write_timeout", cfg.Server.WriteTimeout,
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received, shutting down server...")

	// Stop cleanup service
	if cleanupService != nil {
		cleanupService.Stop()
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}
