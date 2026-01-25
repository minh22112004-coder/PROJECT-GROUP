# Malicious Network Package

Sample package for testing Pack-A-Mal's network simulation integration with INetSim.

## Purpose

This package simulates malicious network behavior by attempting to:
1. Resolve DNS for dead/fake malicious domains
2. Make HTTP requests to non-existent C2 servers
3. Attempt data exfiltration
4. Download fake malware payloads

## Dead URLs Used

- `http://malicious-c2-server.example.com/api/exfiltrate` - Fake C2 server
- `http://expired-malware-repo.net/payload.bin` - Fake malware repository
- `http://dead-phishing-site.org/credentials` - Fake phishing site
- `http://fake-analytics.com/track` - Fake tracking endpoint

## Expected Behavior

### Without INetSim Integration
- All DNS resolutions will fail
- All HTTP requests will timeout or fail
- No network activity will be captured

### With INetSim Integration
- DNS queries are intercepted and redirected to INetSim (172.20.0.2)
- All domains resolve to INetSim's IP
- HTTP requests receive simulated responses from INetSim
- Network activity is logged and analyzed

## Usage for Testing

```python
import malicious_network_package

# Run all network tests
malicious_network_package.run_all_checks()

# Or run individual tests
malicious_network_package.check_network_connectivity()
malicious_network_package.attempt_http_requests()
malicious_network_package.exfiltrate_data()
malicious_network_package.download_payload()
```

## Installation

```bash
cd sample_packages/malicious_network_package
pip install -e .
```

## Testing with Pack-A-Mal

```bash
# Build the package
cd sample_packages/malicious_network_package
python setup.py sdist

# Analyze with Pack-A-Mal
# (Instructions will be added after integration)
```
