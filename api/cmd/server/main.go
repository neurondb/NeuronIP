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
	"github.com/neurondb/NeuronIP/api/internal/alerts"
	"github.com/neurondb/NeuronIP/api/internal/analytics"
	"github.com/neurondb/NeuronIP/api/internal/audit"
	"github.com/neurondb/NeuronIP/api/internal/auth"
	"github.com/neurondb/NeuronIP/api/internal/billing"
	"github.com/neurondb/NeuronIP/api/internal/catalog"
	"github.com/neurondb/NeuronIP/api/internal/compliance"
	"github.com/neurondb/NeuronIP/api/internal/config"
	"github.com/neurondb/NeuronIP/api/internal/datasources"
	"github.com/neurondb/NeuronIP/api/internal/db"
	"github.com/neurondb/NeuronIP/api/internal/handlers"
	"github.com/neurondb/NeuronIP/api/internal/ingestion"
	"github.com/neurondb/NeuronIP/api/internal/ingestion/connectors"
	"github.com/neurondb/NeuronIP/api/internal/integrations"
	"github.com/neurondb/NeuronIP/api/internal/knowledgegraph"
	"github.com/neurondb/NeuronIP/api/internal/lineage"
	"github.com/neurondb/NeuronIP/api/internal/logging"
	"github.com/neurondb/NeuronIP/api/internal/metrics"
	"github.com/neurondb/NeuronIP/api/internal/middleware"
	"github.com/neurondb/NeuronIP/api/internal/models"
	"github.com/neurondb/NeuronIP/api/internal/neurondb"
	"github.com/neurondb/NeuronIP/api/internal/observability"
	"github.com/neurondb/NeuronIP/api/internal/policy"
	"github.com/neurondb/NeuronIP/api/internal/semantic"
	"github.com/neurondb/NeuronIP/api/internal/support"
	"github.com/neurondb/NeuronIP/api/internal/versioning"
	"github.com/neurondb/NeuronIP/api/internal/warehouse"
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

	// Create database connection pool
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Error("Failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("Database pool created successfully",
		"max_conns", cfg.Database.MaxOpenConns,
		"min_conns", cfg.Database.MaxIdleConns,
	)

	// Initialize database queries
	queries := db.NewQueries(pool.Pool)

	// Initialize NeuronDB client
	neurondbClient := neurondb.NewClient(pool.Pool)

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

	// Health check endpoint (no auth required)
	healthHandler := handlers.NewHealthHandler(pool.Pool)
	router.Handle("/health", healthHandler).Methods("GET")
	router.Handle("/api/v1/health", healthHandler).Methods("GET")

	// Metrics endpoint (no auth required) - Prometheus metrics
	router.Handle("/metrics", metrics.Handler()).Methods("GET")

	// API routes (require auth)
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	if cfg.Auth.EnableAPIKeys {
		apiRouter.Use(auth.Middleware(queries))
	}

	// Initialize agent client
	agentClient := agent.NewClient(cfg.NeuronAgent.Endpoint, cfg.NeuronAgent.APIKey)

	// Initialize services
	semanticService := semantic.NewService(queries, pool.Pool, neurondbClient)
	semanticHandler := handlers.NewSemanticHandler(semanticService)

	// Initialize warehouse service
	warehouseService := warehouse.NewService(pool.Pool, agentClient, neurondbClient)
	warehouseHandler := handlers.NewWarehouseHandler(warehouseService)

	// Initialize workflow service
	workflowService := workflows.NewService(pool.Pool, agentClient, neurondbClient)
	workflowHandler := handlers.NewWorkflowHandler(workflowService)

	// Initialize compliance services
	complianceService := compliance.NewService(pool.Pool, neurondbClient)
	anomalyService := compliance.NewAnomalyService(pool.Pool, neurondbClient)
	complianceHandler := handlers.NewComplianceHandler(complianceService, anomalyService)

	// Initialize analytics service
	analyticsService := analytics.NewService(pool.Pool)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)

	// Initialize models service
	modelsService := models.NewService(pool.Pool, neurondbClient)
	modelsHandler := handlers.NewModelHandler(modelsService)

	// Initialize integrations service
	helpdeskService := integrations.NewHelpdeskService(pool.Pool)
	integrationHandler := handlers.NewIntegrationHandler(helpdeskService)

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
	businessMetricsHandler := handlers.NewMetricsHandler(businessMetricsService)

	// Initialize agents service
	agentsService := agents.NewAgentsService(pool.Pool)
	agentsHandler := handlers.NewAgentsHandler(agentsService)

	// Initialize observability service
	observabilityService := observability.NewObservabilityService(pool.Pool)
	observabilityHandler := handlers.NewObservabilityHandler(observabilityService)

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

	// Initialize catalog service
	catalogService := catalog.NewCatalogService(pool.Pool)
	catalogHandler := handlers.NewCatalogHandler(catalogService)

	// Initialize metrics service (metric catalog)
	metricsCatalogService := catalog.NewMetricsService(pool.Pool)
	metricsCatalogHandler := handlers.NewMetricsHandler(metricsCatalogService)

	// Initialize semantic definitions service
	semanticDefinitionsService := catalog.NewSemanticService(pool.Pool)
	// Note: semantic search handler uses semantic.Service, not catalog.SemanticService

	// Initialize ingestion service
	ingestionService := ingestion.NewService(pool.Pool)
	// Register connector factories to avoid import cycles
	ingestionService.RegisterConnectorFactory("zendesk", func(ct string) ingestion.Connector {
		return connectors.NewZendeskConnector()
	})
	ingestionService.RegisterConnectorFactory("salesforce", func(ct string) ingestion.Connector {
		return connectors.NewSalesforceConnector()
	})
	ingestionService.RegisterConnectorFactory("slack", func(ct string) ingestion.Connector {
		return connectors.NewSlackConnector()
	})
	ingestionService.RegisterConnectorFactory("teams", func(ct string) ingestion.Connector {
		return connectors.NewTeamsConnector()
	})
	ingestionHandler := handlers.NewIngestionHandler(ingestionService)

	// Initialize audit service
	auditService := audit.NewAuditService(pool.Pool)
	auditHandler := handlers.NewAuditHandler(auditService)

	// Initialize policy engine
	policyEngine := policy.NewPolicyEngine(pool.Pool)
	policyHandler := handlers.NewPolicyHandler(policyEngine)

	// Initialize webhook service
	webhookService := integrations.NewWebhookService(pool.Pool)
	webhookHandler := handlers.NewWebhookHandler(webhookService)

	// Semantic search routes
	apiRouter.HandleFunc("/semantic/search", semanticHandler.Search).Methods("POST")
	apiRouter.HandleFunc("/semantic/rag", semanticHandler.RAG).Methods("POST")
	apiRouter.HandleFunc("/semantic/documents", semanticHandler.CreateDocument).Methods("POST")
	apiRouter.HandleFunc("/semantic/documents/{id}", semanticHandler.UpdateDocument).Methods("PUT")
	apiRouter.HandleFunc("/semantic/collections/{id}", semanticHandler.GetCollection).Methods("GET")

	// Warehouse routes
	apiRouter.HandleFunc("/warehouse/query", warehouseHandler.Query).Methods("POST")
	apiRouter.HandleFunc("/warehouse/queries/{id}", warehouseHandler.GetQuery).Methods("GET")
	apiRouter.HandleFunc("/warehouse/schemas", warehouseHandler.ListSchemas).Methods("GET")
	apiRouter.HandleFunc("/warehouse/schemas", warehouseHandler.CreateSchema).Methods("POST")
	apiRouter.HandleFunc("/warehouse/schemas/{id}", warehouseHandler.GetSchema).Methods("GET")

	// Workflow routes
	apiRouter.HandleFunc("/workflows/{id}/execute", workflowHandler.ExecuteWorkflow).Methods("POST")
	apiRouter.HandleFunc("/workflows/{id}", workflowHandler.GetWorkflow).Methods("GET")

	// Compliance routes
	apiRouter.HandleFunc("/compliance/check", complianceHandler.CheckCompliance).Methods("POST")
	apiRouter.HandleFunc("/compliance/anomalies", complianceHandler.GetAnomalyDetections).Methods("GET")

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
	apiRouter.HandleFunc("/observability/metrics", observabilityHandler.GetSystemMetrics).Methods("GET")
	apiRouter.HandleFunc("/observability/agent-logs", observabilityHandler.GetAgentLogs).Methods("GET")
	apiRouter.HandleFunc("/observability/workflow-logs", observabilityHandler.GetWorkflowLogs).Methods("GET")

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

	// Metric catalog routes
	apiRouter.HandleFunc("/catalog/metrics", metricsCatalogHandler.ListMetrics).Methods("GET")
	apiRouter.HandleFunc("/catalog/metrics", metricsCatalogHandler.CreateMetric).Methods("POST")
	apiRouter.HandleFunc("/catalog/metrics/{id}", metricsCatalogHandler.GetMetric).Methods("GET")
	apiRouter.HandleFunc("/catalog/metrics/{id}/lineage", metricsCatalogHandler.GetMetric).Methods("GET") // TODO: Add lineage endpoint

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

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}
