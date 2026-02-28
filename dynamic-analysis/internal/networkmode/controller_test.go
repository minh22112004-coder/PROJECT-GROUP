package networkmode

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestMode_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		mode  Mode
		valid bool
	}{
		{
			name:  "Full Mode",
			mode:  ModeFull,
			valid: true,
		},
		{
			name:  "Half Mode",
			mode:  ModeHalf,
			valid: true,
		},
		{
			name:  "Transparent Mode",
			mode:  ModeTransparent,
			valid: true,
		},
		{
			name:  "Invalid",
			mode:  Mode("invalid"),
			valid: false,
		},
		{
			name:  "Empty",
			mode:  Mode(""),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsValid(); got != tt.valid {
				t.Errorf("Mode.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Mode != ModeFull {
		t.Errorf("Default mode should be Full, got %s", config.Mode)
	}

	if !config.FullMode.CompleteIsolation {
		t.Error("Default Full Mode should have complete isolation")
	}

	if config.HalfMode.Enabled {
		t.Error("Default Half Mode should be disabled")
	}

	if err := config.Validate(); err != nil {
		t.Errorf("Default config should be valid: %v", err)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "Valid Full Mode",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "Valid Half Mode",
			config: &Config{
				Mode: ModeHalf,
				HalfMode: &HalfModeConfig{
					Enabled:       true,
					DefaultAction: ActionSimulate,
					Proxy: &ProxyConfig{
						ListenAddress: "0.0.0.0:8888",
					},
				},
				Logging: &LoggingConfig{Level: "info"},
			},
			wantErr: false,
		},
		{
			name: "Valid Transparent Mode",
			config: &Config{
				Mode: ModeTransparent,
				TransparentMode: &TransparentModeConfig{
					Enabled:            true,
					ExtractPayloads:    true,
					LogConnections:     true,
					SupportedProtocols: []string{"http", "dns"},
				},
				Logging: &LoggingConfig{Level: "info"},
			},
			wantErr: false,
		},
		{
			name: "Transparent Mode Not Enabled",
			config: &Config{
				Mode: ModeTransparent,
				TransparentMode: &TransparentModeConfig{
					Enabled: false,
				},
			},
			wantErr: true,
		},
		{
			name: "Missing Transparent Mode Config",
			config: &Config{
				Mode:            ModeTransparent,
				TransparentMode: nil,
			},
			wantErr: true,
		},
		{
			name: "Invalid Mode",
			config: &Config{
				Mode: Mode("invalid"),
			},
			wantErr: true,
		},
		{
			name: "Missing Full Mode Config",
			config: &Config{
				Mode:     ModeFull,
				FullMode: nil,
			},
			wantErr: true,
		},
		{
			name: "Half Mode Not Enabled",
			config: &Config{
				Mode: ModeHalf,
				HalfMode: &HalfModeConfig{
					Enabled: false,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAction_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		action Action
		valid  bool
	}{
		{"Forward", ActionForward, true},
		{"Block", ActionBlock, true},
		{"Modify", ActionModify, true},
		{"Simulate", ActionSimulate, true},
		{"Invalid", Action("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.action.IsValid(); got != tt.valid {
				t.Errorf("Action.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestDecisionEngine_AddRule(t *testing.T) {
	config := &HalfModeConfig{
		DefaultAction: ActionSimulate,
	}
	engine := NewDecisionEngine(config, slog.Default())

	rule := DecisionRule{
		Name:     "test_rule",
		Priority: 100,
		Enabled:  true,
		Action:   ActionBlock,
		Condition: &RuleCondition{
			Type: ConditionDefault,
		},
	}

	err := engine.AddRule(rule)
	if err != nil {
		t.Errorf("AddRule() error = %v", err)
	}

	rules := engine.GetRules()
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
}

func TestDecisionEngine_Decide(t *testing.T) {
	config := &HalfModeConfig{
		DefaultAction: ActionSimulate,
	}
	engine := NewDecisionEngine(config, slog.Default())

	// Add rules
	rules := []DecisionRule{
		{
			Name:     "block_evil",
			Priority: 100,
			Enabled:  true,
			Action:   ActionBlock,
			Condition: &RuleCondition{
				Type:    ConditionDomainBlacklist,
				Domains: []string{"evil.com"},
			},
		},
		{
			Name:     "allow_good",
			Priority: 90,
			Enabled:  true,
			Action:   ActionForward,
			Condition: &RuleCondition{
				Type:    ConditionDomainWhitelist,
				Domains: []string{"good.com"},
			},
		},
	}

	for _, rule := range rules {
		if err := engine.AddRule(rule); err != nil {
			t.Fatalf("Failed to add rule: %v", err)
		}
	}

	tests := []struct {
		name           string
		req            *Request
		expectedAction Action
	}{
		{
			name: "Block evil domain",
			req: &Request{
				ID:     "1",
				Domain: "evil.com",
			},
			expectedAction: ActionBlock,
		},
		{
			name: "Forward good domain",
			req: &Request{
				ID:     "2",
				Domain: "good.com",
			},
			expectedAction: ActionForward,
		},
		{
			name: "Default action for unknown",
			req: &Request{
				ID:     "3",
				Domain: "unknown.com",
			},
			expectedAction: ActionSimulate,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := engine.Decide(ctx, tt.req)
			if err != nil {
				t.Errorf("Decide() error = %v", err)
				return
			}
			if decision.Action != tt.expectedAction {
				t.Errorf("Decide() action = %v, want %v", decision.Action, tt.expectedAction)
			}
		})
	}
}

func TestDecisionEngine_DomainMatching(t *testing.T) {
	config := &HalfModeConfig{
		DefaultAction: ActionSimulate,
	}
	engine := NewDecisionEngine(config, slog.Default())

	tests := []struct {
		name     string
		domain   string
		patterns []string
		expected bool
	}{
		{
			name:     "Exact match",
			domain:   "example.com",
			patterns: []string{"example.com"},
			expected: true,
		},
		{
			name:     "Wildcard match",
			domain:   "sub.example.com",
			patterns: []string{"*.example.com"},
			expected: true,
		},
		{
			name:     "No match",
			domain:   "other.com",
			patterns: []string{"example.com"},
			expected: false,
		},
		{
			name:     "Wildcard root match",
			domain:   "example.com",
			patterns: []string{"*.example.com"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.matchDomainList(tt.domain, tt.patterns, true)
			if got != tt.expected {
				t.Errorf("matchDomainList() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestController_NewController(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "Valid Full Mode",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "Valid Half Mode",
			config: &Config{
				Mode: ModeHalf,
				HalfMode: &HalfModeConfig{
					Enabled:       true,
					DefaultAction: ActionSimulate,
					Proxy: &ProxyConfig{
						ListenAddress: "0.0.0.0:8888",
					},
					TrafficModifier: &TrafficModifierConfig{
						Enabled: true,
					},
				},
				Logging: &LoggingConfig{
					Level: "info",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid Config",
			config: &Config{
				Mode: Mode("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, err := NewController(tt.config, slog.Default())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewController() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if controller != nil {
				defer controller.Close()
			}
		})
	}
}

func TestController_HandleRequest_FullMode(t *testing.T) {
	config := DefaultConfig()
	controller, err := NewController(config, slog.Default())
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}
	defer controller.Close()

	req := &Request{
		ID:        "test-1",
		Timestamp: time.Now(),
		Protocol:  string(ProtocolHTTP),
		Domain:    "example.com",
		Path:      "/test",
		Method:    "GET",
		Headers:   make(map[string]string),
	}

	ctx := context.Background()
	resp, err := controller.HandleRequest(ctx, req)
	if err != nil {
		t.Errorf("HandleRequest() error = %v", err)
		return
	}

	if resp == nil {
		t.Error("HandleRequest() returned nil response")
		return
	}

	if resp.Source != "simulated" {
		t.Errorf("Expected simulated response, got %s", resp.Source)
	}

	if resp.Decision == nil {
		t.Error("Response should have decision")
	} else if resp.Decision.Action != ActionSimulate {
		t.Errorf("Expected simulate action, got %s", resp.Decision.Action)
	}
}

func TestController_SwitchMode(t *testing.T) {
	config := DefaultConfig()
	config.HalfMode.Enabled = true // Enable Half Mode

	controller, err := NewController(config, slog.Default())
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}
	defer controller.Close()

	// Should start in Full Mode
	if controller.GetMode() != ModeFull {
		t.Errorf("Expected Full Mode, got %s", controller.GetMode())
	}

	// Switch to Half Mode
	ctx := context.Background()
	err = controller.SwitchMode(ctx, ModeHalf)
	if err != nil {
		t.Errorf("SwitchMode() error = %v", err)
	}

	if controller.GetMode() != ModeHalf {
		t.Errorf("Expected Half Mode after switch, got %s", controller.GetMode())
	}

	// Switch back to Full Mode
	err = controller.SwitchMode(ctx, ModeFull)
	if err != nil {
		t.Errorf("SwitchMode() error = %v", err)
	}

	if controller.GetMode() != ModeFull {
		t.Errorf("Expected Full Mode after switch, got %s", controller.GetMode())
	}
}

func TestController_TransparentMode(t *testing.T) {
	config := DefaultConfig()
	config.Mode = ModeTransparent
	config.TransparentMode = &TransparentModeConfig{
		Enabled:            true,
		ExtractPayloads:    true,
		LogConnections:     true,
		LogICMP:            true,
		SupportedProtocols: []string{"http", "dns", "smtp", "ftp"},
		// Leave log files empty to use slogger fallback
	}

	controller, err := NewController(config, slog.Default())
	if err != nil {
		t.Fatalf("Failed to create controller in transparent mode: %v", err)
	}
	defer controller.Close()

	if controller.GetMode() != ModeTransparent {
		t.Errorf("Expected Transparent Mode, got %s", controller.GetMode())
	}

	ctx := context.Background()

	// Handle an HTTP request - should pass through unmodified
	req := &Request{
		ID:         "test-transparent-001",
		Timestamp:  time.Now(),
		Protocol:   string(ProtocolHTTP),
		Method:     "GET",
		Domain:     "example.com",
		IP:         "93.184.216.34",
		Port:       80,
		Path:       "/index.html",
		SourceIP:   "192.168.1.100",
		SourcePort: 54321,
		Headers: map[string]string{
			"User-Agent": "TestMalware/1.0",
			"Host":       "example.com",
		},
	}

	resp, err := controller.HandleRequest(ctx, req)
	if err != nil {
		t.Fatalf("HandleRequest() in transparent mode error = %v", err)
	}

	if resp == nil {
		t.Fatal("Expected non-nil response in transparent mode")
	}

	// Transparent mode must always produce a passthrough response
	if resp.Source != "transparent_passthrough" {
		t.Errorf("Expected source=transparent_passthrough, got %s", resp.Source)
	}

	// Decision must be ActionForward (never block in transparent mode)
	if resp.Decision == nil {
		t.Error("Response should have decision")
	} else if resp.Decision.Action != ActionForward {
		t.Errorf("Transparent mode decision must be Forward, got %s", resp.Decision.Action)
	}

	// Handle a DNS request
	dnsReq := &Request{
		ID:         "test-transparent-dns-001",
		Timestamp:  time.Now(),
		Protocol:   string(ProtocolDNS),
		Domain:     "malware-c2.example.com",
		IP:         "8.8.8.8",
		Port:       53,
		SourceIP:   "192.168.1.100",
		SourcePort: 12345,
	}

	dnsResp, err := controller.HandleRequest(ctx, dnsReq)
	if err != nil {
		t.Fatalf("HandleRequest() DNS in transparent mode error = %v", err)
	}
	if dnsResp == nil {
		t.Fatal("Expected non-nil DNS response in transparent mode")
	}

	// Verify transparent mode stats are accessible
	stats, err := controller.GetTransparentStats()
	if err != nil {
		t.Errorf("GetTransparentStats() error = %v", err)
	}
	if v, ok := stats["total_connections"]; !ok || v.(int64) < 1 {
		t.Errorf("Expected at least 1 connection tracked, got %v", stats["total_connections"])
	}

	// Summary should be non-empty
	summary, err := controller.GetTransparentSummary()
	if err != nil {
		t.Errorf("GetTransparentSummary() error = %v", err)
	}
	if summary == "" {
		t.Error("Expected non-empty transparent summary")
	}
}

func TestController_SwitchToTransparentMode(t *testing.T) {
	config := DefaultConfig()
	config.HalfMode.Enabled = true
	config.TransparentMode = &TransparentModeConfig{
		Enabled:            true,
		SupportedProtocols: []string{"http", "dns"},
	}

	controller, err := NewController(config, slog.Default())
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}
	defer controller.Close()

	ctx := context.Background()

	// Switch Full -> Transparent
	err = controller.SwitchMode(ctx, ModeTransparent)
	if err != nil {
		t.Errorf("SwitchMode to Transparent error = %v", err)
	}
	if controller.GetMode() != ModeTransparent {
		t.Errorf("Expected Transparent Mode, got %s", controller.GetMode())
	}

	// Switch Transparent -> Full
	err = controller.SwitchMode(ctx, ModeFull)
	if err != nil {
		t.Errorf("SwitchMode back to Full error = %v", err)
	}
	if controller.GetMode() != ModeFull {
		t.Errorf("Expected Full Mode, got %s", controller.GetMode())
	}
}

func TestController_GetStats(t *testing.T) {
	config := DefaultConfig()
	controller, err := NewController(config, slog.Default())
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}
	defer controller.Close()

	stats := controller.GetStats()
	if stats == nil {
		t.Error("GetStats() returned nil")
		return
	}

	if stats.TotalRequests != 0 {
		t.Errorf("Expected 0 total requests, got %d", stats.TotalRequests)
	}

	// Handle a request
	req := &Request{
		ID:        "test-1",
		Timestamp: time.Now(),
		Protocol:  string(ProtocolHTTP),
		Domain:    "example.com",
	}

	ctx := context.Background()
	_, err = controller.HandleRequest(ctx, req)
	if err != nil {
		t.Errorf("HandleRequest() error = %v", err)
	}

	stats = controller.GetStats()
	if stats.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", stats.TotalRequests)
	}

	if stats.SimulatedRequests != 1 {
		t.Errorf("Expected 1 simulated request, got %d", stats.SimulatedRequests)
	}
}

func TestController_Health(t *testing.T) {
	config := DefaultConfig()
	controller, err := NewController(config, slog.Default())
	if err != nil {
		t.Fatalf("Failed to create controller: %v", err)
	}
	defer controller.Close()

	ctx := context.Background()
	err = controller.Health(ctx)
	if err != nil {
		t.Errorf("Health() error = %v", err)
	}
}

func TestDefaultRules(t *testing.T) {
	rules := DefaultRules()

	if len(rules) == 0 {
		t.Error("DefaultRules() returned no rules")
	}

	// Check that rules have priorities
	for i, rule := range rules {
		if i > 0 && rules[i-1].Priority < rule.Priority {
			t.Errorf("Rules are not sorted by priority: rule %d has lower priority than rule %d",
				i-1, i)
		}

		if !rule.Action.IsValid() {
			t.Errorf("Rule %s has invalid action: %s", rule.Name, rule.Action)
		}

		if rule.Condition == nil {
			t.Errorf("Rule %s has nil condition", rule.Name)
		}
	}
}
