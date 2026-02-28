package networkmode

// transparent.go - Transparent Mode Handler
//
// Inspired by siemens/sparring transparent mode:
// https://github.com/siemens/sparring
//
// In TRANSPARENT mode: sparring will NOT alter any transmitted data and
// only log connections and try to extract interesting data for supported protocols.
//
// This mode is ideal for passive monitoring and forensic analysis without
// affecting the behavior of the sample under analysis.

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ConnectionInfo holds information about a monitored connection
// Mirrors sparring's connection tracking in transparent mode
type ConnectionInfo struct {
	// ID - Unique connection identifier
	ID string `json:"id"`

	// Protocol - Transport protocol (TCP, UDP, ICMP)
	Protocol string `json:"protocol"`

	// AppProtocol - Identified application protocol (HTTP, DNS, SMTP, FTP)
	AppProtocol string `json:"app_protocol,omitempty"`

	// SrcIP - Source IP address
	SrcIP string `json:"src_ip"`

	// SrcPort - Source port
	SrcPort int `json:"src_port"`

	// DstIP - Destination IP address
	DstIP string `json:"dst_ip"`

	// DstPort - Destination port
	DstPort int `json:"dst_port"`

	// Domain - Resolved destination domain (if available)
	Domain string `json:"domain,omitempty"`

	// StartTime - Connection start time
	StartTime time.Time `json:"start_time"`

	// EndTime - Connection end time
	EndTime *time.Time `json:"end_time,omitempty"`

	// BytesSent - Bytes sent by source
	BytesSent int64 `json:"bytes_sent"`

	// BytesReceived - Bytes received by source
	BytesReceived int64 `json:"bytes_received"`
}

// ExtractedPayload holds extracted protocol-specific payload data
// Mirrors sparring's protocol extraction in transparent mode
type ExtractedPayload struct {
	// ConnectionID - Parent connection ID
	ConnectionID string `json:"connection_id"`

	// Timestamp - Extraction timestamp
	Timestamp time.Time `json:"timestamp"`

	// Protocol - Application protocol
	Protocol string `json:"protocol"`

	// Direction - "outgoing" (from sample) or "incoming" (to sample)
	Direction string `json:"direction"`

	// RawData - Raw payload bytes (truncated to MaxPayloadSize)
	RawData []byte `json:"raw_data,omitempty"`

	// ParsedData - Parsed protocol-specific data
	ParsedData map[string]interface{} `json:"parsed_data,omitempty"`

	// Size - Original payload size
	Size int64 `json:"size"`

	// Truncated - Whether the captured payload was truncated
	Truncated bool `json:"truncated"`
}

// TransparentModeHandler handles network traffic in transparent mode
// Key principle: DO NOT MODIFY any traffic, only observe and log
type TransparentModeHandler struct {
	config *TransparentModeConfig
	logger *slog.Logger

	// Connection tracking (mirrors sparring's connection dict)
	connections   map[string]*ConnectionInfo
	connectionsMu sync.RWMutex

	// Log writers
	connLogFile    *os.File
	payloadLogFile *os.File
	connWriter     *bufio.Writer
	payloadWriter  *bufio.Writer
	writerMu       sync.Mutex

	// Statistics
	stats TransparentStats
}

// TransparentStats holds transparent mode statistics
type TransparentStats struct {
	TotalConnections    int64
	TCPConnections      int64
	UDPConnections      int64
	ICMPPackets         int64
	TotalBytesObserved  int64
	ProtocolBreakdown   sync.Map // map[string]int64
	ExtractedPayloads   int64
	UnknownProtocols    int64
}

// NewTransparentModeHandler creates a new transparent mode handler
func NewTransparentModeHandler(config *TransparentModeConfig, logger *slog.Logger) (*TransparentModeHandler, error) {
	if config == nil {
		return nil, fmt.Errorf("transparent mode config is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	h := &TransparentModeHandler{
		config:      config,
		logger:      logger,
		connections: make(map[string]*ConnectionInfo),
	}

	// Open connection log file
	if config.LogConnections && config.ConnectionLogFile != "" {
		f, err := openLogFile(config.ConnectionLogFile)
		if err != nil {
			logger.Warn("Failed to open connection log file, logging to stderr",
				"path", config.ConnectionLogFile,
				"error", err)
		} else {
			h.connLogFile = f
			h.connWriter = bufio.NewWriter(f)
		}
	}

	// Open payload log file
	if config.ExtractPayloads && config.PayloadLogFile != "" {
		f, err := openLogFile(config.PayloadLogFile)
		if err != nil {
			logger.Warn("Failed to open payload log file, logging to stderr",
				"path", config.PayloadLogFile,
				"error", err)
		} else {
			h.payloadLogFile = f
			h.payloadWriter = bufio.NewWriter(f)
		}
	}

	logger.Info("Transparent Mode handler initialized",
		"extract_payloads", config.ExtractPayloads,
		"log_connections", config.LogConnections,
		"log_icmp", config.LogICMP,
		"supported_protocols", config.SupportedProtocols)

	return h, nil
}

// HandleRequest processes a request in transparent mode.
// CRITICAL: This method NEVER modifies the request or decides to block/forward.
// It only observes, logs, and passes through.
// This is the core principle from sparring's transparent mode.
func (h *TransparentModeHandler) HandleRequest(ctx context.Context, req *Request) (*Response, error) {
	// Track connection
	conn := h.trackConnection(req)

	// Log connection event
	if h.config.LogConnections {
		h.writeConnectionLog(conn, "observed")
	}

	// Extract payload data from known protocols
	if h.config.ExtractPayloads {
		h.extractAndLogPayload(ctx, req, conn)
	}

	// Log ICMP if configured
	if h.config.LogICMP && req.Protocol == "ICMP" {
		atomic.AddInt64(&h.stats.ICMPPackets, 1)
		h.logger.InfoContext(ctx, "ICMP observed",
			"src_ip", req.SourceIP,
			"dst_ip", req.IP,
			"conn_id", conn.ID)
	}

	// Update byte counters
	atomic.AddInt64(&h.stats.TotalBytesObserved, req.ContentLength)

	h.logger.DebugContext(ctx, "Transparent mode: traffic observed (not modified)",
		"req_id", req.ID,
		"protocol", req.Protocol,
		"src", fmt.Sprintf("%s:%d", req.SourceIP, req.SourcePort),
		"dst", fmt.Sprintf("%s:%d", req.IP, req.Port),
		"domain", req.Domain)

	// Return a pass-through response:
	// In transparent mode we return a special response that tells
	// the caller to allow the traffic to pass unmodified
	return &Response{
		ID:        req.ID,
		Timestamp: time.Now(),
		Source:    "transparent_passthrough",
		Metadata: map[string]interface{}{
			"mode":         "transparent",
			"connection_id": conn.ID,
			"action":       "passthrough",
			"note":         "Traffic observed only - no modification applied",
		},
	}, nil
}

// trackConnection tracks or updates a connection in the connection table
// Mirrors sparring's handling of the TCP/UDP connection dictionary
func (h *TransparentModeHandler) trackConnection(req *Request) *ConnectionInfo {
	connKey := fmt.Sprintf("%s:%d->%s:%d/%s",
		req.SourceIP, req.SourcePort,
		req.IP, req.Port,
		req.Protocol)

	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()

	if conn, exists := h.connections[connKey]; exists {
		// Update existing connection
		conn.BytesSent += req.ContentLength
		if req.Domain != "" {
			conn.Domain = req.Domain
		}
		if req.Protocol != "" && conn.AppProtocol == "" {
			appProto := h.identifyAppProtocol(req)
			if appProto != "" {
				conn.AppProtocol = appProto
			}
		}
		return conn
	}

	// New connection
	conn := &ConnectionInfo{
		ID:        generateRequestID(),
		Protocol:  req.Protocol,
		SrcIP:     req.SourceIP,
		SrcPort:   req.SourcePort,
		DstIP:     req.IP,
		DstPort:   req.Port,
		Domain:    req.Domain,
		StartTime: time.Now(),
		BytesSent: req.ContentLength,
	}

	// Identify application-layer protocol
	conn.AppProtocol = h.identifyAppProtocol(req)

	h.connections[connKey] = conn

	// Update stats
	atomic.AddInt64(&h.stats.TotalConnections, 1)
	switch strings.ToUpper(req.Protocol) {
	case "TCP":
		atomic.AddInt64(&h.stats.TCPConnections, 1)
	case "UDP":
		atomic.AddInt64(&h.stats.UDPConnections, 1)
	}
	if conn.AppProtocol != "" {
		v, _ := h.stats.ProtocolBreakdown.LoadOrStore(conn.AppProtocol, new(int64))
		atomic.AddInt64(v.(*int64), 1)
	} else {
		atomic.AddInt64(&h.stats.UnknownProtocols, 1)
	}

	return conn
}

// identifyAppProtocol identifies the application-layer protocol from a request
// Mirrors sparring's protocol classification (classify())
func (h *TransparentModeHandler) identifyAppProtocol(req *Request) string {
	// Use already-identified protocol if available
	switch strings.ToUpper(req.Protocol) {
	case "HTTP", "HTTPS", "DNS", "SMTP", "FTP":
		return strings.ToUpper(req.Protocol)
	}

	// Port-based identification (mirrors sparring's application modules)
	portProtoMap := map[int]string{
		80:  "HTTP",
		443: "HTTPS",
		53:  "DNS",
		25:  "SMTP",
		587: "SMTP",
		465: "SMTPS",
		21:  "FTP",
		22:  "SSH",
		110: "POP3",
		143: "IMAP",
	}

	if proto, ok := portProtoMap[req.Port]; ok {
		return proto
	}

	// Payload-based identification
	if len(req.Body) >= 4 {
		prefix := string(req.Body[:4])
		switch {
		case strings.HasPrefix(string(req.Body), "GET "),
			strings.HasPrefix(string(req.Body), "POST"),
			strings.HasPrefix(string(req.Body), "PUT "),
			strings.HasPrefix(string(req.Body), "HEAD"):
			return "HTTP"
		case prefix == "EHLO" || prefix == "HELO" || prefix == "MAIL":
			return "SMTP"
		case prefix == "USER" || prefix == "PASS" || prefix == "RETR":
			return "FTP"
		}
	}

	return ""
}

// extractAndLogPayload extracts protocol-specific payload data for supported protocols
// Mirrors sparring's protocol handlers (applications/) behavior in transparent mode
func (h *TransparentModeHandler) extractAndLogPayload(ctx context.Context, req *Request, conn *ConnectionInfo) {
	appProtocol := conn.AppProtocol
	if appProtocol == "" {
		appProtocol = h.identifyAppProtocol(req)
	}

	// Check if this protocol is in the supported list
	if !h.isProtocolSupported(appProtocol) {
		return
	}

	// Respect MaxPayloadSize limit
	rawData := req.Body
	truncated := false
	if h.config.MaxPayloadSize > 0 && int64(len(rawData)) > h.config.MaxPayloadSize {
		rawData = rawData[:h.config.MaxPayloadSize]
		truncated = true
	}

	payload := &ExtractedPayload{
		ConnectionID: conn.ID,
		Timestamp:    time.Now(),
		Protocol:     appProtocol,
		Direction:    "outgoing",
		RawData:      rawData,
		ParsedData:   make(map[string]interface{}),
		Size:         req.ContentLength,
		Truncated:    truncated,
	}

	// Parse protocol-specific data
	switch appProtocol {
	case "HTTP", "HTTPS":
		h.parseHTTPPayload(req, payload)
	case "DNS":
		h.parseDNSPayload(req, payload)
	case "SMTP":
		h.parseSMTPPayload(req, payload)
	case "FTP":
		h.parseFTPPayload(req, payload)
	}

	// Write to payload log
	h.writePayloadLog(payload)

	atomic.AddInt64(&h.stats.ExtractedPayloads, 1)

	h.logger.DebugContext(ctx, "Payload extracted in transparent mode",
		"protocol", appProtocol,
		"conn_id", conn.ID,
		"size", req.ContentLength,
		"truncated", truncated)
}

// parseHTTPPayload extracts HTTP-specific data from the payload
func (h *TransparentModeHandler) parseHTTPPayload(req *Request, payload *ExtractedPayload) {
	parsed := payload.ParsedData

	if req.Method != "" {
		parsed["method"] = req.Method
	}
	if req.Path != "" {
		parsed["path"] = req.Path
	}
	if req.Domain != "" {
		parsed["host"] = req.Domain
	}
	if len(req.Headers) > 0 {
		parsed["headers"] = req.Headers
	}
	if len(req.Query) > 0 {
		parsed["query_params"] = req.Query
	}

	// Try to parse User-Agent for malware fingerprinting
	if ua, ok := req.Headers["User-Agent"]; ok {
		parsed["user_agent"] = ua
	}

	// Extract interesting headers for malware analysis
	interestingHeaders := []string{
		"Authorization", "X-Api-Key", "X-Auth-Token",
		"Cookie", "Referer", "Origin",
	}
	extracted := map[string]string{}
	for _, h := range interestingHeaders {
		if v, ok := req.Headers[h]; ok {
			extracted[h] = v
		}
	}
	if len(extracted) > 0 {
		parsed["sensitive_headers"] = extracted
	}

	// Try to reconstruct full URL
	if req.Domain != "" && req.Path != "" {
		scheme := "http"
		if req.Port == 443 {
			scheme = "https"
		}
		parsed["full_url"] = fmt.Sprintf("%s://%s%s", scheme, req.Domain, req.Path)
	}

	// Detect if body contains interesting content type
	if ct, ok := req.Headers["Content-Type"]; ok {
		parsed["content_type"] = ct
		if strings.Contains(ct, "application/x-www-form-urlencoded") {
			parsed["body_type"] = "form_data"
		} else if strings.Contains(ct, "application/json") {
			parsed["body_type"] = "json"
		} else if strings.Contains(ct, "multipart") {
			parsed["body_type"] = "multipart"
		}
	}
}

// parseDNSPayload extracts DNS query information
func (h *TransparentModeHandler) parseDNSPayload(req *Request, payload *ExtractedPayload) {
	parsed := payload.ParsedData

	if req.Domain != "" {
		parsed["queried_domain"] = req.Domain
	}
	parsed["dst_ip"] = req.IP

	// Simple DNS packet parsing for query type from port
	if req.Port == 53 {
		parsed["dns_port"] = "standard"
	} else if req.Port == 853 {
		parsed["dns_port"] = "DNS_over_TLS"
	}
}

// parseSMTPPayload extracts SMTP command data
func (h *TransparentModeHandler) parseSMTPPayload(req *Request, payload *ExtractedPayload) {
	parsed := payload.ParsedData

	if len(req.Body) == 0 {
		return
	}

	body := string(req.Body)
	lines := strings.Split(body, "\r\n")
	if len(lines) == 0 {
		lines = strings.Split(body, "\n")
	}

	commands := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		upper := strings.ToUpper(line)
		// Extract SMTP commands (not passwords - for ethical analysis)
		if strings.HasPrefix(upper, "EHLO") ||
			strings.HasPrefix(upper, "HELO") ||
			strings.HasPrefix(upper, "MAIL FROM") ||
			strings.HasPrefix(upper, "RCPT TO") {
			commands = append(commands, line)
		}
	}

	if len(commands) > 0 {
		parsed["smtp_commands"] = commands
	}
	parsed["dst_ip"] = req.IP
	parsed["dst_port"] = req.Port
}

// parseFTPPayload extracts FTP command data
func (h *TransparentModeHandler) parseFTPPayload(req *Request, payload *ExtractedPayload) {
	parsed := payload.ParsedData

	if len(req.Body) == 0 {
		return
	}

	body := strings.TrimSpace(string(req.Body))
	upper := strings.ToUpper(body)

	// Extract FTP commands (exclude PASS for security)
	if strings.HasPrefix(upper, "USER") {
		parts := strings.SplitN(body, " ", 2)
		if len(parts) == 2 {
			parsed["ftp_user"] = parts[1]
		}
	} else if strings.HasPrefix(upper, "RETR") ||
		strings.HasPrefix(upper, "STOR") ||
		strings.HasPrefix(upper, "LIST") ||
		strings.HasPrefix(upper, "CWD") ||
		strings.HasPrefix(upper, "PWD") {
		parsed["ftp_command"] = body
	}

	parsed["dst_ip"] = req.IP
}

// isProtocolSupported checks if a protocol is in the supported list
func (h *TransparentModeHandler) isProtocolSupported(protocol string) bool {
	if protocol == "" {
		return false
	}
	protocol = strings.ToLower(protocol)
	for _, p := range h.config.SupportedProtocols {
		if strings.ToLower(p) == protocol {
			return true
		}
	}
	return false
}

// writeConnectionLog writes a connection event to the connection log file
func (h *TransparentModeHandler) writeConnectionLog(conn *ConnectionInfo, event string) {
	entry := map[string]interface{}{
		"timestamp":    time.Now().UTC().Format(time.RFC3339Nano),
		"event":        event,
		"connection_id": conn.ID,
		"protocol":     conn.Protocol,
		"app_protocol": conn.AppProtocol,
		"src":          fmt.Sprintf("%s:%d", conn.SrcIP, conn.SrcPort),
		"dst":          fmt.Sprintf("%s:%d", conn.DstIP, conn.DstPort),
		"domain":       conn.Domain,
		"bytes_sent":   conn.BytesSent,
	}

	h.writeJSON(h.connWriter, entry, "connection")
}

// writePayloadLog writes an extracted payload entry to the payload log file
func (h *TransparentModeHandler) writePayloadLog(payload *ExtractedPayload) {
	h.writeJSON(h.payloadWriter, payload, "payload")
}

// writeJSON writes a JSON entry to the given writer
func (h *TransparentModeHandler) writeJSON(w *bufio.Writer, data interface{}, logType string) {
	h.writerMu.Lock()
	defer h.writerMu.Unlock()

	b, err := json.Marshal(data)
	if err != nil {
		h.logger.Warn("Failed to marshal log entry", "type", logType, "error", err)
		return
	}

	if w != nil {
		_, _ = w.Write(b)
		_ = w.WriteByte('\n')
		_ = w.Flush()
	} else {
		// Fallback: log to slogger
		h.logger.Info("Transparent mode log entry",
			"type", logType,
			"data", string(b))
	}
}

// GetStats returns transparent mode statistics
func (h *TransparentModeHandler) GetStats() TransparentStats {
	return h.stats
}

// GetConnections returns a snapshot of all tracked connections
func (h *TransparentModeHandler) GetConnections() []*ConnectionInfo {
	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()

	result := make([]*ConnectionInfo, 0, len(h.connections))
	for _, conn := range h.connections {
		result = append(result, conn)
	}
	return result
}

// GetConnectionStats returns statistics about tracked connections
func (h *TransparentModeHandler) GetConnectionStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_connections":   atomic.LoadInt64(&h.stats.TotalConnections),
		"tcp_connections":     atomic.LoadInt64(&h.stats.TCPConnections),
		"udp_connections":     atomic.LoadInt64(&h.stats.UDPConnections),
		"icmp_packets":        atomic.LoadInt64(&h.stats.ICMPPackets),
		"total_bytes":         atomic.LoadInt64(&h.stats.TotalBytesObserved),
		"extracted_payloads":  atomic.LoadInt64(&h.stats.ExtractedPayloads),
		"unknown_protocols":   atomic.LoadInt64(&h.stats.UnknownProtocols),
	}

	// Collect protocol breakdown
	protocols := map[string]int64{}
	h.stats.ProtocolBreakdown.Range(func(k, v interface{}) bool {
		protocols[k.(string)] = atomic.LoadInt64(v.(*int64))
		return true
	})
	stats["protocol_breakdown"] = protocols

	return stats
}

// PrintSummary prints a text summary of observed traffic
// Mirrors sparring's print_connections() and print_stats() functions
func (h *TransparentModeHandler) PrintSummary() string {
	var sb strings.Builder

	sb.WriteString("\n=== TRANSPARENT MODE TRAFFIC SUMMARY ===\n")
	sb.WriteString(fmt.Sprintf("Total Connections:    %d\n", atomic.LoadInt64(&h.stats.TotalConnections)))
	sb.WriteString(fmt.Sprintf("  TCP:               %d\n", atomic.LoadInt64(&h.stats.TCPConnections)))
	sb.WriteString(fmt.Sprintf("  UDP:               %d\n", atomic.LoadInt64(&h.stats.UDPConnections)))
	sb.WriteString(fmt.Sprintf("  ICMP packets:      %d\n", atomic.LoadInt64(&h.stats.ICMPPackets)))
	sb.WriteString(fmt.Sprintf("Total Bytes Observed: %d\n", atomic.LoadInt64(&h.stats.TotalBytesObserved)))
	sb.WriteString(fmt.Sprintf("Extracted Payloads:   %d\n", atomic.LoadInt64(&h.stats.ExtractedPayloads)))
	sb.WriteString(fmt.Sprintf("Unknown Protocols:    %d\n", atomic.LoadInt64(&h.stats.UnknownProtocols)))

	sb.WriteString("\nProtocol Breakdown:\n")
	h.stats.ProtocolBreakdown.Range(func(k, v interface{}) bool {
		sb.WriteString(fmt.Sprintf("  %-10s: %d connections\n",
			k.(string), atomic.LoadInt64(v.(*int64))))
		return true
	})

	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()

	sb.WriteString(fmt.Sprintf("\nActive Connections (%d):\n", len(h.connections)))
	for _, conn := range h.connections {
		sb.WriteString(fmt.Sprintf("  [%s] %s:%s -> %s:%s",
			conn.Protocol,
			conn.SrcIP, intToStr(conn.SrcPort),
			conn.DstIP, intToStr(conn.DstPort)))
		if conn.Domain != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", conn.Domain))
		}
		if conn.AppProtocol != "" {
			sb.WriteString(fmt.Sprintf(" [%s]", conn.AppProtocol))
		}
		sb.WriteString(fmt.Sprintf(" sent=%d bytes\n", conn.BytesSent))
	}

	sb.WriteString("=========================================\n")
	return sb.String()
}

// Close flushes and closes log files
func (h *TransparentModeHandler) Close() error {
	h.writerMu.Lock()
	defer h.writerMu.Unlock()

	if h.connWriter != nil {
		_ = h.connWriter.Flush()
	}
	if h.payloadWriter != nil {
		_ = h.payloadWriter.Flush()
	}
	if h.connLogFile != nil {
		_ = h.connLogFile.Close()
	}
	if h.payloadLogFile != nil {
		_ = h.payloadLogFile.Close()
	}
	return nil
}

// --- Helpers ---

// openLogFile opens (or creates) a log file for appending
func openLogFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

// intToStr converts an int to string (helper to avoid fmt.Sprintf in hot path)
func intToStr(n int) string {
	return strconv.Itoa(n)
}

// parseUserAgent parses a User-Agent string into components
// Extra utility for HTTP payload analysis
func parseUserAgent(ua string) map[string]string {
	result := map[string]string{"raw": ua}
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("User-Agent", ua)
	result["parsed"] = r.Header.Get("User-Agent")
	return result
}
