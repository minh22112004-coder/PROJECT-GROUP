"""
Demo script to showcase Pack-A-Mal Network Simulation capabilities.
This demonstrates how the network simulation detects dead URLs and redirects to INetSim.
"""

import os
import sys

# Add the networksim path (simulated - in real usage this would be Go code)
print("="*70)
print(" Pack-A-Mal Network Simulation - Demo")
print("="*70)
print()

# Simulate the configuration
class NetworkSimConfig:
    def __init__(self):
        self.enabled = os.getenv("OSSF_NETWORK_SIMULATION_ENABLED", "false").lower() == "true"
        self.inetsim_dns = os.getenv("OSSF_INETSIM_DNS_ADDR", "172.20.0.2:53")
        self.inetsim_http = os.getenv("OSSF_INETSIM_HTTP_ADDR", "172.20.0.2:80")
        
config = NetworkSimConfig()

print("Current Configuration:")
print(f"  Network Simulation Enabled: {config.enabled}")
print(f"  INetSim DNS Address:        {config.inetsim_dns}")
print(f"  INetSim HTTP Address:       {config.inetsim_http}")
print()

# Test URLs (from the sample package)
test_urls = [
    "http://malicious-c2-server.example.com/api/exfiltrate",
    "http://expired-malware-repo.net/payload.bin",
    "http://dead-phishing-site.org/credentials",
    "http://fake-analytics.com/track",
]

print("="*70)
print(" URL Liveness Detection (Simulated)")
print("="*70)
print()

import requests

for url in test_urls:
    print(f"Testing: {url}")
    try:
        response = requests.head(url, timeout=3, allow_redirects=False)
        is_alive = 200 <= response.status_code < 400
        print(f"  ✓ Status: {response.status_code}")
        print(f"  → Alive: {is_alive}")
        if config.enabled and not is_alive:
            print(f"  → Action: Would redirect to INetSim ({config.inetsim_http})")
        print()
    except requests.exceptions.RequestException as e:
        error_type = type(e).__name__
        print(f"  ✗ Error: {error_type}")
        print(f"  → Alive: False")
        if config.enabled:
            print(f"  → Action: Would redirect to INetSim ({config.inetsim_http})")
        else:
            print(f"  → Action: Connection would fail (simulation disabled)")
        print()

print("="*70)
print(" How Network Simulation Works")
print("="*70)
print("""
When OSSF_NETWORK_SIMULATION_ENABLED=true:

1. Before Analysis:
   - Worker validates INetSim connection
   - Configures sandbox with INetSim DNS (172.20.0.2)

2. During Package Analysis:
   - Package attempts DNS lookup → Goes to INetSim DNS
   - INetSim resolves ALL domains to 172.20.0.2
   - Package attempts HTTP request → Goes to INetSim HTTP
   - INetSim responds with simulated content
   - All traffic is logged

3. Results:
   - Network behavior captured without real C2 communication
   - Complete logs of malicious network activity
   - Safe analysis of dangerous packages

Example Flow:
  Package code:
    → requests.get("http://malicious-c2.example.com/steal-data")
  
  With simulation DISABLED:
    → DNS lookup fails (domain doesn't exist)
    → Connection error
    → No network analysis possible
  
  With simulation ENABLED:
    → DNS lookup to INetSim → resolves to 172.20.0.2
    → HTTP request to INetSim:80
    → INetSim logs the attempt and responds
    → Full network behavior captured! ✓
""")

print("="*70)
print(" To Enable Network Simulation")
print("="*70)
print("""
Windows (PowerShell):
  $env:OSSF_NETWORK_SIMULATION_ENABLED='true'
  $env:OSSF_INETSIM_DNS_ADDR='172.20.0.2:53'
  $env:OSSF_INETSIM_HTTP_ADDR='172.20.0.2:80'

Linux/Mac (Bash):
  export OSSF_NETWORK_SIMULATION_ENABLED=true
  export OSSF_INETSIM_DNS_ADDR=172.20.0.2:53
  export OSSF_INETSIM_HTTP_ADDR=172.20.0.2:80

Then run your Pack-A-Mal analysis as usual!
""")

print("="*70)
print(" INetSim Service Status")
print("="*70)
print()

import subprocess

try:
    # Check INetSim container
    result = subprocess.run(
        ["docker", "ps", "--filter", "name=inetsim", "--format", "{{.Status}}"],
        capture_output=True,
        text=True,
        timeout=5
    )
    
    if result.returncode == 0 and result.stdout.strip():
        print(f"  ✓ INetSim Container: {result.stdout.strip()}")
    else:
        print(f"  ✗ INetSim Container: Not running")
        print(f"    Run: docker-compose -f docker-compose.network-sim.yml up -d")
    
    # Test HTTP
    try:
        response = requests.get("http://localhost:8080/", timeout=3)
        print(f"  ✓ INetSim HTTP: Accessible (Status {response.status_code})")
    except:
        print(f"  ✗ INetSim HTTP: Not accessible")
    
except Exception as e:
    print(f"  ✗ Error checking INetSim: {e}")

print()
print("="*70)
print(" Demo Complete")
print("="*70)
print()
print("Next steps:")
print("  1. Start INetSim if not running")
print("  2. Enable network simulation environment variables")
print("  3. Run Pack-A-Mal analysis on a package")
print("  4. Check INetSim logs: docker logs inetsim")
print()
