package networkmode

import (
	"context"
	"fmt"
	"log/slog"
)

// Router routes traffic based on the current mode
type Router struct {
	config  *Config
	logger  *slog.Logger
}

// NewRouter creates a new router
func NewRouter(config *Config, logger *slog.Logger) *Router {
	if logger == nil {
		logger = slog.Default()
	}

	return &Router{
		config: config,
		logger: logger,
	}
}

// RouteRequest routes a request based on the current mode
func (r *Router) RouteRequest(ctx context.Context, req *Request, decision *Decision) (*Response, error) {
	switch r.config.Mode {
	case ModeFull:
		return r.routeFullMode(ctx, req)
	case ModeHalf:
		return r.routeHalfMode(ctx, req, decision)
	case ModeTransparent:
		// In transparent mode the router returns a pass-through response.
		// The actual traffic observation is handled by TransparentModeHandler;
		// the router simply acknowledges the pass-through.
		return r.routeTransparentMode(ctx, req)
	default:
		// If mode is invalid or unknown, fail safe to Full Mode
		r.logger.WarnContext(ctx, "Invalid mode, falling back to Full Mode",
			"mode", r.config.Mode)
		return r.routeFullMode(ctx, req)
	}
}

// routeFullMode routes request in Full Mode (all simulated)
func (r *Router) routeFullMode(ctx context.Context, req *Request) (*Response, error) {
	r.logger.InfoContext(ctx, "Routing in Full Mode",
		"req_id", req.ID,
		"protocol", req.Protocol,
		"domain", req.Domain)

	// Route to appropriate simulation service based on protocol
	serviceAddr, err := r.getServiceAddress(req.Protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to get service address: %w", err)
	}

	r.logger.DebugContext(ctx, "Routing to simulation service",
		"service", serviceAddr,
		"protocol", req.Protocol)

	// In Full Mode, we return a simulated response
	// The actual communication with simulation services (INetSim, FakeNet-NG)
	// would be implemented here
	resp := &Response{
		ID:        req.ID,
		Timestamp: req.Timestamp,
		StatusCode: 200,
		Headers:   make(map[string]string),
		Source:    "simulated",
		Metadata: map[string]interface{}{
			"service":     serviceAddr,
			"mode":        "full",
			"isolated":    true,
		},
	}

	// Add simulation headers
	resp.Headers["X-Pack-A-Mal-Mode"] = "full"
	resp.Headers["X-Pack-A-Mal-Source"] = "simulated"
	resp.Headers["X-Pack-A-Mal-Service"] = serviceAddr

	// Generate simulated response based on protocol
	resp.Body = r.generateSimulatedResponse(req)
	resp.ContentLength = int64(len(resp.Body))

	return resp, nil
}

// routeTransparentMode returns a pass-through response for Transparent Mode.
// No traffic is altered; the response signals that the packet should be allowed.
// Inspired by siemens/sparring transparent mode: NF_STOP (allow packet through).
func (r *Router) routeTransparentMode(ctx context.Context, req *Request) (*Response, error) {
	r.logger.DebugContext(ctx, "Routing in Transparent Mode (pass-through)",
		"req_id", req.ID,
		"protocol", req.Protocol,
		"domain", req.Domain)

	resp := &Response{
		ID:        req.ID,
		Timestamp: req.Timestamp,
		Headers:   make(map[string]string),
		Source:    "transparent_passthrough",
		Metadata: map[string]interface{}{
			"mode":   "transparent",
			"action": "passthrough",
		},
	}

	resp.Headers["X-Pack-A-Mal-Mode"] = "transparent"
	resp.Headers["X-Pack-A-Mal-Source"] = "passthrough"

	return resp, nil
}

// routeHalfMode routes request in Half Mode based on decision
func (r *Router) routeHalfMode(ctx context.Context, req *Request, decision *Decision) (*Response, error) {
	if decision == nil {
		return nil, ErrNoDecision
	}

	r.logger.InfoContext(ctx, "Routing in Half Mode",
		"req_id", req.ID,
		"domain", req.Domain,
		"action", decision.Action,
		"rule", decision.RuleName)

	switch decision.Action {
	case ActionForward:
		return r.forwardToRealDestination(ctx, req)

	case ActionBlock:
		return r.blockRequest(ctx, req, decision.Reason)

	case ActionModify:
		// Forward with modifications will be handled by the caller
		// Here we just forward to real destination
		return r.forwardToRealDestination(ctx, req)

	case ActionSimulate:
		// Use Full Mode path for simulation
		return r.routeFullMode(ctx, req)

	default:
		// Fail safe to simulation
		r.logger.WarnContext(ctx, "Unknown action, falling back to simulation",
			"action", decision.Action)
		return r.routeFullMode(ctx, req)
	}
}

// forwardToRealDestination forwards request to real internet destination
func (r *Router) forwardToRealDestination(ctx context.Context, req *Request) (*Response, error) {
	r.logger.WarnContext(ctx, "Forwarding to real destination",
		"req_id", req.ID,
		"domain", req.Domain,
		"ip", req.IP,
		"port", req.Port)

	// TODO: Implement actual HTTP/HTTPS client to forward request
	// This should use proper HTTP client with timeout configured
	// For now, return a placeholder response

	resp := &Response{
		ID:         req.ID,
		Timestamp:  req.Timestamp,
		StatusCode: 200,
		Headers:    make(map[string]string),
		Source:     "real",
		Metadata: map[string]interface{}{
			"mode":        "half",
			"action":      "forward",
			"destination": fmt.Sprintf("%s:%d", req.Domain, req.Port),
		},
	}

	resp.Headers["X-Pack-A-Mal-Mode"] = "half"
	resp.Headers["X-Pack-A-Mal-Source"] = "real"
	resp.Body = []byte(fmt.Sprintf("Forwarded to %s", req.Domain))
	resp.ContentLength = int64(len(resp.Body))

	return resp, nil
}

// blockRequest blocks a request and returns an error response
func (r *Router) blockRequest(ctx context.Context, req *Request, reason string) (*Response, error) {
	r.logger.WarnContext(ctx, "Request blocked",
		"req_id", req.ID,
		"domain", req.Domain,
		"reason", reason)

	resp := &Response{
		ID:         req.ID,
		Timestamp:  req.Timestamp,
		StatusCode: 403, // Forbidden
		Headers:    make(map[string]string),
		Source:     "blocked",
		Metadata: map[string]interface{}{
			"mode":   "half",
			"action": "block",
			"reason": reason,
		},
	}

	resp.Headers["X-Pack-A-Mal-Mode"] = "half"
	resp.Headers["X-Pack-A-Mal-Source"] = "blocked"
	resp.Headers["X-Pack-A-Mal-Reason"] = reason
	
	resp.Body = []byte(fmt.Sprintf("Request blocked: %s", reason))
	resp.ContentLength = int64(len(resp.Body))

	return resp, nil
}

// getServiceAddress returns the service address for a protocol
func (r *Router) getServiceAddress(protocol string) (string, error) {
	if r.config.FullMode == nil || r.config.FullMode.Services == nil {
		return "", fmt.Errorf("full mode services not configured")
	}

	services := r.config.FullMode.Services

	switch protocol {
	case string(ProtocolDNS):
		if services.DNSAddress != "" {
			return services.DNSAddress, nil
		}
		return services.DNS, nil

	case string(ProtocolHTTP):
		if services.HTTPAddress != "" {
			return services.HTTPAddress, nil
		}
		return services.HTTP, nil

	case string(ProtocolHTTPS):
		if services.HTTPSAddress != "" {
			return services.HTTPSAddress, nil
		}
		return services.HTTPS, nil

	case string(ProtocolSMTP):
		if services.SMTPAddress != "" {
			return services.SMTPAddress, nil
		}
		return services.SMTP, nil

	case string(ProtocolFTP):
		if services.FTPAddress != "" {
			return services.FTPAddress, nil
		}
		return services.FTP, nil

	default:
		// Default to HTTP for unknown protocols
		return services.HTTPAddress, nil
	}
}

// generateSimulatedResponse generates a simulated response based on request
func (r *Router) generateSimulatedResponse(req *Request) []byte {
	switch req.Protocol {
	case string(ProtocolDNS):
		return []byte(fmt.Sprintf("DNS response for %s: 127.0.0.1", req.Domain))

	case string(ProtocolHTTP), string(ProtocolHTTPS):
		return []byte(fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>Simulated Response</title></head>
<body>
<h1>Pack-A-Mal Simulated Response</h1>
<p>Domain: %s</p>
<p>Path: %s</p>
<p>This is a simulated response for security analysis.</p>
</body>
</html>`, req.Domain, req.Path))

	case string(ProtocolSMTP):
		return []byte("250 OK - Simulated SMTP response")

	case string(ProtocolFTP):
		return []byte("230 User logged in - Simulated FTP response")

	default:
		return []byte(fmt.Sprintf("Simulated response for %s protocol", req.Protocol))
	}
}
