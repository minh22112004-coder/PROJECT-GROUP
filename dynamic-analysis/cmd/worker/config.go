package main

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/ossf/package-analysis/internal/networksim"
	"github.com/ossf/package-analysis/internal/resultstore"
	"github.com/ossf/package-analysis/internal/worker"
)

// resultBucketPaths holds bucket paths for the different types of results.
type resultBucketPaths struct {
	analyzedPkg     string
	dynamicAnalysis string
	executionLog    string
	fileWrites      string
	staticAnalysis  string
}

type sandboxImageSpec struct {
	tag    string
	noPull bool
}

type config struct {
	imageSpec sandboxImageSpec

	resultStores *worker.ResultStores

	subURL               string
	packagesBucket       string
	notificationTopicURL string

	userAgentExtra string
	
	// Network simulation configuration
	networkSimConfig *networksim.Config
}

func (c *config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("subscription", c.subURL),
		slog.String("package_bucket", c.packagesBucket),
		slog.String("dynamic_results_store", c.resultStores.DynamicAnalysis.String()),
		slog.String("static_results_store", c.resultStores.StaticAnalysis.String()),
		slog.String("file_write_results_store", c.resultStores.FileWrites.String()),
		slog.String("analyzed_packages_store", c.resultStores.AnalyzedPackage.String()),
		slog.String("execution_log_store", c.resultStores.ExecutionLog.String()),
		slog.String("image_tag", c.imageSpec.tag),
		slog.Bool("image_nopull", c.imageSpec.noPull),
		slog.String("topic_notification", c.notificationTopicURL),
		slog.String("user_agent_extra", c.userAgentExtra),
		slog.Bool("network_sim_enabled", c.networkSimConfig.Enabled),
		slog.String("network_sim_dns", c.networkSimConfig.INetSimDNSAddr),
		slog.String("network_sim_http", c.networkSimConfig.INetSimHTTPAddr),
	)
}

func resultStoreForEnv(key string) *resultstore.ResultStore {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}
	return resultstore.New(val, resultstore.ConstructPath())
}// Parse network simulation configuration from environment
	netSimConfig := networksim.DefaultConfig()
	
	if os.Getenv("OSSF_NETWORK_SIMULATION_ENABLED") == "true" {
		netSimConfig.Enabled = true
	}
	
	if dnsAddr := os.Getenv("OSSF_INETSIM_DNS_ADDR"); dnsAddr != "" {
		netSimConfig.INetSimDNSAddr = dnsAddr
	}
	
	if httpAddr := os.Getenv("OSSF_INETSIM_HTTP_ADDR"); httpAddr != "" {
		netSimConfig.INetSimHTTPAddr = httpAddr
	}
	
	if timeout := os.Getenv("OSSF_URL_LIVENESS_TIMEOUT"); timeout != "" {
		if seconds, err := strconv.Atoi(timeout); err == nil {
			netSimConfig.LivenessTimeout = time.Duration(seconds) * time.Second
		}
	}
	
	return &config{
		imageSpec: sandboxImageSpec{
			tag:    os.Getenv("OSSF_SANDBOX_IMAGE_TAG"),
			noPull: os.Getenv("OSSF_SANDBOX_NOPULL") != "",
		},
		resultStores: &worker.ResultStores{
			AnalyzedPackage: resultStoreForEnv("OSSF_MALWARE_ANALYZED_PACKAGES"),
			DynamicAnalysis: resultStoreForEnv("OSSF_MALWARE_ANALYSIS_RESULTS"),
			ExecutionLog:    resultStoreForEnv("OSSF_MALWARE_ANALYSIS_EXECUTION_LOGS"),
			FileWrites:      resultStoreForEnv("OSSF_MALWARE_ANALYSIS_FILE_WRITE_RESULTS"),
			StaticAnalysis:  resultStoreForEnv("OSSF_MALWARE_STATIC_ANALYSIS_RESULTS"),
		},
		subURL:               os.Getenv("OSSMALWARE_WORKER_SUBSCRIPTION"),
		packagesBucket:       os.Getenv("OSSF_MALWARE_ANALYSIS_PACKAGES"),
		notificationTopicURL: os.Getenv("OSSF_MALWARE_NOTIFICATION_TOPIC"),
		userAgentExtra:       os.Getenv("OSSF_MALWARE_USER_AGENT_EXTRA"),
		networkSimConfig:     netSimConfigKAGES"),
		notificationTopicURL: os.Getenv("OSSF_MALWARE_NOTIFICATION_TOPIC"),

		userAgentExtra: os.Getenv("OSSF_MALWARE_USER_AGENT_EXTRA"),
	}
}
