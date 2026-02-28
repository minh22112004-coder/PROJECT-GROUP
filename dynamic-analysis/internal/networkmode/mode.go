package networkmode

import (
	"time"
)

// Mode represents the network operation mode
type Mode string

const (
	// ModeFull - Complete isolation, no external traffic allowed
	ModeFull Mode = "full"
	// ModeHalf - Transparent proxy with selective forwarding
	ModeHalf Mode = "half"
	// ModeTransparent - Pass-through monitoring only, no traffic alteration
	// Inspired by siemens/sparring transparent mode
	ModeTransparent Mode = "transparent"
)

// String returns string representation of Mode
func (m Mode) String() string {
	return string(m)
}

// IsValid checks if mode is valid
func (m Mode) IsValid() bool {
	return m == ModeFull || m == ModeHalf || m == ModeTransparent
}

// Config holds the network mode configuration
type Config struct {
	// Mode - Current network mode (full or half)
	Mode Mode `json:"mode" yaml:"mode"`

	// FullModeConfig - Configuration for Full Mode
	FullMode *FullModeConfig `json:"full_mode,omitempty" yaml:"full_mode,omitempty"`

	// HalfModeConfig - Configuration for Half Mode
	HalfMode *HalfModeConfig `json:"half_mode,omitempty" yaml:"half_mode,omitempty"`

	// TransparentModeConfig - Configuration for Transparent Mode
	TransparentMode *TransparentModeConfig `json:"transparent_mode,omitempty" yaml:"transparent_mode,omitempty"`

	// Logging configuration
	Logging *LoggingConfig `json:"logging" yaml:"logging"`
}

// FullModeConfig holds Full Mode specific settings
type FullModeConfig struct {
	// CompleteIsolation - Block ALL external traffic
	CompleteIsolation bool `json:"complete_isolation" yaml:"complete_isolation"`

	// Services configuration
	Services *ServiceConfig `json:"services" yaml:"services"`

	// CapturePCAP - Enable PCAP capture
	CapturePCAP bool `json:"capture_pcap" yaml:"capture_pcap"`
}

// ServiceConfig defines which service handles which protocol
type ServiceConfig struct {
	// DNS handler (inetsim, fakenet-ng, custom)
	DNS string `json:"dns" yaml:"dns"`

	// DNSAddress - Address of DNS service
	DNSAddress string `json:"dns_address" yaml:"dns_address"`

	// HTTP handler
	HTTP string `json:"http" yaml:"http"`

	// HTTPAddress - Address of HTTP service
	HTTPAddress string `json:"http_address" yaml:"http_address"`

	// HTTPS handler
	HTTPS string `json:"https" yaml:"https"`

	// HTTPSAddress - Address of HTTPS service
	HTTPSAddress string `json:"https_address" yaml:"https_address"`

	// SMTP handler
	SMTP string `json:"smtp" yaml:"smtp"`

	// SMTPAddress - Address of SMTP service
	SMTPAddress string `json:"smtp_address" yaml:"smtp_address"`

	// FTP handler
	FTP string `json:"ftp" yaml:"ftp"`

	// FTPAddress - Address of FTP service
	FTPAddress string `json:"ftp_address" yaml:"ftp_address"`
}

// TransparentModeConfig holds Transparent Mode specific settings
// In transparent mode, traffic is NOT altered - only logged and analyzed
// This mirrors siemens/sparring's TRANSPARENT mode behavior
type TransparentModeConfig struct {
	// Enabled - Must be explicitly enabled
	Enabled bool `json:"enabled" yaml:"enabled"`

	// ExtractPayloads - Extract and log protocol payloads
	ExtractPayloads bool `json:"extract_payloads" yaml:"extract_payloads"`

	// LogConnections - Log all TCP/UDP connections
	LogConnections bool `json:"log_connections" yaml:"log_connections"`

	// LogICMP - Log ICMP traffic
	LogICMP bool `json:"log_icmp" yaml:"log_icmp"`

	// SupportedProtocols - Protocols to extract data from (http, dns, smtp, ftp)
	SupportedProtocols []string `json:"supported_protocols" yaml:"supported_protocols"`

	// ConnectionLogFile - Path to connection log file
	ConnectionLogFile string `json:"connection_log_file" yaml:"connection_log_file"`

	// PayloadLogFile - Path to payload extraction log file
	PayloadLogFile string `json:"payload_log_file" yaml:"payload_log_file"`

	// MaxPayloadSize - Maximum payload size to capture (bytes)
	MaxPayloadSize int64 `json:"max_payload_size" yaml:"max_payload_size"`
}

// HalfModeConfig holds Half Mode specific settings
type HalfModeConfig struct {
	// Enabled - Must be explicitly enabled for safety
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Proxy configuration
	Proxy *ProxyConfig `json:"proxy" yaml:"proxy"`

	// DecisionRulesFile - Path to decision rules file
	DecisionRulesFile string `json:"decision_rules_file" yaml:"decision_rules_file"`

	// DefaultAction - Default action when no rule matches
	DefaultAction Action `json:"default_action" yaml:"default_action"`

	// Whitelist - Domains to always forward
	Whitelist []string `json:"whitelist,omitempty" yaml:"whitelist,omitempty"`

	// Blacklist - Domains to always block
	Blacklist []string `json:"blacklist,omitempty" yaml:"blacklist,omitempty"`

	// TrafficModifier configuration
	TrafficModifier *TrafficModifierConfig `json:"traffic_modifier" yaml:"traffic_modifier"`

	// Timeout for external requests
	ExternalRequestTimeout time.Duration `json:"external_request_timeout" yaml:"external_request_timeout"`
}

// ProxyConfig holds proxy settings
type ProxyConfig struct {
	// Transparent - Use transparent proxy
	Transparent bool `json:"transparent" yaml:"transparent"`

	// ListenAddress - Proxy listen address
	ListenAddress string `json:"listen_address" yaml:"listen_address"`

	// DNSInterception - Intercept DNS queries
	DNSInterception bool `json:"dns_interception" yaml:"dns_interception"`

	// SSLInterception - Intercept HTTPS traffic
	SSLInterception bool `json:"ssl_interception" yaml:"ssl_interception"`

	// SSLCertPath - Path to SSL certificate for MITM
	SSLCertPath string `json:"ssl_cert_path" yaml:"ssl_cert_path"`

	// SSLKeyPath - Path to SSL private key
	SSLKeyPath string `json:"ssl_key_path" yaml:"ssl_key_path"`
}

// TrafficModifierConfig holds traffic modification settings
type TrafficModifierConfig struct {
	// Enabled - Enable traffic modification
	Enabled bool `json:"enabled" yaml:"enabled"`

	// StripAuthHeaders - Remove authentication headers
	StripAuthHeaders bool `json:"strip_auth_headers" yaml:"strip_auth_headers"`

	// InjectTrackingHeaders - Add tracking headers
	InjectTrackingHeaders bool `json:"inject_tracking_headers" yaml:"inject_tracking_headers"`

	// SandboxExecutables - Sandbox executable downloads
	SandboxExecutables bool `json:"sandbox_executables" yaml:"sandbox_executables"`

	// SandboxDir - Directory to store sandboxed files
	SandboxDir string `json:"sandbox_dir" yaml:"sandbox_dir"`

	// MaxResponseSize - Maximum response size in bytes
	MaxResponseSize int64 `json:"max_response_size" yaml:"max_response_size"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	// Level - Log level (debug, info, warn, error)
	Level string `json:"level" yaml:"level"`

	// CapturePCAP - Enable PCAP capture
	CapturePCAP bool `json:"capture_pcap" yaml:"capture_pcap"`

	// PCAPFile - Path to PCAP file
	PCAPFile string `json:"pcap_file" yaml:"pcap_file"`

	// LogAllRequests - Log all requests
	LogAllRequests bool `json:"log_all_requests" yaml:"log_all_requests"`

	// LogResponses - Log responses
	LogResponses bool `json:"log_responses" yaml:"log_responses"`

	// LogDecisions - Log decision engine decisions
	LogDecisions bool `json:"log_decisions" yaml:"log_decisions"`

	// LogModifications - Log traffic modifications
	LogModifications bool `json:"log_modifications" yaml:"log_modifications"`

	// DecisionsFile - Path to decisions log file
	DecisionsFile string `json:"decisions_file" yaml:"decisions_file"`

	// TrafficLogFile - Path to traffic log file
	TrafficLogFile string `json:"traffic_log_file" yaml:"traffic_log_file"`
}

// DefaultConfig returns a safe default configuration
func DefaultConfig() *Config {
	return &Config{
		Mode: ModeFull, // Default to Full Mode for safety
		TransparentMode: &TransparentModeConfig{
			Enabled:            false,
			ExtractPayloads:    true,
			LogConnections:     true,
			LogICMP:            true,
			SupportedProtocols: []string{"http", "dns", "smtp", "ftp"},
			ConnectionLogFile:  "/logs/transparent_connections.log",
			PayloadLogFile:     "/logs/transparent_payloads.log",
			MaxPayloadSize:     1 * 1024 * 1024, // 1MB
		},
		FullMode: &FullModeConfig{
			CompleteIsolation: true,
			Services: &ServiceConfig{
				DNS:          "inetsim",
				DNSAddress:   "172.20.0.2:53",
				HTTP:         "fakenet-ng",
				HTTPAddress:  "172.20.0.3:80",
				HTTPS:        "fakenet-ng",
				HTTPSAddress: "172.20.0.3:443",
				SMTP:         "inetsim",
				SMTPAddress:  "172.20.0.2:25",
				FTP:          "inetsim",
				FTPAddress:   "172.20.0.2:21",
			},
			CapturePCAP: true,
		},
		HalfMode: &HalfModeConfig{
			Enabled:                false, // Disabled by default for safety
			DefaultAction:          ActionSimulate,
			ExternalRequestTimeout: 10 * time.Second,
			Proxy: &ProxyConfig{
				Transparent:     true,
				ListenAddress:   "0.0.0.0:8888",
				DNSInterception: true,
				SSLInterception: true,
			},
			TrafficModifier: &TrafficModifierConfig{
				Enabled:            true,
				StripAuthHeaders:   true,
				SandboxExecutables: true,
				SandboxDir:         "/logs/executables",
				MaxResponseSize:    10 * 1024 * 1024, // 10MB
			},
		},
		Logging: &LoggingConfig{
			Level:            "info",
			CapturePCAP:      true,
			PCAPFile:         "/logs/traffic.pcap",
			LogAllRequests:   true,
			LogResponses:     true,
			LogDecisions:     true,
			LogModifications: true,
			DecisionsFile:    "/logs/decisions.log",
			TrafficLogFile:   "/logs/traffic.log",
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if !c.Mode.IsValid() {
		return ErrInvalidMode
	}

	// Full Mode validation
	if c.Mode == ModeFull {
		if c.FullMode == nil {
			return ErrMissingConfig
		}
		if c.FullMode.Services == nil {
			return ErrMissingServiceConfig
		}
	}

	// Half Mode validation
	if c.Mode == ModeHalf {
		if c.HalfMode == nil {
			return ErrMissingConfig
		}
		if !c.HalfMode.Enabled {
			return ErrHalfModeNotEnabled
		}
		if c.HalfMode.Proxy == nil {
			return ErrMissingProxyConfig
		}
		if !c.HalfMode.DefaultAction.IsValid() {
			return ErrInvalidAction
		}
	}

	// Transparent Mode validation
	if c.Mode == ModeTransparent {
		if c.TransparentMode == nil {
			return ErrMissingConfig
		}
		if !c.TransparentMode.Enabled {
			return ErrTransparentModeNotEnabled
		}
	}

	return nil
}
