package networksim

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Config holds the network simulation configuration
type Config struct {
	// INetSimDNSAddr is the DNS server address (e.g., "172.20.0.2:53")
	INetSimDNSAddr string
	// INetSimHTTPAddr is the HTTP server address (e.g., "172.20.0.2:80")
	INetSimHTTPAddr string
	// Enabled indicates if network simulation is enabled
	Enabled bool
	// LivenessTimeout is the timeout for checking URL liveness
	LivenessTimeout time.Duration
}

// DefaultConfig returns the default configuration for INetSim integration
func DefaultConfig() *Config {
	return &Config{
		INetSimDNSAddr:  "172.20.0.2:53",
		INetSimHTTPAddr: "172.20.0.2:80",
		Enabled:         false,
		LivenessTimeout: 3 * time.Second,
	}
}

// NetworkSimulator handles network simulation and redirection
type NetworkSimulator struct {
	config *Config
}

// New creates a new NetworkSimulator instance
func New(config *Config) *NetworkSimulator {
	if config == nil {
		config = DefaultConfig()
	}
	return &NetworkSimulator{
		config: config,
	}
}

// IsURLAlive checks if a given URL is accessible
func (ns *NetworkSimulator) IsURLAlive(ctx context.Context, url string) bool {
	if !ns.config.Enabled {
		// If simulation is disabled, assume URLs are alive
		return true
	}

	slog.InfoContext(ctx, "Checking URL liveness", "url", url)

	client := &http.Client{
		Timeout: ns.config.LivenessTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects for liveness check
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		slog.WarnContext(ctx, "Failed to create request for URL liveness check",
			"url", url,
			"error", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.InfoContext(ctx, "URL is not alive",
			"url", url,
			"error", err)
		return false
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx status codes as "alive"
	isAlive := resp.StatusCode >= 200 && resp.StatusCode < 400
	slog.InfoContext(ctx, "URL liveness check result",
		"url", url,
		"status_code", resp.StatusCode,
		"is_alive", isAlive)

	return isAlive
}

// IsHostAlive checks if a host is resolvable via DNS
func (ns *NetworkSimulator) IsHostAlive(ctx context.Context, host string) bool {
	if !ns.config.Enabled {
		return true
	}

	slog.InfoContext(ctx, "Checking host DNS resolution", "host", host)

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: ns.config.LivenessTimeout,
			}
			return d.DialContext(ctx, network, address)
		},
	}

	ips, err := resolver.LookupHost(ctx, host)
	if err != nil {
		slog.InfoContext(ctx, "Host DNS resolution failed",
			"host", host,
			"error", err)
		return false
	}

	slog.InfoContext(ctx, "Host resolved successfully",
		"host", host,
		"ips", ips)
	return true
}

// ShouldRedirectToINetSim determines if network traffic should be redirected to INetSim
func (ns *NetworkSimulator) ShouldRedirectToINetSim(ctx context.Context, url string) bool {
	if !ns.config.Enabled {
		return false
	}

	isAlive := ns.IsURLAlive(ctx, url)
	shouldRedirect := !isAlive

	if shouldRedirect {
		slog.InfoContext(ctx, "URL will be redirected to INetSim",
			"url", url,
			"inetsim_http", ns.config.INetSimHTTPAddr)
	}

	return shouldRedirect
}

// GetINetSimDNS returns the INetSim DNS server address
func (ns *NetworkSimulator) GetINetSimDNS() string {
	return ns.config.INetSimDNSAddr
}

// GetINetSimHTTP returns the INetSim HTTP server address
func (ns *NetworkSimulator) GetINetSimHTTP() string {
	return ns.config.INetSimHTTPAddr
}

// IsEnabled returns whether network simulation is enabled
func (ns *NetworkSimulator) IsEnabled() bool {
	return ns.config.Enabled
}

// GetDNSServers returns DNS servers to use for the sandbox
// If simulation is enabled and should redirect, returns INetSim DNS
// Otherwise returns default DNS servers
func (ns *NetworkSimulator) GetDNSServers() []string {
	if !ns.config.Enabled {
		return []string{"8.8.8.8", "8.8.4.4"} // Default Google DNS
	}

	// When simulation is enabled, use INetSim DNS
	host, _, err := net.SplitHostPort(ns.config.INetSimDNSAddr)
	if err != nil {
		// If parsing fails, return the full address
		return []string{ns.config.INetSimDNSAddr}
	}
	return []string{host}
}

// Stats holds statistics about network simulation
type Stats struct {
	URLsChecked      int
	URLsAlive        int
	URLsRedirected   int
	HostsChecked     int
	HostsResolved    int
}

// GetStats returns statistics (placeholder for future implementation)
func (ns *NetworkSimulator) GetStats() *Stats {
	return &Stats{}
}

// ValidateINetSimConnection checks if INetSim is accessible
func (ns *NetworkSimulator) ValidateINetSimConnection(ctx context.Context) error {
	if !ns.config.Enabled {
		return nil
	}

	slog.InfoContext(ctx, "Validating INetSim connection")

	// Check DNS connectivity
	dnsHost, dnsPort, err := net.SplitHostPort(ns.config.INetSimDNSAddr)
	if err != nil {
		return fmt.Errorf("invalid INetSim DNS address: %w", err)
	}

	conn, err := net.DialTimeout("udp", net.JoinHostPort(dnsHost, dnsPort), ns.config.LivenessTimeout)
	if err != nil {
		return fmt.Errorf("failed to connect to INetSim DNS: %w", err)
	}
	conn.Close()

	// Check HTTP connectivity
	httpURL := fmt.Sprintf("http://%s/", ns.config.INetSimHTTPAddr)
	client := &http.Client{
		Timeout: ns.config.LivenessTimeout,
	}

	resp, err := client.Get(httpURL)
	if err != nil {
		return fmt.Errorf("failed to connect to INetSim HTTP: %w", err)
	}
	defer resp.Body.Close()

	slog.InfoContext(ctx, "INetSim connection validated successfully",
		"dns", ns.config.INetSimDNSAddr,
		"http", ns.config.INetSimHTTPAddr)

	return nil
}
