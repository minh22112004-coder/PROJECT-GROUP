"""
Malicious Network Package - Sample package for testing network simulation.
This package attempts to connect to a URL that is no longer alive to demonstrate
network redirection to INetSim.
"""

import requests

# Dead URL that should be redirected to INetSim
DEAD_URL = "http://malicious-c2-server.example.com/api/data"

def connect_to_dead_url():
    """Attempt to connect to a dead URL"""
    print("="*60)
    print("Malicious Network Package - Connecting to dead URL")
    print("="*60)
    print(f"\n[*] Target URL: {DEAD_URL}")
    
    try:
        print(f"[*] Attempting connection...")
        response = requests.get(DEAD_URL, timeout=5)
        print(f"[+] Success! Status Code: {response.status_code}")
        print(f"[+] Response Content Length: {len(response.content)} bytes")
        print(f"[+] Response Preview: {response.text[:100]}")
    except requests.exceptions.ConnectionError as e:
        print(f"[-] Connection failed: {e}")
    except requests.exceptions.Timeout:
        print(f"[-] Request timed out")
    except Exception as e:
        print(f"[-] Error occurred: {e}")
    
    print("\n" + "="*60)
