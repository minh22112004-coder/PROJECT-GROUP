# Hướng Dẫn Chạy Network Simulation

## Bước 1: Khởi động Docker Services

```powershell
# Di chuyển vào thư mục dynamic-analysis
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis

# Khởi động INetSim và Service Simulation
docker-compose -f docker-compose.network-sim.yml up -d
```

**Kết quả mong đợi:**
```
✔ Network pack-a-mal-network    Created
✔ Container pack-a-mal-inetsim  Healthy  
✔ Container pack-a-mal-sim-api  Started (healthy)
```

## Bước 2: Kiểm tra Services đang chạy

```powershell
# Xem trạng thái containers
docker ps --filter "name=pack-a-mal"
```

**Kết quả mong đợi:** Cả 2 containers hiển thị status `(healthy)`

## Bước 3: Test INetSim HTTP Service

```powershell
# Test HTTP service (port 8080)
curl.exe http://localhost:8080
```

**Kết quả mong đợi:** Trả về trang HTML với nội dung "INetSim default HTML page"

## Bước 4: Test Service Simulation API

```powershell
# Test API status
curl.exe http://localhost:5000/status
```

**Kết quả mong đợi:**
```json
{"service":"simulation","status":"running"}
```

## Bước 5: Cấu hình Environment Variables

```powershell
# Set biến môi trường cho network simulation
$env:OSSF_NETWORK_SIMULATION_ENABLED = "true"
$env:OSSF_INETSIM_DNS_ADDR = "172.20.0.2:53"
$env:OSSF_INETSIM_HTTP_ADDR = "172.20.0.2:80"

# Kiểm tra
Write-Host "Network Simulation: $env:OSSF_NETWORK_SIMULATION_ENABLED"
Write-Host "DNS Server: $env:OSSF_INETSIM_DNS_ADDR"
Write-Host "HTTP Server: $env:OSSF_INETSIM_HTTP_ADDR"
```

## Bước 6: Chạy Go Unit Tests

```powershell
# Di chuyển vào thư mục networksim
cd internal\networksim

# Chạy tests
go test -v

# Kết quả: Tất cả tests phải PASS
```

**Kết quả thực tế (đã test):**
```
=== RUN   TestDefaultConfig
--- PASS: TestDefaultConfig (0.00s)

=== RUN   TestIsURLAlive
2026/01/28 18:34:20 INFO URL check url=http://127.0.0.1:65408 status=200 alive=true
2026/01/28 18:34:21 INFO URL not alive url=http://dead-url-12345.com
--- PASS: TestIsURLAlive (0.20s)

=== RUN   TestShouldRedirectToINetSim
2026/01/28 18:34:21 INFO URL check url=http://127.0.0.1:65410 status=200 alive=true
2026/01/28 18:34:21 INFO URL not alive url=http://dead-url.com
2026/01/28 18:34:21 INFO Redirecting to INetSim url=http://dead-url.com
--- PASS: TestShouldRedirectToINetSim (0.18s)

=== RUN   TestGetDNSServers
--- PASS: TestGetDNSServers (0.00s)

PASS
ok      github.com/ossf/package-analysis/internal/networksim    1.949s
```

✅ **Tất cả 4 tests PASS** - Logic kiểm tra URL và redirect hoạt động đúng!

## Bước 7: Test với Sample Malicious Package

### 7a. Test KHÔNG có INetSim (Chứng minh URL dead)

```powershell
# Quay lại dynamic-analysis
cd ..\..

# Di chuyển vào sample packages
cd sample_packages\malicious_network_package

# Cài đặt package
pip install -e .

# Chạy test cơ bản
python test_network.py
```

**Kết quả mong đợi:** 
```
============================================================
Malicious Network Package - Connecting to dead URL
============================================================

[*] Target URL: http://malicious-c2-server.example.com/api/data
[*] Attempting connection...
[-] Connection failed: HTTPConnectionPool(...): Max retries exceeded...
============================================================
```

👉 **Chứng minh:** URL không alive (dead URL) - Đáp ứng **Yêu cầu 1**

### 7b. Test CÓ INetSim (Chứng minh redirect thành công)

```powershell
# Chạy script demo redirect (đã tích hợp sẵn)
python test_with_inetsim.py
```

**Kết quả mong đợi:**
```
╔════════════════════════════════════════════════════════╗
║  Dead URL Redirect to INetSim - Demo Script          ║
║  Yêu cầu 2: Kiểm tra URL alive & redirect INetSim    ║
╚════════════════════════════════════════════════════════╝

============================================================
Testing Dead URL WITHOUT INetSim (Should Fail)
============================================================

[*] Target URL: http://malicious-c2-server.example.com/api/data
[*] No proxy - direct connection attempt

✓ Connection failed (as expected)
✓ This confirms the URL is indeed dead

------------------------------------------------------------

============================================================
Testing Dead URL Redirect to INetSim
============================================================

[*] INetSim Proxy: http://localhost:8080
[*] Testing dead URLs...

[*] Testing: http://malicious-c2-server.example.com/api/data
    ✓ Status: 200
    ✓ Connected via INetSim!
    ✓ Response confirmed from INetSim

[*] Testing: http://expired-malware-repo.net/payload.exe
    ✓ Status: 200
    ✓ Connected via INetSim!
    ✓ Response confirmed from INetSim

[*] Testing: http://dead-phishing-site.org/login
    ✓ Status: 200
    ✓ Connected via INetSim!
    ✓ Response confirmed from INetSim

============================================================
Summary: 3/3 URLs successfully redirected
============================================================

✓ All dead URLs successfully redirected to INetSim!
```

👉 **Chứng minh:** Dead URLs được redirect đến INetSim - Đáp ứng **Yêu cầu 2**

## Bước 8: Xem Logs

```powershell
# Xem logs của INetSim
docker logs pack-a-mal-inetsim

# Xem logs của Service Simulation
docker logs pack-a-mal-sim-api

# Xem logs file (nếu cần)
Get-Content "..\..\service-simulation-module\shared\logs\inetsim\service.log" -Tail 20
```

## Bước 9: Dừng Services (khi hoàn thành)

```powershell
# Quay lại dynamic-analysis
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis

# Dừng tất cả services
docker-compose -f docker-compose.network-sim.yml down
```


