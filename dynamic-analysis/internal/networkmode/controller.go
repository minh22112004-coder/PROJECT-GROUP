package networkmode

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Controller is the main Network Mode Controller
type Controller struct {
	// Mode - Current network mode
	mode Mode

	// Config - Controller configuration
	config *Config

	// Components
	interceptor        *TrafficInterceptor
	decisionEngine     *DecisionEngine
	modifier           *TrafficModifier
	router             *Router
	logger             *NetworkLogger
	transparentHandler *TransparentModeHandler

	// Structured logger
	slogger *slog.Logger

	// Stats
	stats      *Stats
	statsMutex sync.RWMutex

	// Shutdown
	shutdown chan struct{}
	wg       sync.WaitGroup
}

// Stats holds controller statistics
type Stats struct {
	TotalRequests      int64
	ForwardedRequests  int64
	BlockedRequests    int64
	ModifiedRequests   int64
	SimulatedRequests  int64
	Errors             int64
	LastRequestTime    time.Time
	StartTime          time.Time
}

// NewController creates a new Network Mode Controller
func NewController(config *Config, logger *slog.Logger) (*Controller, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Use default logger if not provided
	if logger == nil {
		logger = slog.Default()
	}

	// Create network logger
	networkLogger, err := NewLogger(config.Logging, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Create components
	interceptor := NewInterceptor(config, logger)
	router := NewRouter(config, logger)
	
	var decisionEngine *DecisionEngine
	var modifier *TrafficModifier
	
	// Only create Half Mode components if in Half Mode
	if config.Mode == ModeHalf {
		if config.HalfMode == nil {
			return nil, fmt.Errorf("half mode config is required for Half Mode")
		}
		
		decisionEngine = NewDecisionEngine(config.HalfMode, logger)
		modifier = NewModifier(config.HalfMode.TrafficModifier, logger)
		
		// Load default rules
		if err := decisionEngine.AddRules(DefaultRules()); err != nil {
			return nil, fmt.Errorf("failed to add default rules: %w", err)
		}
	}

	// Create Transparent Mode handler if in Transparent Mode
	var transparentHandler *TransparentModeHandler
	if config.Mode == ModeTransparent {
		if config.TransparentMode == nil {
			return nil, fmt.Errorf("transparent mode config is required for Transparent Mode")
		}
		var err error
		transparentHandler, err = NewTransparentModeHandler(config.TransparentMode, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create transparent mode handler: %w", err)
		}
	}

	controller := &Controller{
		mode:               config.Mode,
		config:             config,
		interceptor:        interceptor,
		decisionEngine:     decisionEngine,
		modifier:           modifier,
		router:             router,
		logger:             networkLogger,
		slogger:            logger,
		transparentHandler: transparentHandler,
		shutdown:           make(chan struct{}),
		stats: &Stats{
			StartTime: time.Now(),
		},
	}

	logger.Info("Network Mode Controller initialized",
		"mode", config.Mode,
		"full_mode_isolation", config.Mode == ModeFull,
		"transparent_mode", config.Mode == ModeTransparent)

	return controller, nil
}

// HandleRequest handles a network request
// This is the main entry point for processing requests
func (c *Controller) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
	startTime := time.Now()
	
	// Update stats
	c.incrementStat("total")
	
	// Log request
	if err := c.logger.LogRequest(ctx, req); err != nil {
		c.slogger.WarnContext(ctx, "Failed to log request", "error", err)
	}

	// Safety check: If any error occurs, fall back to Full Mode
	defer func() {
		if r := recover(); r != nil {
			c.slogger.ErrorContext(ctx, "Panic in HandleRequest, falling back to Full Mode",
				"panic", r,
				"req_id", req.ID)
			c.incrementStat("errors")
		}
	}()

	var resp *Response
	var decision *Decision
	var err error

	// Process based on mode
	switch c.mode {
	case ModeFull:
		resp, err = c.handleFullMode(ctx, req)
		decision = &Decision{
			Action:     ActionSimulate,
			Reason:     "Full Mode - all traffic simulated",
			RuleName:   "full_mode",
			Confidence: 1.0,
		}

	case ModeHalf:
		resp, decision, err = c.handleHalfMode(ctx, req)

	case ModeTransparent:
		resp, decision, err = c.handleTransparentMode(ctx, req)

	default:
		// Invalid mode - fail safe to Full Mode
		c.slogger.ErrorContext(ctx, "Invalid mode, failing safe to Full Mode",
			"mode", c.mode,
			"req_id", req.ID)
		resp, err = c.handleFullMode(ctx, req)
		decision = &Decision{
			Action:     ActionSimulate,
			Reason:     "Invalid mode - failed safe to Full Mode",
			RuleName:   "failsafe",
			Confidence: 1.0,
		}
	}

	if err != nil {
		c.incrementStat("errors")
		c.logger.LogError(ctx, req, err)
		return nil, fmt.Errorf("failed to handle request: %w", err)
	}

	// Add decision to response
	if resp != nil {
		resp.Decision = decision
	}

	// Log response
	if err := c.logger.LogResponse(ctx, resp); err != nil {
		c.slogger.WarnContext(ctx, "Failed to log response", "error", err)
	}

	// Log traffic
	duration := time.Since(startTime)
	trafficLog := &TrafficLog{
		Timestamp: startTime,
		Request:   req,
		Response:  resp,
		Decision:  decision,
		Action:    decision.Action,
		Duration:  duration,
	}
	if err := c.logger.LogTraffic(ctx, trafficLog); err != nil {
		c.slogger.WarnContext(ctx, "Failed to log traffic", "error", err)
	}

	// Update last request time
	c.statsMutex.Lock()
	c.stats.LastRequestTime = time.Now()
	c.statsMutex.Unlock()

	return resp, nil
}

// handleFullMode handles request in Full Mode
func (c *Controller) handleFullMode(ctx context.Context, req *Request) (*Response, error) {
	c.slogger.InfoContext(ctx, "Handling in Full Mode",
		"req_id", req.ID,
		"domain", req.Domain,
		"protocol", req.Protocol)

	// In Full Mode, all traffic is simulated
	c.incrementStat("simulated")

	// Route to simulation service
	resp, err := c.router.RouteRequest(ctx, req, &Decision{
		Action: ActionSimulate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to route request: %w", err)
	}

	return resp, nil
}

// handleTransparentMode handles a request in Transparent Mode.
// Key principle (from siemens/sparring): DO NOT modify traffic, only log and observe.
// All connections pass through unmodified; payloads are extracted for analysis.
func (c *Controller) handleTransparentMode(ctx context.Context, req *Request) (*Response, *Decision, error) {
	c.slogger.InfoContext(ctx, "Handling in Transparent Mode (observe only)",
		"req_id", req.ID,
		"domain", req.Domain,
		"protocol", req.Protocol)

	// Ensure handler is available
	if c.transparentHandler == nil {
		c.slogger.ErrorContext(ctx, "Transparent handler not initialized, failing safe to Full Mode",
			"req_id", req.ID)
		resp, err := c.handleFullMode(ctx, req)
		decision := &Decision{
			Action:     ActionSimulate,
			Reason:     "Transparent handler not ready - failed safe to Full Mode",
			RuleName:   "transparent_failsafe",
			Confidence: 0.0,
		}
		return resp, decision, err
	}

	// Let the transparent handler observe (no modification)
	resp, err := c.transparentHandler.HandleRequest(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("transparent mode handler error: %w", err)
	}

	// Decision for transparent mode: always "forward" (pass through)
	// Traffic is NOT blocked or modified - this is the core invariant
	decision := &Decision{
		Action:     ActionForward,
		Reason:     "Transparent Mode - traffic observed and passed through unmodified",
		RuleName:   "transparent_passthrough",
		Confidence: 1.0,
	}

	c.incrementStat("forwarded")

	return resp, decision, nil
}

// handleHalfMode handles request in Half Mode
func (c *Controller) handleHalfMode(ctx context.Context, req *Request) (*Response, *Decision, error) {
	c.slogger.InfoContext(ctx, "Handling in Half Mode",
		"req_id", req.ID,
		"domain", req.Domain,
		"protocol", req.Protocol)

	// Make decision
	decision, err := c.decisionEngine.Decide(ctx, req)
	if err != nil {
		c.slogger.ErrorContext(ctx, "Decision engine failed, falling back to simulation",
			"error", err,
			"req_id", req.ID)
		
		// Fail safe to simulation
		decision = &Decision{
			Action:     ActionSimulate,
			Reason:     fmt.Sprintf("Decision failed: %v", err),
			RuleName:   "error_fallback",
			Confidence: 0.0,
		}
	}

	// Log decision
	if err := c.logger.LogDecision(ctx, req, decision); err != nil {
		c.slogger.WarnContext(ctx, "Failed to log decision", "error", err)
	}

	// Apply modifications if needed
	modifiedReq := req
	if decision.Action == ActionModify && decision.Modifier != nil {
		modifiedReq, err = c.modifier.ModifyRequest(ctx, req, decision.Modifier)
		if err != nil {
			c.slogger.ErrorContext(ctx, "Failed to modify request",
				"error", err,
				"req_id", req.ID)
			// Continue with original request
			modifiedReq = req
		} else {
			c.logger.LogModification(ctx, req, "request_modified")
			c.incrementStat("modified")
		}
	}

	// Route based on decision
	resp, err := c.router.RouteRequest(ctx, modifiedReq, decision)
	if err != nil {
		return nil, decision, fmt.Errorf("failed to route request: %w", err)
	}

	// Update stats based on action
	switch decision.Action {
	case ActionForward:
		c.incrementStat("forwarded")
	case ActionBlock:
		c.incrementStat("blocked")
	case ActionModify:
		// Already incremented above
	case ActionSimulate:
		c.incrementStat("simulated")
	}

	// Apply response modifications if needed
	if decision.Action == ActionModify && decision.Modifier != nil {
		modifiedResp, err := c.modifier.ModifyResponse(ctx, resp, req, decision.Modifier)
		if err != nil {
			c.slogger.ErrorContext(ctx, "Failed to modify response",
				"error", err,
				"req_id", req.ID)
			// Continue with original response
		} else {
			c.logger.LogModification(ctx, req, "response_modified")
			resp = modifiedResp
		}
	}

	return resp, decision, nil
}

// GetMode returns the current mode
func (c *Controller) GetMode() Mode {
	return c.mode
}

// SwitchMode switches the network mode
func (c *Controller) SwitchMode(ctx context.Context, newMode Mode) error {
	if !newMode.IsValid() {
		return ErrInvalidMode
	}

	// Validate that Half Mode is enabled if switching to it
	if newMode == ModeHalf {
		if c.config.HalfMode == nil || !c.config.HalfMode.Enabled {
			return ErrHalfModeNotEnabled
		}
	}

	// Validate that Transparent Mode is enabled if switching to it
	if newMode == ModeTransparent {
		if c.config.TransparentMode == nil || !c.config.TransparentMode.Enabled {
			return ErrTransparentModeNotEnabled
		}
	}

	c.slogger.InfoContext(ctx, "Switching network mode",
		"from", c.mode,
		"to", newMode)

	c.mode = newMode
	c.config.Mode = newMode

	// Reinitialize components if needed
	if newMode == ModeHalf && c.decisionEngine == nil {
		c.decisionEngine = NewDecisionEngine(c.config.HalfMode, c.slogger)
		c.modifier = NewModifier(c.config.HalfMode.TrafficModifier, c.slogger)
		if err := c.decisionEngine.AddRules(DefaultRules()); err != nil {
			return fmt.Errorf("failed to add default rules: %w", err)
		}
	}

	// Initialize Transparent Mode handler if switching to it
	if newMode == ModeTransparent && c.transparentHandler == nil {
		handler, err := NewTransparentModeHandler(c.config.TransparentMode, c.slogger)
		if err != nil {
			return fmt.Errorf("failed to initialize transparent mode handler: %w", err)
		}
		c.transparentHandler = handler
	}

	return nil
}

// GetStats returns controller statistics
func (c *Controller) GetStats() *Stats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()

	// Return a copy to avoid race conditions
	statsCopy := *c.stats
	return &statsCopy
}

// incrementStat increments a statistic
func (c *Controller) incrementStat(stat string) {
	c.statsMutex.Lock()
	defer c.statsMutex.Unlock()

	switch stat {
	case "total":
		c.stats.TotalRequests++
	case "forwarded":
		c.stats.ForwardedRequests++
	case "blocked":
		c.stats.BlockedRequests++
	case "modified":
		c.stats.ModifiedRequests++
	case "simulated":
		c.stats.SimulatedRequests++
	case "errors":
		c.stats.Errors++
	}
}

// AddDecisionRule adds a decision rule (Half Mode only)
func (c *Controller) AddDecisionRule(ctx context.Context, rule DecisionRule) error {
	if c.mode != ModeHalf {
		return fmt.Errorf("decision rules can only be added in Half Mode")
	}

	if c.decisionEngine == nil {
		return fmt.Errorf("decision engine not initialized")
	}

	return c.decisionEngine.AddRule(rule)
}

// GetDecisionRules returns all decision rules
func (c *Controller) GetDecisionRules() []DecisionRule {
	if c.decisionEngine == nil {
		return []DecisionRule{}
	}
	return c.decisionEngine.GetRules()
}

// ClearDecisionCache clears the decision cache
func (c *Controller) ClearDecisionCache() {
	if c.decisionEngine != nil {
		c.decisionEngine.ClearCache()
	}
}

// GetTransparentStats returns transparent mode statistics (only valid in Transparent Mode)
func (c *Controller) GetTransparentStats() (map[string]interface{}, error) {
	if c.mode != ModeTransparent || c.transparentHandler == nil {
		return nil, fmt.Errorf("transparent stats only available in Transparent Mode")
	}
	return c.transparentHandler.GetConnectionStats(), nil
}

// GetTransparentSummary returns a human-readable traffic summary for Transparent Mode
func (c *Controller) GetTransparentSummary() (string, error) {
	if c.mode != ModeTransparent || c.transparentHandler == nil {
		return "", fmt.Errorf("transparent summary only available in Transparent Mode")
	}
	return c.transparentHandler.PrintSummary(), nil
}

// Close gracefully shuts down the controller
func (c *Controller) Close() error {
	c.slogger.Info("Shutting down Network Mode Controller")

	// Signal shutdown
	close(c.shutdown)

	// Wait for goroutines to finish
	c.wg.Wait()

	// Close transparent handler and flush logs
	if c.transparentHandler != nil {
		if err := c.transparentHandler.Close(); err != nil {
			c.slogger.Warn("Error closing transparent handler", "error", err)
		}
	}

	// Close logger
	if c.logger != nil {
		if err := c.logger.Close(); err != nil {
			c.slogger.Warn("Error closing logger", "error", err)
		}
	}

	c.slogger.Info("Network Mode Controller shut down complete")
	return nil
}

// Health checks the health of the controller
func (c *Controller) Health(ctx context.Context) error {
	// Check if mode is valid
	if !c.mode.IsValid() {
		return fmt.Errorf("invalid mode: %s", c.mode)
	}

	// Check if config is valid
	if err := c.config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Check Half Mode specific health
	if c.mode == ModeHalf {
		if c.decisionEngine == nil {
			return fmt.Errorf("decision engine not initialized in Half Mode")
		}
		if c.modifier == nil {
			return fmt.Errorf("modifier not initialized in Half Mode")
		}
	}

	// Check Transparent Mode specific health
	if c.mode == ModeTransparent {
		if c.transparentHandler == nil {
			return fmt.Errorf("transparent handler not initialized in Transparent Mode")
		}
	}

	return nil
}
