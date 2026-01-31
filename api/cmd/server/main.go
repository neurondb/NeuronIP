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
	"github.com/neurondb/NeuronIP/api/internal/alerts"
	"github.com/neurondb/NeuronIP/api/internal/analytics"
	"github.com/neurondb/NeuronIP/api/internal/audit"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/backup"
	"github.com/neurondb/NeuronIP/api/internal/billing"
	"github.com/neurondb/NeuronIP/api/internal/blocks"
	"github.com/neurondb/NeuronIP/api/internal/cache"
	"github.com/neurondb/NeuronIP/api/internal/catalog"
	"github.com/neurondb/NeuronIP/api/internal/classification"

	// collaboration is used internally by handlers.NewCollaborationHandler
	_ "github.com/neurondb/NeuronIP/api/internal/collaboration"
	"github.com/neurondb/NeuronIP/api/internal/comments"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/config"
	"github.com/neurondb/NeuronIP/api/internal/connectors"
	"github.com/neurondb/NeuronIP/api/internal/databases"
	"github.com/neurondb/NeuronIP/api/internal/dataquality"
	"github.com/neurondb/NeuronIP/api/internal/datasources"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/execution"
	"github.com/neurondb/NeuronIP/api/internal/governance"
	"github.com/neurondb/NeuronIP/api/internal/handlers"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
	ingestionconnectors "github.com/neurondb/NeuronIP/api/internal/ingestion/connectors"
	"github.com/neurondb/NeuronIP/api/internal/integrations"
	bibot "github.com/neurondb/NeuronIP/api/internal/integrations/bi"
	slackbot "github.com/neurondb/NeuronIP/api/internal/integrations/slack"
	teamsbot "github.com/neurondb/NeuronIP/api/internal/integrations/teams"
	"github.com/neurondb/NeuronIP/api/internal/itsm"
	"github.com/neurondb/NeuronIP/api/internal/knowledgegraph"
	"github.com/neurondb/NeuronIP/api/internal/lineage"
	"github.com/neurondb/NeuronIP/api/internal/logging"
	"github.com/neurondb/NeuronIP/api/internal/masking"
	"github.com/neurondb/NeuronIP/api/internal/mcp"
	"github.com/neurondb/NeuronIP/api/internal/metrics"
	"github.com/neurondb/NeuronIP/api/internal/middleware"
	"github.com/neurondb/NeuronIP/api/internal/ml"
	"github.com/neurondb/NeuronIP/api/internal/models"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
	"github.com/neurondb/NeuronIP/api/internal/notion"
	"github.com/neurondb/NeuronIP/api/internal/ownership"
	"github.com/neurondb/NeuronIP/api/internal/policy"
	"github.com/neurondb/NeuronIP/api/internal/profiling"
	"github.com/neurondb/NeuronIP/api/internal/rag"
	"github.com/neurondb/NeuronIP/api/internal/semantic"
	"github.com/neurondb/NeuronIP/api/internal/session"
	"github.com/neurondb/NeuronIP/api/internal/support"
	"github.com/neurondb/NeuronIP/api/internal/tenancy"
	"github.com/neurondb/NeuronIP/api/internal/tracing"
	"github.com/neurondb/NeuronIP/api/internal/versioning"
	"github.com/neurondb/NeuronIP/api/internal/warehouse"
	"github.com/neurondb/NeuronIP/api/internal/webhooks"
	"github.com/neurondb/NeuronIP/api/internal/workflows"
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

	// Pool for NeuronDB: use dedicated neurondb pool if configured, else same as neuronip
	neurondbPool, err := multiPool.GetPool("neurondb")
	if err != nil {
		neurondbPool = pool
	}
	dbNames := []string{"neuronip", "neuronai-demo"}
	if neurondbPool != pool {
		dbNames = append(dbNames, "neurondb")
	}

	// Create Pool wrapper for backward compatibility (for future route/handler wiring)
	_ = &db.Pool{Pool: pool}

	logger.Info("Database pools created successfully",
		"max_conns", cfg.Database.MaxOpenConns,
		"min_conns", cfg.Database.MaxIdleConns,
		"databases", dbNames,
	)

	// Initialize database queries (uses default pool, but queries can be context-aware)
	queries := db.NewQueries(pool)

	// Initialize NeuronDB client with config (feature flags + capability checks)
	neurondbClient := neurondb.NewClientWithConfig(neurondbPool, &cfg.NeuronDB)
	if extVersion, err := neurondbClient.ExtensionVersion(ctx); err != nil {
		logger.Warn("NeuronDB extension not found or version check failed; vector/ML/RAG ops may fail", "error", err)
	} else {
		logger.Info("NeuronDB extension ready", "version", extVersion)
	}

	// Initialize cache (Redis with memory fallback)
	redisURL := os.Getenv("REDIS_URL")
	var cacheService cache.CacheInterface
	if redisURL != "" {
		redisCache, err := cache.NewRedisCache(redisURL, 5*time.Minute)
		if err == nil {
			cacheService = redisCache
			defer redisCache.Close()
			logger.Info("Redis cache initialized", "url", redisURL)
		} else {
			logger.Warn("Failed to initialize Redis cache, using memory cache only", "error", err)
			cacheService = cache.NewMemoryCache(5 * time.Minute)
		}
	} else {
		// Use existing cache implementation as fallback
		existingCache := cache.NewCache(5*time.Minute, 10000)
		cacheService = cache.NewAdapter(existingCache)
		logger.Info("Using in-memory cache (no Redis configured)")
	}

	// Create router
	router := mux.NewRouter()

	// Apply global middleware (order matters)
	router.Use(middleware.Recovery)        // Recover from panics
	router.Use(middleware.RequestID)       // Add request ID
	router.Use(middleware.SecurityHeaders) // Security headers
	router.Use(middleware.HTTPLogging)     // Request/response logging

	// Apply caching middleware (before auth to cache public endpoints)
	cacheConfig := middleware.DefaultCacheConfig(cacheService)
	router.Use(middleware.CacheMiddleware(cacheConfig))

	// Apply compression middleware
	router.Use(middleware.CompressionMiddleware)

	// Note: CORS is applied at server level to ensure it runs for all requests

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

	// Database connection test endpoint (no auth required - must be before apiRouter)
	// CORS middleware handles OPTIONS preflight requests, router just needs to allow them through
	router.HandleFunc("/api/v1/database/test", func(w http.ResponseWriter, r *http.Request) {
		// OPTIONS is handled by CORS middleware (returns before reaching here)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handlers.TestDatabaseConnection(w, r)
	}).Methods("POST", "OPTIONS")

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

	// Initialize metrics service (metric catalog) - needed for query rewriter
	metricsCatalogService := catalog.NewMetricsService(pool)
	semanticCatalogService := catalog.NewSemanticService(pool)

	// Initialize warehouse service with query rewriter
	warehouseService := warehouse.NewService(pool, agentClient, neurondbClient, mcpClient)
	queryRewriterService := semantic.NewQueryRewriter(pool, semanticCatalogService, metricsCatalogService)
	queryRewriterHandler := handlers.NewQueryRewriterHandler(queryRewriterService)
	warehouseHandler := handlers.NewWarehouseHandler(warehouseService)

	// Initialize saved search service
	savedSearchService := warehouse.NewSavedSearchService(pool)
	savedSearchHandler := handlers.NewSavedSearchHandler(savedSearchService, warehouseService)

	// Initialize governance service
	governanceService := warehouse.NewGovernanceService(pool)
	governanceHandler := handlers.NewGovernanceHandler(governanceService)

	// Initialize warehouse cache service (for query result caching)
	warehouseCacheService := warehouse.NewCacheService(pool)
	cacheHandler := handlers.NewCacheHandler(warehouseCacheService)

	// Initialize workload service (elastic analytics queues)
	workloadService := execution.NewWorkloadService(pool)
	workloadHandler := handlers.NewWorkloadHandler(workloadService)

	// Initialize data products service (shares / consumption layer)
	dataProductService := warehouse.NewDataProductService(pool)
	dataProductsHandler := handlers.NewDataProductsHandler(dataProductService)

	// Initialize notebooks service
	notebookService := ml.NewNotebookService(pool)
	notebooksHandler := handlers.NewNotebooksHandler(notebookService)

	// Initialize workflow service
	workflowService := workflows.NewService(pool, agentClient, neurondbClient, mcpClient)
	workflowHandler := handlers.NewWorkflowHandler(workflowService)
	workflowOrchestrator := workflows.NewOrchestrator(pool)
	workflowEnhancedHandler := handlers.NewWorkflowEnhancedHandler(workflowOrchestrator, workflowService)

	// Initialize compliance services
	complianceService := compliance.NewService(pool, neurondbClient)
	anomalyService := compliance.NewAnomalyService(pool, neurondbClient)
	policyService := compliance.NewPolicyService(pool, neurondbClient)
	complianceHandler := handlers.NewComplianceHandler(complianceService, anomalyService, policyService)

	// Initialize analytics service
	analyticsService := analytics.NewService(pool, neurondbClient, mcpClient)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	// Initialize models service
	modelsService := models.NewService(pool, neurondbClient, mcpClient)
	qualityScorer := models.NewQualityScorer(pool)
	modelsHandler := handlers.NewModelHandler(modelsService)
	modelQualityHandler := handlers.NewModelQualityHandler(qualityScorer)

	// Initialize integrations services
	integrationsService := integrations.NewIntegrationsService(pool)
	helpdeskService := integrations.NewHelpdeskService(pool)
	webhookService := integrations.NewWebhookService(pool)
	integrationHandler := handlers.NewIntegrationHandler(integrationsService, helpdeskService, webhookService)

	// Initialize alerts service
	alertsService := alerts.NewService(pool, anomalyService)
	alertsHandler := handlers.NewAlertsHandler(alertsService)

	// Initialize support service
	supportService := support.NewService(queries, pool, agentClient, neurondbClient)
	supportHandler := handlers.NewSupportHandler(supportService)

	// Initialize knowledge graph service
	knowledgeGraphService := knowledgegraph.NewService(pool, neurondbClient)
	reasoningService := knowledgegraph.NewReasoningService(pool, knowledgeGraphService)
	knowledgeGraphHandler := handlers.NewKnowledgeGraphHandlerWithReasoning(knowledgeGraphService, reasoningService)

	// Initialize data sources service
	dataSourceService := datasources.NewDataSourceService(pool)
	dataSourceHandler := handlers.NewDataSourceHandler(dataSourceService)

	// Initialize business metrics service (semantic layer)
	businessMetricsService := metrics.NewMetricsService(pool)

	// Initialize glossary linker service
	glossaryLinkerService := catalog.NewGlossaryLinker(pool)
	businessMetricsHandler := handlers.NewMetricsHandler(businessMetricsService, metricsCatalogService)
	glossaryLinkHandler := handlers.NewGlossaryLinkHandler(glossaryLinkerService)

	// Initialize agents service
	agentsService := agents.NewAgentsService(pool, agentClient)
	agentsHandler := handlers.NewAgentsHandler(agentsService)

	// Initialize agent observability services
	agentTracingService := agent.NewTracingService(pool)
	agentEvidenceTracker := agent.NewEvidenceTracker(pool)
	agentHallucinationDetector := agent.NewHallucinationDetector(pool)
	agentAuditTrailService := agent.NewAuditTrailService(pool)

	// Initialize observability service
	observabilityHandler := handlers.NewObservabilityHandler(pool)
	agentObservabilityHandler := handlers.NewAgentObservabilityHandler(
		agentTracingService,
		agentEvidenceTracker,
		agentHallucinationDetector,
		agentAuditTrailService,
	)

	// Initialize model governance handler
	modelGovernanceHandler := handlers.NewModelGovernanceHandler(pool)
	promptTemplateService := governance.NewPromptTemplateService(pool)
	approvalWorkflowService := governance.NewApprovalWorkflowService(pool)
	uiRLSService := governance.NewUIRLSService(pool)
	promptTemplateHandler := handlers.NewPromptTemplateHandler(promptTemplateService)
	approvalWorkflowHandler := handlers.NewApprovalWorkflowHandler(approvalWorkflowService)
	uiRLSHandler := handlers.NewUIRLSHandler(uiRLSService)

	// Initialize decision dashboards
	decisionDashboardService := governance.NewDecisionDashboardService(pool)
	decisionDashboardsHandler := handlers.NewDecisionDashboardsHandler(decisionDashboardService)

	// Initialize ITSM
	itsmService := itsm.NewService(pool)
	itsmHandler := handlers.NewITSMHandler(itsmService)

	// Initialize collaboration handler (internally uses its own services)
	collaborationHandler := handlers.NewCollaborationHandler(pool)

	// Initialize execution services
	replicaService := execution.NewReplicaService(pool)
	shardService := execution.NewShardService(pool)
	jobQueueService := execution.NewJobQueueService(pool)
	priorityQueueManager := execution.NewPriorityQueueManager()
	distributedExecutor := execution.NewDistributedExecutor(pool, cfg.Execution.NodeID, cfg.Execution.MaxConcurrent)
	_ = replicaService       // Used for read routing
	_ = shardService         // Used for sharding
	_ = jobQueueService      // Used for job queue
	_ = priorityQueueManager // Used for priority queues
	_ = distributedExecutor  // Used for distributed execution

	// Initialize quota service
	quotaService := tenancy.NewQuotaService(pool)
	isolationService := tenancy.NewIsolationService(pool)
	_ = quotaService     // Used for resource limits
	_ = isolationService // Used for tenant isolation

	// Initialize row security service
	rowSecurityService := auth.NewRowSecurityService(queries)

	// Initialize policy-aware service
	policyAwareService := semantic.NewPolicyAwareService(pool, rowSecurityService)
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
	metricsCollector := metrics.NewMetricsCollector(pool)
	metricsEnhancedHandler := handlers.NewMetricsEnhancedHandler(metricsCollector)

	// Initialize tracing service
	tracerService := tracing.NewTracerService(cfg.Observability.EnableTracing)
	router.Use(middleware.Tracing(tracerService))

	// Apply timeout middleware to API routes
	timeoutConfig := middleware.DefaultTimeoutConfig()
	apiRouter.Use(middleware.TimeoutByRoute(timeoutConfig))

	// Initialize lineage service from lineage package (separate from semantic lineage)
	lineageServiceFromPackage := lineage.NewLineageService(pool)
	lineageHandler := handlers.NewLineageHandler(lineageServiceFromPackage)

	// Audit service is initialized above with new audit package

	// Initialize billing service
	billingService := billing.NewBillingService(pool)
	billingHandler := handlers.NewBillingHandler(billingService)

	// Initialize versioning service
	versioningService := versioning.NewVersioningService(pool)
	versioningHandler := handlers.NewVersioningHandler(versioningService)

	// Initialize catalog service with NeuronDB client for multimodal embeddings
	catalogService := catalog.NewCatalogServiceWithNeuronDB(pool, neurondbClient)
	catalogHandler := handlers.NewCatalogHandler(catalogService)

	// Initialize semantic definitions service
	// semanticCatalogService is created above and used by queryRewriterService

	// Initialize unified AI service for orchestration
	unifiedAIService := ai.NewUnifiedAIService(neurondbClient, mcpClient, agentClient)
	unifiedAIHandler := handlers.NewUnifiedAIHandler(unifiedAIService)

	// Initialize unified RAG service
	unifiedRAGService := rag.NewUnifiedRAGService(neurondbClient, mcpClient, agentClient)
	unifiedRAGHandler := handlers.NewUnifiedRAGHandler(unifiedRAGService)

	// Initialize ingestion service
	ingestionService := ingestion.NewService(pool, mcpClient)
	// dataQualityValidator := ingestion.NewValidator(pool) // Validator not yet implemented
	// _ = dataQualityValidator // Used for data quality checks
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
	auditService := audit.NewAuditService(pool)
	auditHandler := handlers.NewAuditHandler(auditService)

	// Initialize session manager
	sessionManager := session.NewManager(
		pool,
		cfg.Auth.Session.AccessTokenTTL,
		cfg.Auth.Session.RefreshTokenTTL,
		cfg.Auth.Session.CookieDomain,
		cfg.Auth.Session.CookieSecure,
		cfg.Auth.Session.CookieSameSite,
	)

	// Start session cleanup service (runs every hour)
	var cleanupService *session.CleanupService
	cleanupService = session.NewCleanupService(pool, 1*time.Hour)
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
	// tenancyService := tenancy.NewTenancyService(pool, tenancyMode)

	// Initialize policy engine
	policyEngine := policy.NewPolicyEngine(pool)
	policyHandler := handlers.NewPolicyHandler(policyEngine)

	// Initialize webhook handler (webhookService already initialized above)
	webhookHandler := handlers.NewWebhookHandler(webhookService)

	// Semantic search routes - commented out as methods don't exist yet
	// apiRouter.HandleFunc("/semantic/search", semanticHandler.Search).Methods("POST")
	// apiRouter.HandleFunc("/semantic/rag", semanticHandler.RAG).Methods("POST")
	// apiRouter.HandleFunc("/semantic/documents", semanticHandler.CreateDocument).Methods("POST")
	// apiRouter.HandleFunc("/semantic/documents/{id}", semanticHandler.UpdateDocument).Methods("PUT")
	// apiRouter.HandleFunc("/semantic/collections/{id}", semanticHandler.GetCollection).Methods("GET")

	// Unified AI routes
	apiRouter.HandleFunc("/ai/embedding", unifiedAIHandler.GenerateEmbedding).Methods("POST")
	apiRouter.HandleFunc("/ai/workflow", unifiedAIHandler.ExecuteWorkflow).Methods("POST")
	apiRouter.HandleFunc("/ai/register-tools", unifiedAIHandler.RegisterTools).Methods("POST")

	// Unified RAG routes
	apiRouter.HandleFunc("/rag/query", unifiedRAGHandler.PerformRAG).Methods("POST")
	apiRouter.HandleFunc("/rag/query/stream", unifiedRAGHandler.PerformRAGStream).Methods("POST")
	apiRouter.HandleFunc("/rag/status", unifiedRAGHandler.GetRAGStatus).Methods("GET")
	// AI assistant: policy-aware RAG - same pipeline as /rag/query, use with user context for policy filtering
	apiRouter.HandleFunc("/ai/assistant", unifiedRAGHandler.PerformRAG).Methods("POST")

	// Pipeline routes
	apiRouter.HandleFunc("/semantic/pipelines", pipelineHandler.CreatePipeline).Methods("POST")
	apiRouter.HandleFunc("/semantic/pipelines/{id}", pipelineHandler.GetPipeline).Methods("GET")
	apiRouter.HandleFunc("/semantic/pipelines/{id}/versions", pipelineHandler.ListPipelineVersions).Methods("GET")
	apiRouter.HandleFunc("/semantic/pipelines/{id}/replay", pipelineHandler.ReplayPipeline).Methods("POST")
	apiRouter.HandleFunc("/semantic/pipelines/{id}/activate", pipelineHandler.ActivatePipeline).Methods("POST")

	// Warehouse routes
	apiRouter.HandleFunc("/warehouse/query", warehouseHandler.Query).Methods("POST")
	apiRouter.HandleFunc("/warehouse/query/rewrite", queryRewriterHandler.RewriteQuery).Methods("POST")
	apiRouter.HandleFunc("/warehouse/query/semantics", queryRewriterHandler.GetQuerySemantics).Methods("GET")
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

	// Workload routes (elastic analytics queues)
	apiRouter.HandleFunc("/warehouse/workload/queues", workloadHandler.ListQueues).Methods("GET")
	apiRouter.HandleFunc("/warehouse/workload/queues", workloadHandler.CreateQueue).Methods("POST")
	apiRouter.HandleFunc("/warehouse/workload/queues/{name}", workloadHandler.GetQueueByName).Methods("GET")

	// Data products routes (shares / consumption layer)
	apiRouter.HandleFunc("/data-products", dataProductsHandler.ListDataProducts).Methods("GET")
	apiRouter.HandleFunc("/data-products", dataProductsHandler.CreateDataProduct).Methods("POST")
	apiRouter.HandleFunc("/data-products/{id}", dataProductsHandler.GetDataProduct).Methods("GET")
	apiRouter.HandleFunc("/data-products/{id}/share", dataProductsHandler.ShareDataProduct).Methods("POST")
	apiRouter.HandleFunc("/data-products/{id}/revoke", dataProductsHandler.RevokeDataProduct).Methods("POST")
	apiRouter.HandleFunc("/data-products/{id}/consumers", dataProductsHandler.ListConsumers).Methods("GET")

	// Notebooks routes
	apiRouter.HandleFunc("/notebooks", notebooksHandler.ListNotebooks).Methods("GET")
	apiRouter.HandleFunc("/notebooks", notebooksHandler.CreateNotebook).Methods("POST")
	apiRouter.HandleFunc("/notebooks/{id}", notebooksHandler.GetNotebook).Methods("GET")
	apiRouter.HandleFunc("/notebooks/{id}/cells", notebooksHandler.ListCells).Methods("GET")
	apiRouter.HandleFunc("/notebooks/{id}/cells", notebooksHandler.AddCell).Methods("POST")
	apiRouter.HandleFunc("/notebooks/{id}/runs", notebooksHandler.ListRuns).Methods("GET")
	apiRouter.HandleFunc("/notebooks/{id}/runs", notebooksHandler.CreateRun).Methods("POST")

	// Workflow routes
	apiRouter.HandleFunc("/workflows", workflowHandler.ListWorkflows).Methods("GET")
	// apiRouter.HandleFunc("/workflows/{id}/stream", workflowHandler.StreamExecution).Methods("GET") // WebSocket for live execution - not yet implemented
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
	apiRouter.HandleFunc("/workflows/{id}/execute-distributed", workflowEnhancedHandler.ExecuteDistributedWorkflow).Methods("POST")
	apiRouter.HandleFunc("/workflows/{id}/metrics", workflowEnhancedHandler.GetWorkflowMetrics).Methods("GET")
	apiRouter.HandleFunc("/workflows/{id}/cost", workflowEnhancedHandler.GetWorkflowCost).Methods("GET")
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
	apiRouter.HandleFunc("/compliance/report/export", complianceHandler.ExportComplianceReport).Methods("GET")

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
	apiRouter.HandleFunc("/models/quality/score", modelQualityHandler.ScoreOutput).Methods("POST")
	apiRouter.HandleFunc("/models/{id}/quality/scores", modelQualityHandler.GetScores).Methods("GET")
	apiRouter.HandleFunc("/models/{id}/quality/average", modelQualityHandler.GetAverageScore).Methods("GET")

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
	apiRouter.HandleFunc("/knowledge-graph/reason", knowledgeGraphHandler.Reason).Methods("POST")
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
	apiRouter.HandleFunc("/agents/{id}/runs", agentsHandler.GetRuns).Methods("GET")
	apiRouter.HandleFunc("/agents/{id}/memory", agentsHandler.GetMemory).Methods("GET")
	apiRouter.HandleFunc("/agents/{id}/evaluations", agentsHandler.GetEvaluations).Methods("GET")

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

	// Agent observability routes (traces, evidence, hallucination, audit)
	apiRouter.HandleFunc("/observability/agent/traces", agentObservabilityHandler.GetTraces).Methods("GET")
	apiRouter.HandleFunc("/observability/agent/traces/{id}", agentObservabilityHandler.GetTrace).Methods("GET")
	apiRouter.HandleFunc("/observability/agent/evidence-coverage", agentObservabilityHandler.GetEvidenceCoverage).Methods("GET")
	apiRouter.HandleFunc("/observability/agent/hallucination-risk", agentObservabilityHandler.GetHallucinationRisk).Methods("GET")
	apiRouter.HandleFunc("/observability/agent/audit-trail", agentObservabilityHandler.GetAuditTrail).Methods("GET")

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
	apiRouter.HandleFunc("/audit/export", auditHandler.ExportAuditEvents).Methods("GET")

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
	apiRouter.HandleFunc("/catalog/glossary/link", glossaryLinkHandler.LinkGlossary).Methods("POST")
	apiRouter.HandleFunc("/catalog/glossary/{id}/links", glossaryLinkHandler.GetGlossaryLinks).Methods("GET")
	apiRouter.HandleFunc("/catalog/glossary/entity/{entity_type}/{entity_id}/links", glossaryLinkHandler.GetEntityGlossaryLinks).Methods("GET")

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
	ssoService := auth.NewSSOService(pool, ssoConfig)
	ssoHandler := handlers.NewSSOHandler(ssoService)

	// Initialize comments service
	commentsService := comments.NewService(pool)
	commentsHandler := handlers.NewCommentsHandler(commentsService)

	// Initialize ownership service (separate from semantic metric ownership)
	resourceOwnershipService := ownership.NewService(pool)
	ownershipHandler := handlers.NewOwnershipHandler(resourceOwnershipService)

	// Initialize webhooks service (new implementation)
	webhooksService := webhooks.NewService(pool)
	webhooksHandler := handlers.NewWebhooksHandler(webhooksService)

	// Initialize connector framework service
	connectorService := connectors.NewConnectorService(pool)
	connectorHandler := handlers.NewConnectorHandler(connectorService)

	// Initialize data quality service
	// Initialize data quality service with MCP and Agent clients for ML-powered analysis
	dataQualityService := dataquality.NewServiceWithMCPAndAgent(pool, neurondbClient, mcpClient, agentClient)
	dataQualityHandler := handlers.NewDataQualityHandler(dataQualityService)

	// Initialize profiling service
	profilingService := profiling.NewService(pool, neurondbClient, mcpClient)
	profilingHandler := handlers.NewProfilingHandler(profilingService)

	// Initialize classification service
	classificationService := classification.NewService(pool)
	classificationHandler := handlers.NewClassificationHandler(classificationService)

	// Initialize column lineage service
	columnLineageService := lineage.NewColumnLineageService(pool)
	columnLineageHandler := handlers.NewColumnLineageHandler(columnLineageService)

	// Initialize region service (Phase 2.1)
	regionService := tenancy.NewRegionService(pool)
	regionHandler := handlers.NewRegionHandler(regionService)

	// Initialize backup service (Phase 2.2)
	backupConfig := backup.BackupConfig{
		BackupDir:     "/var/backups/neuronip",
		RetentionDays: 30,
		Compress:      true,
	}
	backupService := backup.NewBackupService(pool, backupConfig)
	backupHandler := handlers.NewBackupHandler(backupService)

	// Initialize column and row security services (Phase 2.3 - already implemented)
	// These services are available for integration with query handlers
	_ = auth.NewColumnSecurityService(queries)
	_ = auth.NewRowSecurityService(queries)

	// Initialize DSAR service (Phase 2.4)
	dsarService := compliance.NewDSARService(pool)
	dsarHandler := handlers.NewDSARHandler(dsarService)

	// Initialize PIA service (Phase 2.4)
	piaService := compliance.NewPIAService(pool)
	piaHandler := handlers.NewPIAHandler(piaService)

	// Initialize consent service (Phase 2.4)
	consentService := compliance.NewConsentService(pool)
	consentHandler := handlers.NewConsentHandler(consentService)

	// Initialize masking service (Phase 2.5)
	maskingService := masking.NewMaskingService(pool)
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

	// Blocks and page/database templates
	blocksService := blocks.NewService(pool)
	blocksHandler := blocks.NewHandler(blocksService)
	templatesService := blocks.NewTemplatesService(pool)
	notionTemplatesHandler := notion.NewTemplatesHandler(templatesService)
	apiRouter.HandleFunc("/blocks", blocksHandler.GetBlocks).Methods("GET")
	apiRouter.HandleFunc("/blocks", blocksHandler.CreateBlock).Methods("POST")
	apiRouter.HandleFunc("/blocks/{id}", blocksHandler.UpdateBlock).Methods("PATCH")
	apiRouter.HandleFunc("/blocks/{id}", blocksHandler.DeleteBlock).Methods("DELETE")
	apiRouter.HandleFunc("/blocks/reorder", blocksHandler.ReorderBlocks).Methods("POST")
	apiRouter.HandleFunc("/notion-ui/templates/pages", notionTemplatesHandler.ListPageTemplates).Methods("GET")
	apiRouter.HandleFunc("/notion-ui/templates/databases", notionTemplatesHandler.ListDatabaseTemplates).Methods("GET")

	// Databases routes
	databasesService := databases.NewService(pool)
	databasesHandler := databases.NewHandler(databasesService)
	apiRouter.HandleFunc("/databases/{id}", databasesHandler.GetDatabase).Methods("GET")
	apiRouter.HandleFunc("/databases", databasesHandler.CreateDatabase).Methods("POST")
	apiRouter.HandleFunc("/databases/{id}/rows/{rowId}", databasesHandler.UpdateRow).Methods("PATCH")
	apiRouter.HandleFunc("/databases/{id}/rows", databasesHandler.CreateRow).Methods("POST")
	apiRouter.HandleFunc("/databases/{id}/rows/{rowId}", databasesHandler.DeleteRow).Methods("DELETE")
	apiRouter.HandleFunc("/databases/{id}/view-preferences", databasesHandler.UpdateViewPreferences).Methods("PATCH")
	apiRouter.HandleFunc("/databases/{id}/view-preferences", databasesHandler.GetViewPreferences).Methods("GET")

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

	// Decision dashboards
	apiRouter.HandleFunc("/decision-dashboards", decisionDashboardsHandler.List).Methods("GET")
	apiRouter.HandleFunc("/decision-dashboards", decisionDashboardsHandler.Create).Methods("POST")
	apiRouter.HandleFunc("/decision-dashboards/{id}", decisionDashboardsHandler.Get).Methods("GET")
	apiRouter.HandleFunc("/decision-dashboards/{id}/runs", decisionDashboardsHandler.RecordRun).Methods("POST")

	// ITSM
	apiRouter.HandleFunc("/itsm/incidents", itsmHandler.ListIncidents).Methods("GET")
	apiRouter.HandleFunc("/itsm/incidents", itsmHandler.CreateIncident).Methods("POST")
	apiRouter.HandleFunc("/itsm/changes", itsmHandler.ListChanges).Methods("GET")
	apiRouter.HandleFunc("/itsm/changes", itsmHandler.CreateChange).Methods("POST")
	apiRouter.HandleFunc("/itsm/runbooks", itsmHandler.ListRunbooks).Methods("GET")
	apiRouter.HandleFunc("/itsm/runbooks", itsmHandler.CreateRunbook).Methods("POST")

	// Enterprise Feature Routes - Collaboration
	apiRouter.HandleFunc("/collaboration/dashboards", collaborationHandler.CreateSharedDashboard).Methods("POST")
	apiRouter.HandleFunc("/collaboration/dashboards", collaborationHandler.GetSharedDashboards).Methods("GET")
	apiRouter.HandleFunc("/collaboration/dashboards/{id}/comments", collaborationHandler.AddDashboardComment).Methods("POST")
	apiRouter.HandleFunc("/collaboration/dashboards/{id}/comments", collaborationHandler.GetDashboardComments).Methods("GET")
	apiRouter.HandleFunc("/collaboration/answer-cards", collaborationHandler.CreateAnswerCard).Methods("POST")
	apiRouter.HandleFunc("/collaboration/saved-questions", collaborationHandler.SaveQuestion).Methods("POST")
	apiRouter.HandleFunc("/collaboration/annotations", collaborationHandler.CreateAnnotation).Methods("POST")
	apiRouter.HandleFunc("/collaboration/annotations", collaborationHandler.GetAnnotations).Methods("GET")
	apiRouter.HandleFunc("/collaboration/threads", collaborationHandler.CreateThread).Methods("POST")
	apiRouter.HandleFunc("/collaboration/threads", collaborationHandler.GetThreads).Methods("GET")
	apiRouter.HandleFunc("/collaboration/threads/{id}", collaborationHandler.GetThread).Methods("GET")
	apiRouter.HandleFunc("/collaboration/threads/{id}/posts", collaborationHandler.AddPost).Methods("POST")
	apiRouter.HandleFunc("/collaboration/decisions", collaborationHandler.RecordDecision).Methods("POST")
	apiRouter.HandleFunc("/collaboration/decisions", collaborationHandler.GetDecisionHistory).Methods("GET")

	// Enterprise Feature Routes - Governance (RLS)
	rlsHandler := handlers.NewRLSHandler(queries)
	apiRouter.HandleFunc("/governance/rls/policies", rlsHandler.GetRLSPolicies).Methods("GET")
	apiRouter.HandleFunc("/governance/rls/policies", rlsHandler.CreateRLSPolicy).Methods("POST")

	// Governance - Prompt Templates
	apiRouter.HandleFunc("/governance/prompts", promptTemplateHandler.CreatePromptTemplate).Methods("POST")
	apiRouter.HandleFunc("/governance/prompts", promptTemplateHandler.ListPromptTemplates).Methods("GET")
	apiRouter.HandleFunc("/governance/prompts/{id}", promptTemplateHandler.GetPromptTemplate).Methods("GET")
	apiRouter.HandleFunc("/governance/prompts/{id}/approve", promptTemplateHandler.ApprovePromptTemplate).Methods("POST")

	// Governance - Approval Workflows
	apiRouter.HandleFunc("/governance/approvals", approvalWorkflowHandler.CreateWorkflow).Methods("POST")
	apiRouter.HandleFunc("/governance/approvals/{id}", approvalWorkflowHandler.GetWorkflow).Methods("GET")
	apiRouter.HandleFunc("/governance/approvals/{id}/submit", approvalWorkflowHandler.SubmitApproval).Methods("POST")

	// Governance - UI RLS Policies
	apiRouter.HandleFunc("/governance/ui-rls/policies", uiRLSHandler.CreateRLSPolicy).Methods("POST")
	apiRouter.HandleFunc("/governance/ui-rls/policies", uiRLSHandler.GetRLSPolicies).Methods("GET")
	apiRouter.HandleFunc("/governance/ui-rls/policies/{id}/toggle", uiRLSHandler.ToggleRLSPolicy).Methods("POST")

	// Enterprise Feature Routes - Resource Quotas
	quotaHandler := handlers.NewQuotaHandler(pool)
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

	// Wrap router with CORS handler to ensure it runs for all requests (including unmatched routes)
	corsHandler := middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		AllowedMethods: cfg.CORS.AllowedMethods,
		AllowedHeaders: cfg.CORS.AllowedHeaders,
	})(router)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      corsHandler,
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
