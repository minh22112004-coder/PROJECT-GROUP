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
✔ Container pack-a-mal-sim-api  Started
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

**Kết quả mong đợi:**
```
PASS: TestNew
PASS: TestDefaultConfig
PASS: TestIsURLAlive
PASS: TestShouldRedirectToINetSim
PASS: TestGetDNSServers
... (tổng 20 tests)
ok      github.com/.../networksim
```

## Bước 7: Test với Sample Malicious Package

```powershell
# Quay lại dynamic-analysis
cd ..\..

# Di chuyển vào sample package
cd sample_packages\malicious_network_package

# Cài đặt dependencies
pip install requests

# Chạy test
python test_network.py
```

**Kết quả mong đợi:** Package sẽ thử kết nối tới các dead URLs và thất bại (đây là behavior mong muốn khi chưa có DNS redirection)

## Bước 8: Xem Logs

```powershell
# Xem logs của INetSim
docker logs pack-a-mal-inetsim

# Xem logs của Service Simulation
docker logs pack-a-mal-sim-api

# Xem logs file (nếu cần)
Get-Content ..\service-simulation-module\shared\logs\inetsim\service.log -Tail 20
```

## Bước 9: Dừng Services (khi hoàn thành)

```powershell
# Quay lại dynamic-analysis
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis

# Dừng tất cả services
docker-compose -f docker-compose.network-sim.yml down
```

---

## Script Nhanh (All-in-One)

Tạo file `quick-test.ps1`:

```powershell
# Quick Test Script
Write-Host "Starting Network Simulation Test..." -ForegroundColor Cyan

# 1. Start Docker
Write-Host "`n[1/4] Starting Docker services..." -ForegroundColor Yellow
docker-compose -f docker-compose.network-sim.yml up -d

# 2. Wait and check
Write-Host "`n[2/4] Waiting for services..." -ForegroundColor Yellow
Start-Sleep -Seconds 10
docker ps --filter "name=pack-a-mal" --format "{{.Names}}: {{.Status}}"

# 3. Test HTTP
Write-Host "`n[3/4] Testing HTTP service..." -ForegroundColor Yellow
$response = curl.exe -s http://localhost:8080
if ($response -match "INetSim") {
    Write-Host "[OK] INetSim HTTP working" -ForegroundColor Green
} else {
    Write-Host "[ERROR] INetSim HTTP not responding" -ForegroundColor Red
}

# 4. Test API
Write-Host "`n[4/4] Testing API..." -ForegroundColor Yellow
$apiResponse = curl.exe -s http://localhost:5000/status
Write-Host "API Response: $apiResponse" -ForegroundColor Green

Write-Host "`nTest completed!" -ForegroundColor Cyan
```

Sau đó chạy:
```powershell
.\quick-test.ps1
```

---

## Troubleshooting

### Lỗi: "Network overlaps with other one"
```powershell
# Xóa network cũ
docker network ls
docker network rm service-simulation-module_simulation_network
# Chạy lại
docker-compose -f docker-compose.network-sim.yml up -d
```

### Lỗi: Container unhealthy
```powershell
# Xem logs để debug
docker logs pack-a-mal-inetsim
# Restart
docker-compose -f docker-compose.network-sim.yml restart
```

### Lỗi: Port đã được sử dụng
```powershell
# Tìm process đang dùng port
netstat -ano | findstr :8080
# Hoặc dừng services cũ
docker stop $(docker ps -q)
```

---

## Tài liệu tham khảo

- [NETWORK_SIMULATION_GUIDE.md](NETWORK_SIMULATION_GUIDE.md) - Hướng dẫn chi tiết
- [QUICK_REFERENCE.md](QUICK_REFERENCE.md) - Tham khảo nhanh
- [HUONG_DAN_TIENG_VIET.md](../HUONG_DAN_TIENG_VIET.md) - Hướng dẫn tiếng Việt đầy đủ
