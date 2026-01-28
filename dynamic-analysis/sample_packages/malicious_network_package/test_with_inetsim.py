#!/usr/bin/env python3
"""
Test script to demonstrate dead URL redirection to INetSim.
This script shows how the package connects to dead URLs through INetSim.
"""

import requests
import sys

def test_dead_url_with_inetsim():
    """Test connecting to dead URL through INetSim proxy"""
    
    print('=' * 60)
    print('Testing Dead URL Redirect to INetSim')
    print('=' * 60)
    print()
    
    # Dead URLs that should be redirected
    dead_urls = [
        'http://malicious-c2-server.example.com/api/data',
        'http://expired-malware-repo.net/payload.exe',
        'http://dead-phishing-site.org/login'
    ]
    
    # INetSim proxy configuration
    proxies = {
        'http': 'http://localhost:8080',
        'https': 'http://localhost:8080'
    }
    
    print('[*] INetSim Proxy: http://localhost:8080')
    print('[*] Testing dead URLs...')
    print()
    
    success_count = 0
    
    for url in dead_urls:
        print(f'[*] Testing: {url}')
        
        try:
            response = requests.get(url, proxies=proxies, timeout=5)
            
            if response.status_code == 200:
                print(f'    ✓ Status: {response.status_code}')
                print(f'    ✓ Connected via INetSim!')
                
                # Check if response is from INetSim
                if 'INetSim' in response.text or 'default HTML page' in response.text:
                    print(f'    ✓ Response confirmed from INetSim')
                    success_count += 1
                else:
                    print(f'    ✓ Response received (length: {len(response.text)} bytes)')
                    success_count += 1
            else:
                print(f'    ⚠ Status: {response.status_code}')
                
        except requests.exceptions.ConnectionError as e:
            print(f'    ✗ Connection failed: {str(e)[:100]}')
            print(f'    ℹ Make sure INetSim is running on localhost:8080')
        except requests.exceptions.Timeout:
            print(f'    ✗ Request timed out')
        except Exception as e:
            print(f'    ✗ Error: {str(e)[:100]}')
        
        print()
    
    print('=' * 60)
    print(f'Summary: {success_count}/{len(dead_urls)} URLs successfully redirected')
    print('=' * 60)
    
    if success_count == len(dead_urls):
        print('\n✓ All dead URLs successfully redirected to INetSim!')
        return 0
    elif success_count > 0:
        print(f'\n⚠ Partially successful: {success_count}/{len(dead_urls)} URLs redirected')
        return 1
    else:
        print('\n✗ No URLs were redirected. Check if INetSim is running.')
        return 2


def test_without_proxy():
    """Test connecting to dead URL WITHOUT INetSim (should fail)"""
    
    print('=' * 60)
    print('Testing Dead URL WITHOUT INetSim (Should Fail)')
    print('=' * 60)
    print()
    
    url = 'http://malicious-c2-server.example.com/api/data'
    
    print(f'[*] Target URL: {url}')
    print('[*] No proxy - direct connection attempt')
    print()
    
    try:
        response = requests.get(url, timeout=5)
        print(f'[!] Unexpected success: {response.status_code}')
        print('[!] This URL should not be reachable!')
    except requests.exceptions.ConnectionError:
        print('✓ Connection failed (as expected)')
        print('✓ This confirms the URL is indeed dead')
    except requests.exceptions.Timeout:
        print('✓ Request timed out (as expected)')
        print('✓ This confirms the URL is not reachable')
    except Exception as e:
        print(f'✓ Error occurred: {type(e).__name__}')
        print(f'✓ This confirms the URL is not accessible')
    
    print()


if __name__ == '__main__':
    print()
    print('╔════════════════════════════════════════════════════════╗')
    print('║  Dead URL Redirect to INetSim - Demo Script          ║')
    print('║  Yêu cầu 2: Kiểm tra URL alive & redirect INetSim    ║')
    print('╚════════════════════════════════════════════════════════╝')
    print()
    
    # Test 1: Without INetSim (should fail)
    test_without_proxy()
    
    print()
    print('-' * 60)
    print()
    
    # Test 2: With INetSim (should succeed)
    exit_code = test_dead_url_with_inetsim()
    
    print()
    sys.exit(exit_code)
