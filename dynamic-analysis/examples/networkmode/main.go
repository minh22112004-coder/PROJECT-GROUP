// Package main demonstrates how to use the Network Mode Controller
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/ossf/package-analysis/internal/networkmode"
)

func main() {
	// Example 1: Full Mode (Isolated)
	fmt.Println("=== Example 1: Full Mode (Isolated) ===")
	runFullModeExample()

	fmt.Println()

	// Example 2: Half Mode (Transparent Proxy)
	fmt.Println("=== Example 2: Half Mode (Transparent Proxy) ===")
	runHalfModeExample()

	fmt.Println()

	// Example 3: Mode Switching
	fmt.Println("=== Example 3: Mode Switching ===")
	runModeSwitchingExample()

	fmt.Println()

	// Example 4: Custom Decision Rules
	fmt.Println("=== Example 4: Custom Decision Rules ===")
	runCustomRulesExample()

	fmt.Println()

	// Example 5: Transparent Mode (Passive Observation)
	// Inspired by siemens/sparring transparent mode
	fmt.Println("=== Example 5: Transparent Mode (Passive Observation) ===")
	runTransparentModeExample()
}

// runFullModeExample demonstrates Full Mode usage
func runFullModeExample() {
	// Create default configuration (Full Mode)
	config := networkmode.DefaultConfig()

	// Create controller
	controller, err := networkmode.NewController(config, slog.Default())
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}
	defer controller.Close()

	// Create sample requests
	requests := []*networkmode.Request{
		{
			ID:        "req-001",
			Timestamp: time.Now(),
			Protocol:  string(networkmode.ProtocolHTTP),
			Domain:    "malware-c2.com",
			Path:      "/api/command",
			Method:    "GET",
			Headers:   make(map[string]string),
		},
		{
			ID:        "req-002",
			Timestamp: time.Now(),
			Protocol:  string(networkmode.ProtocolHTTPS),
			Domain:    "evil-server.net",
			Path:      "/download/payload.exe",
			Method:    "GET",
			Headers:   make(map[string]string),
		},
		{
			ID:        "req-003",
			Timestamp: time.Now(),
			Protocol:  string(networkmode.ProtocolDNS),
			Domain:    "suspicious-domain.xyz",
			Headers:   make(map[string]string),
		},
	}

	// Handle requests
	ctx := context.Background()
	for _, req := range requests {
		resp, err := controller.HandleRequest(ctx, req)
		if err != nil {
			log.Printf("Error handling request %s: %v", req.ID, err)
			continue
		}

		fmt.Printf("Request: %s to %s\n", req.Protocol, req.Domain)
		fmt.Printf("  Response Source: %s\n", resp.Source)
		fmt.Printf("  Decision: %s (reason: %s)\n", resp.Decision.Action, resp.Decision.Reason)
		fmt.Printf("  Body Preview: %s\n", preview(resp.Body, 80))
		fmt.Println()
	}

	// Print statistics
	stats := controller.GetStats()
	fmt.Printf("Statistics:\n")
	fmt.Printf("  Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("  Simulated: %d\n", stats.SimulatedRequests)
	fmt.Printf("  Errors: %d\n", stats.Errors)
}

// runHalfModeExample demonstrates Half Mode usage
func runHalfModeExample() {
	// Create Half Mode configuration
	config := networkmode.DefaultConfig()
	config.Mode = networkmode.ModeHalf
	config.HalfMode.Enabled = true
	config.HalfMode.DefaultAction = networkmode.ActionSimulate

	// Create controller
	controller, err := networkmode.NewController(config, slog.Default())
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}
	defer controller.Close()

	// Create sample requests
	requests := []*networkmode.Request{
		{
			ID:        "req-101",
			Timestamp: time.Now(),
			Protocol:  string(networkmode.ProtocolHTTP),
			Domain:    "malware-c2.com", // Will be blocked
			Path:      "/api/exfiltrate",
			Method:    "POST",
			Headers:   make(map[string]string),
		},
		{
			ID:        "req-102",
			Timestamp: time.Now(),
			Protocol:  string(networkmode.ProtocolHTTPS),
			Domain:    "cdn.cloudflare.com", // Will be forwarded
			Path:      "/libs/jquery.js",
			Method:    "GET",
			Headers:   make(map[string]string),
		},
		{
			ID:        "req-103",
			Timestamp: time.Now(),
			Protocol:  string(networkmode.ProtocolHTTP),
			Domain:    "example.com",
			Path:      "/malware.exe", // Will be sandboxed
			Method:    "GET",
			Headers:   make(map[string]string),
		},
	}

	// Handle requests
	ctx := context.Background()
	for _, req := range requests {
		resp, err := controller.HandleRequest(ctx, req)
		if err != nil {
			log.Printf("Error handling request %s: %v", req.ID, err)
			continue
		}

		fmt.Printf("Request: %s %s%s\n", req.Method, req.Domain, req.Path)
		fmt.Printf("  Decision: %s (rule: %s)\n", resp.Decision.Action, resp.Decision.RuleName)
		fmt.Printf("  Reason: %s\n", resp.Decision.Reason)
		fmt.Printf("  Response Source: %s\n", resp.Source)
		fmt.Println()
	}

	// Print statistics
	stats := controller.GetStats()
	fmt.Printf("Statistics:\n")
	fmt.Printf("  Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("  Forwarded: %d\n", stats.ForwardedRequests)
	fmt.Printf("  Blocked: %d\n", stats.BlockedRequests)
	fmt.Printf("  Modified: %d\n", stats.ModifiedRequests)
	fmt.Printf("  Simulated: %d\n", stats.SimulatedRequests)
}

// runModeSwitchingExample demonstrates mode switching
func runModeSwitchingExample() {
	// Start in Full Mode
	config := networkmode.DefaultConfig()
	config.HalfMode.Enabled = true // Enable Half Mode for switching

	controller, err := networkmode.NewController(config, slog.Default())
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}
	defer controller.Close()

	ctx := context.Background()

	// Test in Full Mode
	fmt.Printf("Current Mode: %s\n", controller.GetMode())
	
	req := &networkmode.Request{
		ID:       "req-201",
		Protocol: string(networkmode.ProtocolHTTP),
		Domain:   "test.com",
		Method:   "GET",
		Headers:  make(map[string]string),
	}
	
	resp, _ := controller.HandleRequest(ctx, req)
	fmt.Printf("  Response Source: %s\n\n", resp.Source)

	// Switch to Half Mode
	if err := controller.SwitchMode(ctx, networkmode.ModeHalf); err != nil {
		log.Printf("Failed to switch mode: %v", err)
		return
	}

	fmt.Printf("Switched to Mode: %s\n", controller.GetMode())
	
	req.ID = "req-202"
	resp, _ = controller.HandleRequest(ctx, req)
	fmt.Printf("  Response Source: %s\n\n", resp.Source)

	// Switch back to Full Mode
	if err := controller.SwitchMode(ctx, networkmode.ModeFull); err != nil {
		log.Printf("Failed to switch mode: %v", err)
		return
	}

	fmt.Printf("Switched back to Mode: %s\n", controller.GetMode())
	
	req.ID = "req-203"
	resp, _ = controller.HandleRequest(ctx, req)
	fmt.Printf("  Response Source: %s\n", resp.Source)
}

// runCustomRulesExample demonstrates custom decision rules
func runCustomRulesExample() {
	// Create Half Mode configuration
	config := networkmode.DefaultConfig()
	config.Mode = networkmode.ModeHalf
	config.HalfMode.Enabled = true
	config.HalfMode.DefaultAction = networkmode.ActionBlock // Block by default

	controller, err := networkmode.NewController(config, slog.Default())
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}
	defer controller.Close()

	ctx := context.Background()

	// Add custom rules
	customRules := []networkmode.DecisionRule{
		{
			Name:     "allow_my_domain",
			Priority: 200,
			Enabled:  true,
			Condition: &networkmode.RuleCondition{
				Type:    networkmode.ConditionDomainWhitelist,
				Domains: []string{"mycompany.com", "*.mycompany.com"},
			},
			Action: networkmode.ActionForward,
		},
		{
			Name:     "block_dangerous_scripts",
			Priority: 150,
			Enabled:  true,
			Condition: &networkmode.RuleCondition{
				Type:           networkmode.ConditionFileExtension,
				FileExtensions: []string{".ps1", ".bat", ".cmd", ".vbs"},
			},
			Action: networkmode.ActionBlock,
		},
		{
			Name:     "monitor_api_calls",
			Priority: 100,
			Enabled:  true,
			Condition: &networkmode.RuleCondition{
				Type: networkmode.ConditionDomainPattern,
				DomainPattern: ".*\\.api\\.",
			},
			Action: networkmode.ActionModify,
			Modifier: &networkmode.Modifier{
				Type:           "content_logging",
				LogFullContent: true,
			},
		},
	}

	for _, rule := range customRules {
		if err := controller.AddDecisionRule(ctx, rule); err != nil {
			log.Printf("Failed to add rule %s: %v", rule.Name, err)
		} else {
			fmt.Printf("Added rule: %s (priority: %d)\n", rule.Name, rule.Priority)
		}
	}

	fmt.Println()

	// Test requests against custom rules
	testRequests := []*networkmode.Request{
		{
			ID:       "req-301",
			Protocol: string(networkmode.ProtocolHTTPS),
			Domain:   "app.mycompany.com",
			Path:     "/api/data",
			Method:   "GET",
			Headers:  make(map[string]string),
		},
		{
			ID:       "req-302",
			Protocol: string(networkmode.ProtocolHTTP),
			Domain:   "evil.com",
			Path:     "/script.ps1",
			Method:   "GET",
			Headers:  make(map[string]string),
		},
		{
			ID:       "req-303",
			Protocol: string(networkmode.ProtocolHTTPS),
			Domain:   "service.api.example.com",
			Path:     "/v1/users",
			Method:   "POST",
			Headers:  make(map[string]string),
		},
	}

	for _, req := range testRequests {
		resp, err := controller.HandleRequest(ctx, req)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		fmt.Printf("Request: %s%s\n", req.Domain, req.Path)
		fmt.Printf("  Decision: %s\n", resp.Decision.Action)
		fmt.Printf("  Rule: %s\n", resp.Decision.RuleName)
		fmt.Printf("  Reason: %s\n", resp.Decision.Reason)
		fmt.Println()
	}

	// Print all rules
	fmt.Println("All Active Rules:")
	for i, rule := range controller.GetDecisionRules() {
		fmt.Printf("%d. %s (priority: %d, action: %s)\n",
			i+1, rule.Name, rule.Priority, rule.Action)
	}
}

// runTransparentModeExample demonstrates Transparent Mode usage.
//
// Transparent Mode is inspired by siemens/sparring:
//   https://github.com/siemens/sparring
//
// In this mode sparring (and this controller) will NOT alter any transmitted
// data and only log connections and try to extract interesting data for
// supported protocols. Ideal for:
//   - Passive forensic analysis (don't disturb malware behavior)
//   - Baseline profiling before switching to Full/Half mode
//   - Observing C2 communication patterns in a real network
func runTransparentModeExample() {
	config := networkmode.DefaultConfig()
	config.Mode = networkmode.ModeTransparent
	config.TransparentMode = &networkmode.TransparentModeConfig{
		Enabled:            true,
		ExtractPayloads:    true,
		LogConnections:     true,
		LogICMP:            true,
		SupportedProtocols: []string{"http", "https", "dns", "smtp", "ftp"},
		// Log files intentionally left empty → falls back to slog output
		// In production: set ConnectionLogFile and PayloadLogFile
	}

	controller, err := networkmode.NewController(config, slog.Default())
	if err != nil {
		log.Fatalf("Failed to create transparent mode controller: %v", err)
	}
	defer controller.Close()

	ctx := context.Background()
	fmt.Printf("Network Mode: %s\n\n", controller.GetMode())

	// Simulate traffic from a malware sample
	requests := []*networkmode.Request{
		// HTTP C2 beacon
		{
			ID:         "obs-001",
			Timestamp:  time.Now(),
			Protocol:   string(networkmode.ProtocolHTTP),
			Method:     "POST",
			Domain:     "malware-c2.example.com",
			IP:         "203.0.113.42",
			Port:       80,
			Path:       "/beacon",
			SourceIP:   "192.168.1.50",
			SourcePort: 55001,
			Headers: map[string]string{
				"User-Agent":   "Mozilla/5.0 (compatible; bot/1.0)",
				"Content-Type": "application/json",
			},
			Body:          []byte(`{"id":"infected-host-01","cmd":"check-in"}`),
			ContentLength: 42,
		},
		// DNS lookup for C2 domain
		{
			ID:         "obs-002",
			Timestamp:  time.Now(),
			Protocol:   string(networkmode.ProtocolDNS),
			Domain:     "update.malware-c2.example.com",
			IP:         "8.8.8.8",
			Port:       53,
			SourceIP:   "192.168.1.50",
			SourcePort: 43210,
		},
		// SMTP exfiltration attempt
		{
			ID:         "obs-003",
			Timestamp:  time.Now(),
			Protocol:   string(networkmode.ProtocolSMTP),
			Domain:     "mail.attacker.net",
			IP:         "198.51.100.7",
			Port:       25,
			SourceIP:   "192.168.1.50",
			SourcePort: 60000,
			Body:       []byte("EHLO infected-host\r\nMAIL FROM:<malware@victim.com>\r\nRCPT TO:<attacker@evil.net>\r\n"),
		},
	}

	for _, req := range requests {
		resp, err := controller.HandleRequest(ctx, req)
		if err != nil {
			log.Printf("Error handling request %s: %v", req.ID, err)
			continue
		}

		fmt.Printf("[OBSERVED] %s %s:%d → %s\n",
			req.Protocol, req.SourceIP, req.SourcePort, req.Domain)
		fmt.Printf("  Action  : %s (traffic NOT modified)\n", resp.Decision.Action)
		fmt.Printf("  Source  : %s\n", resp.Source)
		fmt.Printf("  Rule    : %s\n", resp.Decision.RuleName)
		fmt.Println()
	}

	// Print transparent mode statistics
	stats, err := controller.GetTransparentStats()
	if err != nil {
		log.Printf("GetTransparentStats error: %v", err)
		return
	}

	fmt.Println("--- Transparent Mode Statistics ---")
	fmt.Printf("Total connections tracked : %v\n", stats["total_connections"])
	fmt.Printf("TCP connections           : %v\n", stats["tcp_connections"])
	fmt.Printf("UDP connections           : %v\n", stats["udp_connections"])
	fmt.Printf("Extracted payloads        : %v\n", stats["extracted_payloads"])
	if breakdown, ok := stats["protocol_breakdown"].(map[string]int64); ok {
		fmt.Printf("Protocol breakdown:\n")
		for proto, count := range breakdown {
			fmt.Printf("  %-10s: %d\n", proto, count)
		}
	}

	// Print full connection summary (mirrors sparring's print_connections)
	summary, _ := controller.GetTransparentSummary()
	fmt.Print(summary)
}

// preview returns a preview of data (truncated if too long)
func preview(data []byte, maxLen int) string {
	if len(data) == 0 {
		return "(empty)"
	}
	if len(data) <= maxLen {
		return string(data)
	}
	return string(data[:maxLen]) + "..."
}
