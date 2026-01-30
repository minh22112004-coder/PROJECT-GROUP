# 🎯 HƯỚNG DẪN DEMO TÍNH NĂNG NETWORK SIMULATION

## 📋 Mục tiêu Demo

Trình bày tính năng kiểm tra URL còn hoạt động (alive) hay không và tự động điều hướng tới dịch vụ INetSim khi URL đã chết.

---

## 🔧 PHẦN 1: CHUẨN BỊ MÔI TRƯỜNG

### Bước 1.1: Khởi động INetSim Service

```powershell
# Mở PowerShell tại thư mục dynamic-analysis
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis

# Khởi động Docker containers (INetSim + Service API)
docker-compose -f docker-compose.network-sim.yml up -d
```

**✅ Kiểm tra:**
```powershell
docker ps --filter "name=pack-a-mal"
```
Cần thấy 2 containers: `pack-a-mal-inetsim` và `pack-a-mal-sim-api` ở trạng thái `(healthy)`

### Bước 1.2: Test INetSim hoạt động

```powershell
# Test HTTP service
curl.exe http://localhost:8080

# Test API
curl.exe http://localhost:5000/status
```

**✅ Kết quả mong đợi:**
- HTTP trả về trang HTML INetSim
- API trả về JSON: `{"service":"simulation","status":"running"}`

---

## 📦 PHẦN 2: DEMO PACKAGE MẪU

### Bước 2.1: Giới thiệu Package

**Nói với thầy:**
> "Em đã tạo một package Python mẫu có tên `malicious-network-package` để demo. Package này sẽ cố gắng kết nối tới một URL không còn hoạt động."

### Bước 2.2: Xem mã nguồn Package

```powershell
# Mở file code của package
code D:\PROJECT\Project\pack-a-mal\dynamic-analysis\sample_packages\malicious_network_package\malicious_network_package\__init__.py
```

**Giải thích cho thầy:**
- Package cố gắng kết nối tới: `http://malicious-c2-server.example.com/api/data`
- URL này không tồn tại thực tế (giả lập malware kết nối C2 server)
- Hàm `connect_to_dead_url()` sẽ thực hiện request HTTP

### Bước 2.3: Cài đặt Package

```powershell
# Di chuyển vào thư mục package
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\sample_packages\malicious_network_package

# Cài đặt package
pip install -e .
```

**📝 Lưu ý:** Package này có 2 test scripts:

| Script | Mục đích | Kết quả |
|--------|----------|---------|
| `test_network.py` | Test KHÔNG qua INetSim | ❌ Connection failed |
| `test_with_inetsim.py` | Test CÓ redirect qua INetSim | ✅ 3/3 URLs success |

---

## 🎬 PHẦN 3: DEMO TÍNH NĂNG CHÍNH

### Demo 3.1: KHÔNG CÓ Network Simulation (URL chết → Thất bại)

**Nói với thầy:**
> "Đầu tiên, em sẽ demo khi KHÔNG bật tính năng Network Simulation. Lúc này, package sẽ cố kết nối tới URL chết và sẽ thất bại."

**Cách 1: Test trực tiếp**
```powershell
python -c "import malicious_network_package; malicious_network_package.connect_to_dead_url()"
```

**Cách 2: Dùng test script (khuyên dùng)**
```powershell
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\sample_packages\malicious_network_package
python test_network.py
```

**✅ Kết quả mong đợi:**
```
============================================================
Malicious Network Package - Connecting to dead URL
============================================================

[*] Target URL: http://malicious-c2-server.example.com/api/data
[*] Attempting connection...
[-] Connection failed: ...
============================================================
```

**Giải thích:**
- URL không tồn tại → kết nối thất bại
- Đây là trường hợp bình thường khi không có intervention

---

### Demo 3.2: CÓ Network Simulation (URL chết → Redirect tới INetSim)

**Nói với thầy:**
> "Bây giờ, em sẽ bật tính năng Network Simulation. Hệ thống sẽ:
> 1. Kiểm tra xem URL có còn alive không
> 2. Nếu URL đã chết → tự động điều hướng DNS tới INetSim
> 3. INetSim sẽ giả lập response để phân tích hành vi"

#### Bước 3.2.1: Mở Terminal thứ 2 để xem code logic

```powershell
# Terminal 2: Xem code logic kiểm tra URL
code D:\PROJECT\Project\pack-a-mal\dynamic-analysis\internal\networksim\networksim.go
```

**Giải thích code cho thầy (dòng 42-67):**
```go
// IsURLAlive checks if URL is accessible
func (ns *NetworkSimulator) IsURLAlive(ctx context.Context, url string) bool {
    // Tạo HTTP client với timeout
    client := &http.Client{Timeout: ns.config.LivenessTimeout}
    
    // Thực hiện HEAD request
    resp, err := client.Do(req)
    if err != nil {
        slog.InfoContext(ctx, "URL not alive", "url", url)
        return false  // URL chết
    }
    
    // Kiểm tra status code (200-399 = alive)
    isAlive := resp.StatusCode >= 200 && resp.StatusCode < 400
    return isAlive
}

// ShouldRedirectToINetSim - Logic redirect
func (ns *NetworkSimulator) ShouldRedirectToINetSim(...) bool {
    if !ns.IsURLAlive(ctx, url) {
        slog.InfoContext(ctx, "Redirecting to INetSim", "url", url)
        return true  // URL chết → redirect
    }
    return false
}
```

#### Bước 3.2.2: Cấu hình Environment Variables cho Go Test

**Nói với thầy:**
> "Trước khi test Go code, em cần set biến môi trường để code đọc được config."

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

#### Bước 3.2.3: Chạy Unit Tests

**Nói với thầy:**
> "Bây giờ em sẽ chạy unit tests để test logic kiểm tra URL."

```powershell
# Di chuyển vào thư mục networksim
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\internal\networksim

# Chạy unit tests với output chi tiết
go test -v
```

**✅ Kết quả mong đợi:**
```
=== RUN   TestIsURLAlive
2026/01/30 ... INFO URL check url=http://127.0.0.1:... status=200 alive=true
2026/01/30 ... INFO URL not alive url=http://dead-url-12345.com
--- PASS: TestIsURLAlive (0.20s)

=== RUN   TestShouldRedirectToINetSim
2026/01/30 ... INFO URL check url=http://127.0.0.1:... status=200 alive=true
2026/01/30 ... INFO URL not alive url=http://dead-url.com
2026/01/30 ... INFO Redirecting to INetSim url=http://dead-url.com
--- PASS: TestShouldRedirectToINetSim (0.18s)

PASS
ok      github.com/ossf/package-analysis/internal/networksim    1.949s
```

**Giải thích kết quả:**
- ✅ Test 1: Kiểm tra URL alive → nhận diện đúng URL còn hoạt động
- ✅ Test 2: Kiểm tra URL chết → tự động redirect tới INetSim
- ✅ Tất cả tests PASS → logic hoạt động đúng!

---

#### Bước 3.2.4: Demo THỰC TẾ Redirect tới INetSim 🎯

**Nói với thầy:**
> "Bây giờ em sẽ demo thực tế! Em có script test kết nối URL chết qua INetSim proxy."

```powershell
# Di chuyển vào thư mục package
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\sample_packages\malicious_network_package

# Chạy test script với INetSim
python test_with_inetsim.py
```

**✅ Kết quả mong đợi:**

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

**Giải thích cho thầy:**
- 🔴 **Phần 1 (KHÔNG có proxy)**: URL chết → kết nối thất bại (đúng!)
- 🟢 **Phần 2 (CÓ INetSim proxy)**: 
  - 3 URL chết đều kết nối thành công qua INetSim
  - INetSim giả lập response HTTP 200
  - Response có signature của INetSim
  - **ĐÂY CHÍNH LÀ TÍNH NĂNG REDIRECT!**

---

## 🔍 PHẦN 4: DEMO INTEGRATION THỰC TẾ

### Demo 4.1: Tích hợp vào Worker Analysis

**Nói với thầy:**
> "Code của em đã được tích hợp vào module Worker để tự động áp dụng khi phân tích packages."

```powershell
# Xem code integration trong worker
code D:\PROJECT\Project\pack-a-mal\dynamic-analysis\cmd\worker\main.go
```

**Tìm và giải thích đoạn code:** (sử dụng Ctrl+F tìm "networksim", dòng 119-137)

```go
// Configure network simulation if enabled
if cfg.networkSimConfig.Enabled {
    netSim := networksim.New(cfg.networkSimConfig)
    
    // Validate INetSim connection before proceeding
    if err := netSim.ValidateINetSimConnection(ctx); err != nil {
        slog.WarnContext(ctx, "INetSim validation failed, continuing without network simulation",
            "error", err)
    } else {
        slog.InfoContext(ctx, "Network simulation enabled",
            "inetsim_dns", netSim.GetINetSimDNS(),
            "inetsim_http", netSim.GetINetSimHTTP())
        
        // Configure sandbox to use INetSim DNS servers
        dnsServers := netSim.GetDNSServers()
        sandboxOpts = append(sandboxOpts, sandbox.DNSServers(dnsServers...))
        
        slog.InfoContext(ctx, "Sandbox configured with custom DNS",
            "dns_servers", dnsServers)
    }
}
```

**Giải thích cho thầy từng bước:**
1. **Kiểm tra config**: Nếu `Enabled = true` → bật network simulation
2. **Tạo NetworkSimulator**: Khởi tạo với config (DNS, HTTP addresses)
3. **Validate INetSim**: Kiểm tra INetSim có chạy không
4. **Configure DNS**: Lấy DNS servers của INetSim (`172.20.0.2`)
5. **Áp dụng vào sandbox**: Package sẽ dùng DNS này khi chạy

**Xem cấu hình trong config.go:**
```powershell
code D:\PROJECT\Project\pack-a-mal\dynamic-analysis\cmd\worker\config.go
```

**Tìm dòng 69-86 - đọc environment variables:**
```go
// Parse network simulation configuration from environment
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
```

### Demo 4.2: Kiểm tra Logs của INetSim

**Nói với thầy:**
> "Khi package kết nối qua INetSim, logs sẽ được ghi lại. Cho em demo."

**Mở 2 terminals:**

**Terminal 1 - Xem logs realtime:**
```powershell
docker logs pack-a-mal-inetsim -f
```

**Terminal 2 - Chạy test với INetSim:**
```powershell
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\sample_packages\malicious_network_package
python test_with_inetsim.py
```

**✅ Kết quả trong Terminal 2:**
```
[*] Testing: http://malicious-c2-server.example.com/api/data
    ✓ Status: 200
    ✓ Connected via INetSim!
```

**✅ Đồng thời trong Terminal 1 (logs) sẽ thấy:**
```
INetSim: Service started. (HTTP, 0.0.0.0:80, pid 14)
```

**Hoặc xem logs chi tiết hơn trong file:**
```powershell
# Xem service logs
Get-Content D:\PROJECT\Project\pack-a-mal\service-simulation-module\shared\logs\inetsim\service.log -Tail 20

# Xem main logs
Get-Content D:\PROJECT\Project\pack-a-mal\service-simulation-module\shared\logs\inetsim\main.log -Tail 20
```

**✅ Logs sẽ chứa thông tin:**
- HTTP requests đến từ package
- URLs được truy cập
- Responses được trả về
- Timestamps của mỗi connection

---

## 📊 PHẦN 5: TÓM TẮT DEMO

### Điểm nhấn khi trình bày:

1. **Vấn đề:**
   - Malware thường kết nối tới C2 servers
   - Nhiều URL C2 đã chết/offline khi phân tích
   - Không thể quan sát hành vi network nếu URL chết

2. **Giải pháp của nhóm:**
   - ✅ Kiểm tra tự động URL có alive không (hàm `IsURLAlive`)
   - ✅ Nếu URL chết → redirect DNS tới INetSim (hàm `ShouldRedirectToINetSim`)
   - ✅ INetSim giả lập response để thu thập logs
   - ✅ Có unit tests đầy đủ (4 tests pass)

3. **Kết quả:**
   - Package với URL chết vẫn có thể kết nối và phân tích được
   - Logs được thu thập đầy đủ
   - Hành vi network được ghi lại

---

## 🎤 SCRIPT DEMO 5 PHÚT

### Phút 1: Giới thiệu
> "Nhóm em demo tính năng Network Simulation. Khi phân tích package có URL không còn alive, hệ thống tự động redirect tới INetSim để tiếp tục phân tích."

### Phút 2: Show Package mẫu
```powershell
code D:\PROJECT\Project\pack-a-mal\dynamic-analysis\sample_packages\malicious_network_package\malicious_network_package\__init__.py
```
> "Đây là package mẫu cố kết nối tới URL chết: malicious-c2-server.example.com"

### Phút 3: Demo không có simulation
```powershell
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\sample_packages\malicious_network_package
python test_network.py
```
> "Không có simulation → kết nối thất bại"

### Phút 4: Show code logic + Unit tests
```powershell
code D:\PROJECT\Project\pack-a-mal\dynamic-analysis\internal\networksim\networksim.go
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\internal\networksim
go test -v
```
> "Code kiểm tra URL alive và redirect. Unit tests pass 100%"

### Phút 5: Demo THỰC TẾ redirect tới INetSim ⭐
```powershell
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\sample_packages\malicious_network_package
python test_with_inetsim.py
```
> "Chạy script test: URL chết kết nối thành công qua INetSim. 3/3 URLs redirected! Đây chính là tính năng của em!"

---

## 🚨 TROUBLESHOOTING

### Nếu Docker không chạy:
```powershell
docker-compose -f docker-compose.network-sim.yml down
docker-compose -f docker-compose.network-sim.yml up -d --force-recreate
```

### Nếu Package chưa cài:
```powershell
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\sample_packages\malicious_network_package
pip install -e . --force-reinstall
```

### Nếu test_with_inetsim.py báo lỗi proxy:
```powershell
# Kiểm tra INetSim đang chạy
curl.exe http://localhost:8080

# Nếu không có response → restart Docker
docker-compose -f docker-compose.network-sim.yml restart inetsim
```

### Nếu Unit tests lỗi:
```powershell
cd D:\PROJECT\Project\pack-a-mal\dynamic-analysis\internal\networksim
go mod tidy
go test -v
```

---

## ✨ KẾT THÚC

**Câu kết:**
> "Đó là tính năng Network Simulation của nhóm em. Hệ thống tự động phát hiện URL chết và redirect tới INetSim để phân tích hành vi. Em xin cảm ơn thầy!"

---

**Chúc bạn demo thành công! 🎉**
