# Network Mode Quick Start

## 🚀 5-Minute Quick Start

### 1️⃣ Import the Package

```go
import "github.com/ossf/package-analysis/internal/networkmode"
```

### 2️⃣ Create Controller (Full Mode - Safe Default)

```go
// Default Full Mode - Complete isolation
config := networkmode.DefaultConfig()
controller, err := networkmode.NewController(config, slog.Default())
if err != nil {
    log.Fatal(err)
}
defer controller.Close()
```

### 3️⃣ Handle Requests

```go
req := &networkmode.Request{
    ID:       "req-001",
    Protocol: string(networkmode.ProtocolHTTP),
    Domain:   "malware.com",
    Path:     "/api/data",
    Method:   "GET",
    Headers:  make(map[string]string),
}

ctx := context.Background()
resp, err := controller.HandleRequest(ctx, req)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Response Source: %s\n", resp.Source)  // "simulated"
fmt.Printf("Action: %s\n", resp.Decision.Action)  // "simulate"
```

---

## 🎯 Common Use Cases

### Use Case 1: Analyze Unknown Malware (Full Mode)

```go
// Full Mode - No external communication
config := networkmode.DefaultConfig()
controller, _ := networkmode.NewController(config, slog.Default())
defer controller.Close()

// All requests are simulated - 100% safe
resp, _ := controller.HandleRequest(ctx, malwareRequest)
// resp.Source == "simulated"
```

**✅ Perfect for:**
- Unknown/dangerous malware
- Production environments
- Environments without internet access
- Quick behavior analysis

---

### Use Case 2: Track C2 Infrastructure (Half Mode)

```go
// Half Mode - Controlled external access
config := networkmode.DefaultConfig()
config.Mode = networkmode.ModeHalf
config.HalfMode.Enabled = true
config.HalfMode.DefaultAction = networkmode.ActionSimulate

controller, _ := networkmode.NewController(config, slog.Default())
defer controller.Close()

// Add rule to forward to known C2
controller.AddDecisionRule(ctx, networkmode.DecisionRule{
    Name:     "forward_to_c2",
    Priority: 100,
    Action:   networkmode.ActionForward,
    Condition: &networkmode.RuleCondition{
        Type:    networkmode.ConditionDomainWhitelist,
        Domains: []string{"known-c2-server.com"},
    },
})

// C2 requests forwarded, others simulated
resp, _ := controller.HandleRequest(ctx, c2Request)
```

**✅ Perfect for:**
- C2 infrastructure research
- Collecting real payloads
- Understanding attack chains
- Threat intelligence gathering

---

### Use Case 3: Sandbox Executables

```go
// Automatically sandbox all executable downloads
config := networkmode.DefaultConfig()
config.Mode = networkmode.ModeHalf
config.HalfMode.Enabled = true
config.HalfMode.TrafficModifier.SandboxExecutables = true
config.HalfMode.TrafficModifier.SandboxDir = "/logs/executables"

controller, _ := networkmode.NewController(config, slog.Default())
defer controller.Close()

// Executable downloads are intercepted and sandboxed
resp, _ := controller.HandleRequest(ctx, exeDownloadRequest)
// Original saved to /logs/executables/
// Fake executable returned to malware
```

**✅ Perfect for:**
- Collecting malware samples
- Preventing actual infection
- Safe payload analysis

---

## 📊 Check Statistics

```go
stats := controller.GetStats()
fmt.Printf("Total: %d\n", stats.TotalRequests)
fmt.Printf("Forwarded: %d\n", stats.ForwardedRequests)
fmt.Printf("Blocked: %d\n", stats.BlockedRequests)
fmt.Printf("Simulated: %d\n", stats.SimulatedRequests)
```

---

## 🔄 Switch Modes at Runtime

```go
// Start in Full Mode
controller, _ := networkmode.NewController(config, slog.Default())

// Switch to Half Mode for specific analysis
controller.SwitchMode(ctx, networkmode.ModeHalf)

// Switch back to Full Mode
controller.SwitchMode(ctx, networkmode.ModeFull)
```

---

## 🛡️ Security Best Practices

### ✅ DO
- Use Full Mode for unknown malware
- Test Half Mode rules in isolated networks first
- Review decision logs regularly
- Enable executable sandboxing
- Set appropriate timeouts

### ❌ DON'T
- Run Half Mode in production without proper network isolation
- Forward all traffic to internet by default
- Disable logging
- Skip rule validation
- Ignore blocked requests

---

## 🎓 Decision Rules Quick Reference

### Block Malicious Domains
```go
networkmode.DecisionRule{
    Name:     "block_evil",
    Priority: 100,
    Action:   networkmode.ActionBlock,
    Condition: &networkmode.RuleCondition{
        Type:    networkmode.ConditionDomainBlacklist,
        Domains: []string{"evil.com", "*.malware.net"},
    },
}
```

### Allow CDNs
```go
networkmode.DecisionRule{
    Name:     "allow_cdns",
    Priority: 90,
    Action:   networkmode.ActionForward,
    Condition: &networkmode.RuleCondition{
        Type:    networkmode.ConditionDomainWhitelist,
        Domains: []string{"*.cloudflare.com"},
    },
}
```

### Sandbox Executables
```go
networkmode.DecisionRule{
    Name:     "sandbox_exe",
    Priority: 80,
    Action:   networkmode.ActionModify,
    Condition: &networkmode.RuleCondition{
        Type:           networkmode.ConditionFileExtension,
        FileExtensions: []string{".exe", ".dll"},
    },
    Modifier: &networkmode.Modifier{
        Type:         "sandbox_executable",
        SaveOriginal: true,
    },
}
```

### Monitor Uploads
```go
networkmode.DecisionRule{
    Name:     "monitor_uploads",
    Priority: 70,
    Action:   networkmode.ActionModify,
    Condition: &networkmode.RuleCondition{
        Type:    networkmode.ConditionUploadDetection,
        Method:  "POST",
        MinSize: 1024 * 1024, // 1MB
    },
    Modifier: &networkmode.Modifier{
        Type:           "content_logging",
        LogFullContent: true,
        StripPII:       true,
    },
}
```

---

## 📁 Configuration Files

### Quick Config (Full Mode)
```yaml
# config/network-mode.yaml
network_mode:
  mode: "full"
  
  full_mode:
    complete_isolation: true
    services:
      dns_address: "172.20.0.2:53"
      http_address: "172.20.0.3:80"
  
  logging:
    level: "info"
    capture_pcap: true
```

### Quick Config (Half Mode)
```yaml
# config/network-mode.yaml
network_mode:
  mode: "half"
  
  half_mode:
    enabled: true
    default_action: "simulate"
    
    whitelist:
      - "*.cloudflare.com"
    
    blacklist:
      - "*.evil.com"
    
    traffic_modifier:
      sandbox_executables: true
      sandbox_dir: "/logs/executables"
```

---

## 🧪 Test Your Setup

```bash
# Run example
cd dynamic-analysis/examples/networkmode
go run main.go

# Run tests
cd dynamic-analysis/internal/networkmode
go test -v
```

---

## 📚 More Information

- **Full Documentation:** [internal/networkmode/README.md](dynamic-analysis/internal/networkmode/README.md)
- **Design Document:** [docs/NETWORK_MODE_DESIGN.md](docs/NETWORK_MODE_DESIGN.md)
- **Implementation Summary:** [IMPLEMENTATION_SUMMARY_NETWORK_MODE.md](IMPLEMENTATION_SUMMARY_NETWORK_MODE.md)
- **Examples:** [examples/networkmode/main.go](dynamic-analysis/examples/networkmode/main.go)

---

## 🆘 Troubleshooting

### Q: How do I enable Half Mode?
```go
config.Mode = networkmode.ModeHalf
config.HalfMode.Enabled = true  // Must be explicitly enabled
```

### Q: How do I add custom rules?
```go
rule := networkmode.DecisionRule{
    Name:     "my_rule",
    Priority: 100,
    Action:   networkmode.ActionBlock,
    Condition: &networkmode.RuleCondition{
        Type: networkmode.ConditionDomainBlacklist,
        Domains: []string{"bad.com"},
    },
}
controller.AddDecisionRule(ctx, rule)
```

### Q: How do I see what's happening?
```go
// Enable verbose logging
config.Logging.Level = "debug"
config.Logging.LogAllRequests = true
config.Logging.LogDecisions = true

// Check statistics
stats := controller.GetStats()

// Review log files
// /logs/traffic.log
// /logs/decisions.log
```

### Q: Controller panics - what to do?
The controller has panic recovery and will fall back to Full Mode (safest). Check logs for details.

---

## ⚡ Performance Tips

1. **Enable decision caching** (enabled by default)
2. **Use appropriate log levels** (info in production, debug in development)
3. **Limit rule count** (< 50 rules recommended)
4. **Order rules by priority** (most specific first)
5. **Clear cache periodically** if rules change: `controller.ClearDecisionCache()`

---

**Ready to use! Start with Full Mode and graduate to Half Mode when needed.**
