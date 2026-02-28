package api

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/api/handlers"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/config"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/openclaw"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
	ws "github.com/abelkuruvilla/claw-agent-mission-control/internal/websocket"
)

type Server struct {
	echo             *echo.Echo
	config           *config.Config
	store            *store.Store
	hub              *ws.Hub
	agentSender      *openclaw.AgentSender
	agentHandler     *handlers.AgentHandler
	taskHandler      *handlers.TaskHandler
	projectHandler   *handlers.ProjectHandler
	commentHandler   *handlers.CommentHandler
	reportingHandler *handlers.ReportingHandler
	wsHandler        *handlers.WebSocketHandler
	chatHandler      *handlers.ChatHandler
}

func NewServer(cfg *config.Config, store *store.Store) *Server {
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	// CORS configuration - allow all origins for network access
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
			http.MethodPatch,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			echo.HeaderContentLength,
			echo.HeaderContentType,
		},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}))
	
	e.Use(middleware.Gzip())

	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Create OpenClaw client
	openclawClient, err := openclaw.NewClientFromEnv()
	if err != nil {
		// Log warning but don't fail - client will be nil and chat features won't work
		e.Logger.Warn("Failed to create OpenClaw client: ", err)
	}

	// Build the Mission Control API URL for agent notifications
	mcAPIURL := fmt.Sprintf("http://%s:%d/api/v1", cfg.Host, cfg.Port)
	if cfg.Host == "0.0.0.0" {
		mcAPIURL = fmt.Sprintf("http://127.0.0.1:%d/api/v1", cfg.Port)
	}
	agentSender := openclaw.NewAgentSender(mcAPIURL)

	s := &Server{
		echo:             e,
		config:           cfg,
		store:            store,
		hub:              hub,
		agentSender:      agentSender,
		agentHandler:     handlers.NewAgentHandler(store),
		taskHandler:      handlers.NewTaskHandler(store, hub, agentSender),
		projectHandler:   handlers.NewProjectHandler(store),
		commentHandler:   handlers.NewCommentHandler(store),
		reportingHandler: handlers.NewReportingHandler(store, hub),
		wsHandler:        handlers.NewWebSocketHandler(hub),
		chatHandler:      handlers.NewChatHandler(store, openclawClient),
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	// API v1 routes - all API endpoints under /api/v1
	api := s.echo.Group("/api/v1")

	// Health check
	api.GET("/health", s.healthCheck)

	// Agents
	agents := api.Group("/agents")
	agents.GET("", s.agentHandler.List)
	agents.POST("", s.agentHandler.Create)
	agents.GET("/:id", s.agentHandler.Get)
	agents.PUT("/:id", s.agentHandler.Update)
	agents.DELETE("/:id", s.agentHandler.Delete)

	// Agent Queue
	agents.GET("/:id/queue", s.taskHandler.GetAgentQueue)
	agents.POST("/:id/queue/next", s.taskHandler.DequeueNextTask)

	// Agent Chat
	agentChat := agents.Group("/:id/sessions")
	agentChat.POST("", s.chatHandler.StartSession)
	agentChat.GET("", s.chatHandler.ListSessions)
	agentChat.DELETE("/:sessionId", s.chatHandler.EndSession)
	agentChat.GET("/:sessionId/messages", s.chatHandler.GetMessages)
	agentChat.POST("/:sessionId/messages", s.chatHandler.SendMessage)
	agentChat.GET("/:sessionId/poll", s.chatHandler.PollMessages)

	// Tasks
	tasks := api.Group("/tasks")
	tasks.GET("", s.taskHandler.List)
	tasks.POST("", s.taskHandler.Create)
	tasks.GET("/:id", s.taskHandler.Get)
	tasks.PUT("/:id", s.taskHandler.Update)
	tasks.DELETE("/:id", s.taskHandler.Delete)
	tasks.PUT("/:id/status", s.taskHandler.UpdateStatus)
	tasks.POST("/:id/retry", s.taskHandler.RetryTask)
	tasks.POST("/:id/progress-txt", s.reportingHandler.AppendProgressTxt)
	
	// Task sub-resources
	tasks.GET("/:id/subtasks", s.taskHandler.ListSubtasks)
	tasks.GET("/:id/phases", s.taskHandler.ListPhases)
	tasks.POST("/:id/phases", s.taskHandler.CreatePhase)
	tasks.GET("/:id/stories", s.taskHandler.ListStories)
	tasks.POST("/:id/stories", s.taskHandler.CreateStory)
	
	// Task execution
	tasks.POST("/:id/start", s.taskHandler.StartTask)
	tasks.POST("/:id/stop", s.taskHandler.StopTask)

	// Delegation approval
	tasks.POST("/:id/approve", s.taskHandler.ApproveDelegation)
	tasks.POST("/:id/request-changes", s.taskHandler.RequestChanges)
	
	// Task comments
	tasks.GET("/:id/comments", s.commentHandler.ListByTask)
	tasks.POST("/:id/comments", s.commentHandler.Create)

	// Projects
	projects := api.Group("/projects")
	projects.GET("", s.projectHandler.List)
	projects.POST("", s.projectHandler.Create)
	projects.GET("/:id", s.projectHandler.Get)
	projects.PUT("/:id", s.projectHandler.Update)
	projects.DELETE("/:id", s.projectHandler.Delete)
	projects.GET("/:id/tasks", s.projectHandler.ListTasks)

	// Comments (direct access)
	comments := api.Group("/comments")
	comments.DELETE("/:id", s.commentHandler.Delete)

	// Phases
	phases := api.Group("/phases")
	phases.GET("/:id", s.getPhase)
	phases.PUT("/:id", s.updatePhase)
	phases.POST("/:id/progress", s.reportingHandler.UpdatePhaseProgress)
	phases.POST("/:id/complete", s.reportingHandler.CompletePhase)
	phases.POST("/:id/fail", s.reportingHandler.FailPhase)

	// Stories
	stories := api.Group("/stories")
	stories.GET("/:id", s.getStory)
	stories.PUT("/:id", s.updateStory)
	stories.POST("/:id/pass", s.reportingHandler.PassStory)
	stories.POST("/:id/fail", s.reportingHandler.FailStory)

	// Events
	api.GET("/events", s.listEvents)
	api.POST("/events", s.createEvent)

	// Settings
	api.GET("/settings", s.getSettings)
	api.PUT("/settings", s.updateSettings)
	api.POST("/settings/test-connection", s.testConnection)

	// Status
	api.GET("/status", s.getStatus)

	// Models (from OpenClaw config)
	api.GET("/models", s.listModels)

	// WebSocket
	s.echo.GET("/ws", s.wsHandler.HandleWebSocket)
}

func (s *Server) ServeUI(assets fs.FS) {
	// Serve static files from embedded UI with SPA fallback
	fileServer := http.FileServer(http.FS(assets))
	
	// Handle for SPA - serve index.html for client-side routes
	uiHandler := func(c echo.Context) error {
		path := c.Request().URL.Path
		
		// Don't handle API routes
		if len(path) >= 4 && path[:4] == "/api" {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		
		// Don't handle WebSocket
		if path == "/ws" {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		
		// Clean path - remove leading slash for fs.Open
		cleanPath := path
		if len(cleanPath) > 0 && cleanPath[0] == '/' {
			cleanPath = cleanPath[1:]
		}
		// Remove trailing slash for consistent lookup
		if len(cleanPath) > 0 && cleanPath[len(cleanPath)-1] == '/' {
			cleanPath = cleanPath[:len(cleanPath)-1]
		}
		if cleanPath == "" {
			cleanPath = "index.html"
		}
		
		// Try to open the file to check if it exists
		f, err := assets.Open(cleanPath)
		if err == nil {
			stat, statErr := f.Stat()
			f.Close()
			
			// If it's a directory, serve its index.html directly
			if statErr == nil && stat.IsDir() {
				indexPath := cleanPath + "/index.html"
				indexFile, idxErr := assets.Open(indexPath)
				if idxErr == nil {
					defer indexFile.Close()
					idxStat, _ := indexFile.Stat()
					content := make([]byte, idxStat.Size())
					indexFile.Read(content)
					return c.HTMLBlob(http.StatusOK, content)
				}
			} else if statErr == nil {
				// It's a file, serve it directly
				fileServer.ServeHTTP(c.Response(), c.Request())
				return nil
			}
		}
		
		// Check if cleanPath + "/index.html" exists (for paths without trailing slash)
		indexPath := cleanPath + "/index.html"
		if indexFile, err := assets.Open(indexPath); err == nil {
			defer indexFile.Close()
			idxStat, _ := indexFile.Stat()
			content := make([]byte, idxStat.Size())
			indexFile.Read(content)
			return c.HTMLBlob(http.StatusOK, content)
		}
		
		// File doesn't exist - for SPA routing, serve root index.html
		// But only for non-asset paths (don't serve HTML for .js, .css, etc)
		if isAssetPath(path) {
			return echo.NewHTTPError(http.StatusNotFound, "asset not found")
		}
		
		// Serve root index.html for SPA client-side routing
		indexFile, err := assets.Open("index.html")
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "index.html not found")
		}
		defer indexFile.Close()
		
		stat, _ := indexFile.Stat()
		content := make([]byte, stat.Size())
		indexFile.Read(content)
		
		return c.HTMLBlob(http.StatusOK, content)
	}
	
	// Register for both GET and HEAD methods
	s.echo.GET("/*", uiHandler)
	s.echo.HEAD("/*", uiHandler)
}

// isAssetPath checks if the path is for a static asset that shouldn't fallback to index.html
func isAssetPath(path string) bool {
	assetExtensions := []string{".js", ".css", ".map", ".woff", ".woff2", ".ttf", ".eot", ".svg", ".png", ".jpg", ".jpeg", ".gif", ".ico", ".json"}
	for _, ext := range assetExtensions {
		if len(path) > len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	return false
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return s.echo.Start(addr)
}

func (s *Server) TaskHandler() *handlers.TaskHandler {
	return s.taskHandler
}

func (s *Server) Hub() *ws.Hub {
	return s.hub
}

func (s *Server) AgentSender() *openclaw.AgentSender {
	return s.agentSender
}

// Handler stubs (to be implemented in handlers/)
func (s *Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"version": "1.0.0",
		"status":  "running",
	})
}

// Placeholder handlers - return not implemented for now
func (s *Server) getPhase(c echo.Context) error         { return c.JSON(http.StatusNotImplemented, nil) }
func (s *Server) updatePhase(c echo.Context) error      { return c.JSON(http.StatusNotImplemented, nil) }

func (s *Server) getStory(c echo.Context) error         { return c.JSON(http.StatusNotImplemented, nil) }
func (s *Server) updateStory(c echo.Context) error      { return c.JSON(http.StatusNotImplemented, nil) }

// Events handlers
func (s *Server) listEvents(c echo.Context) error {
	ctx := c.Request().Context()
	
	// Parse limit from query params, default to 50
	limit := int64(50)
	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := parseLimit(limitParam); err == nil {
			limit = l
		}
	}
	
	// Check for task_id or agent_id filters
	taskID := c.QueryParam("task_id")
	agentID := c.QueryParam("agent_id")
	
	var apiEvents []map[string]interface{}
	var err error
	
	if taskID != "" {
		dbEvents, e := s.store.ListEventsByTask(ctx, taskID, limit)
		err = e
		apiEvents = make([]map[string]interface{}, len(dbEvents))
		for i, ev := range dbEvents {
			apiEvents[i] = eventToAPI(ev)
		}
	} else if agentID != "" {
		dbEvents, e := s.store.ListEventsByAgent(ctx, agentID, limit)
		err = e
		apiEvents = make([]map[string]interface{}, len(dbEvents))
		for i, ev := range dbEvents {
			apiEvents[i] = eventToAPI(ev)
		}
	} else {
		dbEvents, e := s.store.ListEvents(ctx, limit)
		err = e
		apiEvents = make([]map[string]interface{}, len(dbEvents))
		for i, ev := range dbEvents {
			apiEvents[i] = eventToAPI(ev)
		}
	}
	
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": apiEvents,
		"meta": map[string]interface{}{
			"total": len(apiEvents),
			"limit": limit,
		},
	})
}

func (s *Server) createEvent(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "Create event not implemented"})
}

// Settings handlers
func (s *Server) getSettings(c echo.Context) error {
	ctx := c.Request().Context()
	
	settings, err := s.store.GetSettings(ctx)
	if err != nil {
		// If no settings exist, return default settings
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"id":                        "default",
				"openclaw_gateway_url":      "ws://127.0.0.1:18789",
				"openclaw_gateway_token":    "",
				"default_approach":          "gsd",
				"default_model":             "anthropic/claude-sonnet-4-5",
				"max_parallel_executions":   3,
				"gsd_depth":                 "standard",
				"gsd_mode":                  "interactive",
				"gsd_research_enabled":      true,
				"gsd_plan_check_enabled":    true,
				"gsd_verifier_enabled":      true,
				"ralph_max_iterations":      10,
				"ralph_auto_commit":         true,
				"theme":                     "dark",
				"default_project_directory": "",
			},
		})
	}
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": settingsToAPI(settings),
	})
}

func (s *Server) updateSettings(c echo.Context) error {
	return c.JSON(http.StatusNotImplemented, map[string]string{"error": "Update settings not implemented yet"})
}

func (s *Server) testConnection(c echo.Context) error {
	// For now, just return success - actual implementation would test OpenClaw connection
	return c.JSON(http.StatusOK, map[string]interface{}{
		"connected": true,
		"message":   "Connection test not fully implemented",
	})
}

// Helper functions
func parseLimit(s string) (int64, error) {
	var limit int64
	_, err := fmt.Sscanf(s, "%d", &limit)
	if err != nil {
		return 50, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	return limit, nil
}

func eventToAPI(e db.Event) map[string]interface{} {
	result := map[string]interface{}{
		"id":         e.ID,
		"type":       e.Type,
		"message":    e.Message,
		"created_at": e.CreatedAt.Time,
	}
	
	// Handle nullable fields
	if e.TaskID.Valid {
		result["task_id"] = e.TaskID.String
	}
	if e.AgentID.Valid {
		result["agent_id"] = e.AgentID.String
	}
	if e.Details.Valid {
		result["details"] = e.Details.String
	}
	
	return result
}

func settingsToAPI(s db.Setting) map[string]interface{} {
	result := map[string]interface{}{
		"id": s.ID,
	}
	
	// Handle nullable fields with defaults
	if s.OpenclawGatewayUrl.Valid {
		result["openclaw_gateway_url"] = s.OpenclawGatewayUrl.String
	} else {
		result["openclaw_gateway_url"] = "ws://127.0.0.1:18789"
	}
	
	if s.OpenclawGatewayToken.Valid {
		result["openclaw_gateway_token"] = s.OpenclawGatewayToken.String
	} else {
		result["openclaw_gateway_token"] = ""
	}
	
	// Note: No default_approach - all tasks use GSD for planning and Ralph for execution
	
	if s.DefaultModel.Valid {
		result["default_model"] = s.DefaultModel.String
	} else {
		result["default_model"] = "anthropic/claude-sonnet-4-5"
	}
	
	if s.MaxParallelExecutions.Valid {
		result["max_parallel_executions"] = s.MaxParallelExecutions.Int64
	} else {
		result["max_parallel_executions"] = 3
	}
	
	if s.GsdDepth.Valid {
		result["gsd_depth"] = s.GsdDepth.String
	} else {
		result["gsd_depth"] = "standard"
	}
	
	if s.GsdMode.Valid {
		result["gsd_mode"] = s.GsdMode.String
	} else {
		result["gsd_mode"] = "interactive"
	}
	
	if s.GsdResearchEnabled.Valid {
		result["gsd_research_enabled"] = s.GsdResearchEnabled.Int64 == 1
	} else {
		result["gsd_research_enabled"] = true
	}
	
	if s.GsdPlanCheckEnabled.Valid {
		result["gsd_plan_check_enabled"] = s.GsdPlanCheckEnabled.Int64 == 1
	} else {
		result["gsd_plan_check_enabled"] = true
	}
	
	if s.GsdVerifierEnabled.Valid {
		result["gsd_verifier_enabled"] = s.GsdVerifierEnabled.Int64 == 1
	} else {
		result["gsd_verifier_enabled"] = true
	}
	
	if s.RalphMaxIterations.Valid {
		result["ralph_max_iterations"] = s.RalphMaxIterations.Int64
	} else {
		result["ralph_max_iterations"] = 10
	}
	
	if s.RalphAutoCommit.Valid {
		result["ralph_auto_commit"] = s.RalphAutoCommit.Int64 == 1
	} else {
		result["ralph_auto_commit"] = true
	}
	
	if s.Theme.Valid {
		result["theme"] = s.Theme.String
	} else {
		result["theme"] = "dark"
	}

	if s.DefaultProjectDirectory.Valid {
		result["default_project_directory"] = s.DefaultProjectDirectory.String
	} else {
		result["default_project_directory"] = ""
	}
	
	return result
}

// Models handler - returns configured models from OpenClaw
func (s *Server) listModels(c echo.Context) error {
	configReader := openclaw.NewConfigReader("")
	models, err := configReader.ReadModels()
	if err != nil {
		// Return empty list on error
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": []interface{}{},
			"error": err.Error(),
		})
	}
	
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": models,
	})
}
