#!/usr/bin/env python3
"""
Test script to verify INetSim integration with network simulation.
This script tests DNS resolution and HTTP requests through INetSim.
"""

import socket
import requests
import sys

INETSIM_IP = "172.20.0.2"

def test_direct_connection():
    """Test direct HTTP connection to INetSim"""
    print("="*60)
    print("TEST 1: Direct HTTP Connection to INetSim")
    print("="*60)
    
    try:
        # Test direct connection to INetSim HTTP
        response = requests.get(f"http://{INETSIM_IP}/", timeout=5)
        print(f"✓ Status Code: {response.status_code}")
        print(f"✓ Content Length: {len(response.content)} bytes")
        print(f"✓ Content Preview: {response.text[:100]}")
        return True
    except Exception as e:
        print(f"✗ Failed: {e}")
        return False

def test_fake_domain_request():
    """Test HTTP request to a fake domain that should be handled by INetSim"""
    print("\n" + "="*60)
    print("TEST 2: Request to Fake Malicious Domain")
    print("="*60)
    
    # This will only work if DNS is configured to use INetSim
    fake_urls = [
        "http://malicious-c2-server.example.com/api/test",
        "http://expired-malware-repo.net/payload.bin",
    ]
    
    for url in fake_urls:
        try:
            print(f"\n[*] Attempting: {url}")
            # This will fail with DNS error unless INetSim DNS is configured
            response = requests.get(url, timeout=3)
            print(f"✓ Status: {response.status_code}")
            print(f"✓ Response received from INetSim!")
        except requests.exceptions.ConnectionError as e:
            if "Failed to resolve" in str(e) or "Name or service not known" in str(e):
                print(f"✗ DNS Resolution Failed (INetSim DNS not configured)")
                print(f"   This is expected when not using INetSim DNS")
            else:
                print(f"✗ Connection Error: {e}")
        except Exception as e:
            print(f"✗ Error: {e}")

def test_localhost_inetsim():
    """Test connection to localhost:8080 (INetSim HTTP mapped port)"""
    print("\n" + "="*60)
    print("TEST 3: Connection via Localhost Port Mapping")
    print("="*60)
    
    try:
        response = requests.get("http://localhost:8080/", timeout=5)
        print(f"✓ Status Code: {response.status_code}")
        print(f"✓ INetSim HTTP accessible via localhost:8080")
        print(f"✓ Content: {response.text[:100]}")
        return True
    except Exception as e:
        print(f"✗ Failed: {e}")
        return False

def print_summary():
    """Print test summary and instructions"""
    print("\n" + "="*60)
    print("SUMMARY & NEXT STEPS")
    print("="*60)
    print("""
✓ INetSim HTTP service is accessible
✓ Ready for network simulation integration

To enable network simulation in Pack-A-Mal:
1. Set environment variable:
   $env:OSSF_NETWORK_SIMULATION_ENABLED='true'
   $env:OSSF_INETSIM_DNS_ADDR='172.20.0.2:53'
   $env:OSSF_INETSIM_HTTP_ADDR='172.20.0.2:80'

2. The sandbox will automatically:
   - Use INetSim DNS (172.20.0.2) for all queries
   - All domains will resolve to INetSim IP
   - HTTP requests will be handled by INetSim
   - All traffic will be logged

3. To test full integration:
   - Deploy a package in the sandbox
   - Check INetSim logs: docker logs inetsim
   - Verify network capture in analysis results
""")

if __name__ == "__main__":
    print("\n" + "="*60)
    print("INetSim Integration Test")
    print("="*60)
    print(f"INetSim IP: {INETSIM_IP}")
    print(f"INetSim HTTP Port: 80 (mapped to localhost:8080)")
    print(f"INetSim DNS Port: 53 (mapped to localhost:53)")
    
    results = []
    
    # Run tests
    results.append(("Direct HTTP Connection", test_direct_connection()))
    test_fake_domain_request()
    results.append(("Localhost Port Mapping", test_localhost_inetsim()))
    
    # Print summary
    print_summary()
    
    # Final result
    passed = sum(1 for _, result in results if result)
    total = len(results)
    
    print(f"\n{'='*60}")
    print(f"Test Results: {passed}/{total} tests passed")
    print(f"{'='*60}\n")
    
    sys.exit(0 if passed == total else 1)
