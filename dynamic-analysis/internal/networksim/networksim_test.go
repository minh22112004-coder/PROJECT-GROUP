package networksim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "with nil config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &Config{
				INetSimDNSAddr:  "192.168.1.100:53",
				INetSimHTTPAddr: "192.168.1.100:80",
				Enabled:         true,
				LivenessTimeout: 5 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := New(tt.config)
			if ns == nil {
				t.Error("New() returned nil")
			}
			if ns.config == nil {
				t.Error("NetworkSimulator config is nil")
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.INetSimDNSAddr != "172.20.0.2:53" {
		t.Errorf("Expected DNS addr 172.20.0.2:53, got %s", cfg.INetSimDNSAddr)
	}
	
	if cfg.INetSimHTTPAddr != "172.20.0.2:80" {
		t.Errorf("Expected HTTP addr 172.20.0.2:80, got %s", cfg.INetSimHTTPAddr)
	}
	
	if cfg.Enabled {
		t.Error("Expected Enabled to be false by default")
	}
	
	if cfg.LivenessTimeout != 3*time.Second {
		t.Errorf("Expected timeout 3s, got %v", cfg.LivenessTimeout)
	}
}

func TestIsURLAlive(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name     string
		enabled  bool
		url      string
		expected bool
	}{
		{
			name:     "simulation disabled - returns true",
			enabled:  false,
			url:      "http://nonexistent.example.com",
			expected: true,
		},
		{
			name:     "simulation enabled - alive URL",
			enabled:  true,
			url:      server.URL,
			expected: true,
		},
		{
			name:     "simulation enabled - dead URL",
			enabled:  true,
			url:      "http://definitely-does-not-exist-12345.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Enabled = tt.enabled
			cfg.LivenessTimeout = 2 * time.Second
			
			ns := New(cfg)
			ctx := context.Background()
			
			result := ns.IsURLAlive(ctx, tt.url)
			if result != tt.expected {
				t.Errorf("IsURLAlive() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestShouldRedirectToINetSim(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name     string
		enabled  bool
		url      string
		expected bool
	}{
		{
			name:     "simulation disabled",
			enabled:  false,
			url:      "http://dead.example.com",
			expected: false,
		},
		{
			name:     "alive URL - no redirect",
			enabled:  true,
			url:      server.URL,
			expected: false,
		},
		{
			name:     "dead URL - should redirect",
			enabled:  true,
			url:      "http://dead-malware-site.example.com",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Enabled = tt.enabled
			cfg.LivenessTimeout = 2 * time.Second
			
			ns := New(cfg)
			ctx := context.Background()
			
			result := ns.ShouldRedirectToINetSim(ctx, tt.url)
			if result != tt.expected {
				t.Errorf("ShouldRedirectToINetSim() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetDNSServers(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected []string
	}{
		{
			name:     "simulation disabled - default DNS",
			enabled:  false,
			expected: []string{"8.8.8.8", "8.8.4.4"},
		},
		{
			name:     "simulation enabled - INetSim DNS",
			enabled:  true,
			expected: []string{"172.20.0.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Enabled = tt.enabled
			
			ns := New(cfg)
			result := ns.GetDNSServers()
			
			if len(result) != len(tt.expected) {
				t.Errorf("GetDNSServers() returned %d servers, expected %d", len(result), len(tt.expected))
				return
			}
			
			for i, dns := range result {
				if dns != tt.expected[i] {
					t.Errorf("GetDNSServers()[%d] = %s, expected %s", i, dns, tt.expected[i])
				}
			}
		})
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Enabled = tt.enabled
			
			ns := New(cfg)
			if ns.IsEnabled() != tt.enabled {
				t.Errorf("IsEnabled() = %v, expected %v", ns.IsEnabled(), tt.enabled)
			}
		})
	}
}

func TestGetINetSimAddresses(t *testing.T) {
	cfg := &Config{
		INetSimDNSAddr:  "192.168.1.100:53",
		INetSimHTTPAddr: "192.168.1.100:8080",
		Enabled:         true,
	}
	
	ns := New(cfg)
	
	if ns.GetINetSimDNS() != "192.168.1.100:53" {
		t.Errorf("GetINetSimDNS() = %s, expected 192.168.1.100:53", ns.GetINetSimDNS())
	}
	
	if ns.GetINetSimHTTP() != "192.168.1.100:8080" {
		t.Errorf("GetINetSimHTTP() = %s, expected 192.168.1.100:8080", ns.GetINetSimHTTP())
	}
}
