# Service Simulation Module

Dự án mô phỏng các Internet services (HTTP, DNS) sử dụng INetSim và Flask API để giám sát và thu thập logs.

## 🆕 Tính Năng Mở Rộng (Version 2.0)

**Hệ thống giả lập HTTP thông minh** với khả năng:
- ✅ **Phân tích yêu cầu HTTP đến** - Trích xuất và phân tích chi tiết request
- ✅ **Nhận diện mục đích truy cập** - Phân loại tự động 9 loại request khác nhau
- ✅ **Trả về phản hồi phù hợp** - Response động dựa trên loại và risk level
- ✅ **Xử lý an toàn file thực thi** - Sandbox, honeypot, và blocking cho executables
- ✅ **Phát hiện tấn công** - Nhận diện XSS, SQL injection, path traversal, command injection
- ✅ **Logging chi tiết** - Track tất cả executable requests với metadata

📖 **[Xem Hướng Dẫn Chi Tiết](HTTP_SIMULATION_GUIDE.md)**

## 📋 Mục lục
- [Tính Năng Mở Rộng](#-tính-năng-mở-rộng-version-20)
- [Giới thiệu](#giới-thiệu)
- [Cấu trúc dự án](#cấu-trúc-dự-án)
- [Yêu cầu hệ thống](#yêu-cầu-hệ-thống)
- [Cài đặt và chạy](#cài-đặt-và-chạy)
- [Kiểm tra hoạt động](#kiểm-tra-hoạt-động)
- [API Endpoints](#api-endpoints)
- [Demo & Testing](#-demo--testing)
- [Cấu hình](#cấu-hình)
- [Troubleshooting](#troubleshooting)

## 🎯 Giới thiệu

Project bao gồm 2 services chính:
1. **INetSim**: Mô phỏng các internet services (HTTP, DNS)
2. **Service-Simulation**: Flask API thông minh với khả năng phân tích và phản hồi HTTP
3. **Artifact Extension**: Bộ sinh fake artifacts nội bộ cho DNS, browser history, Zone.Identifier và footprint Linux để làm môi trường phân tích trông giống máy thật hơn.

### Linux Fake Artifacts
Phần `artifact-extension` tạo thêm dấu vết Linux để môi trường phân tích trông tự nhiên hơn:

 - `linux/etc/hostname`
 - `linux/etc/machine-id`
 - `linux/etc/os-release`
 - `linux/sys/class/dmi/id/*`
 - `linux/home/<user>/.bash_history`

Chạy test riêng cho phần này bằng:

```bash
cd artifact-extension
python -m pytest tests -v
```

Kết quả mong đợi: toàn bộ test pass, gồm cả kiểm tra Linux artifacts.

## 📁 Cấu trúc dự án

```
service-simulation-module/
├── docker-compose.yml          # Orchestration của 2 containers
├── inetsim/                    # INetSim container
│   ├── Dockerfile             # Build image Ubuntu + INetSim
│   └── entrypoint.sh          # Script khởi động INetSim
├── service-simulation/        # Flask API container
│   ├── Dockerfile             # Build image Python + Flask
│   └── app/
│       ├── main.py            # Entry point - khởi động Flask
│       ├── analyzer/          # 🆕 HTTP Analysis Module
│       │   ├── http_analyzer.py      # Phân tích HTTP requests
│       │   └── request_classifier.py # Phân loại requests
│       ├── handler/           # 🆕 Response Handling Module
│       │   ├── response_handler.py        # Tạo responses
│       │   └── safe_executable_handler.py # Xử lý executables an toàn
│       ├── api/
│       │   └── server.py      # Flask API endpoints (expanded)
│       ├── collector/
│       │   └── logs.py        # Log collector
│       └── config/
│           └── inetsim.py     # Generator config INetSim
├── shared/                    # Data được share giữa containers
│   ├── config/
│   │   └── etc/inetsim/
│   │       └── inetsim.conf   # Config file của INetSim
│   └── logs/                  # Log files từ cả 2 services
│       └── executables/       # 🆕 Sandboxed executable files
│           ├── *.exe          # Fake/honeypot executables
│           ├── *.metadata.json  # Request metadata
│           └── executable_requests.log  # Execution logs
├── artifact-extension/        # Fake artifact generator cho Linux và browser-like footprints
├── HTTP_SIMULATION_GUIDE.md   # 🆕 Comprehensive guide
├── demo_http_simulation.py    # 🆕 Demo script
└── test_http_simulation.py    # 🆕 Test suite
```

## 💻 Yêu cầu hệ thống

### Phần mềm cần cài đặt:
- **Docker Desktop** (Windows/Mac) hoặc **Docker Engine** (Linux)
  - Download: https://www.docker.com/products/docker-desktop
- **Docker Compose** (thường đi kèm Docker Desktop)



## 🚀 Cài đặt và chạy

### Bước 1: Clone hoặc tải project về
```bash
cd d:\PROJECT\service-simulation-module
```

### Bước 2: Build và khởi động containers
```bash
docker-compose build
docker-compose up
```

### Bước 3: Dừng containers
```bash


# Dừng và xóa containers (giữ lại images)
docker-compose down
```

## ✅ Kiểm tra hoạt động

### 1. Kiểm tra containers đang chạy
```bash
docker ps
```

Kết quả mong đợi:
```
CONTAINER ID   IMAGE                    STATUS         PORTS
xxxxxxxx       service-simulation       Up 1 minute    0.0.0.0:5000->5000/tcp
xxxxxxxx       inetsim                  Up 1 minute    0.0.0.0:53->53/udp, 0.0.0.0:8080->80/tcp
```

### 2. Test Flask API
```bash
curl http://localhost:5000/status
```

Response:
```json
{
  "service": "simulation",
  "status": "running"
}
```

### 3. Test HTTP service của INetSim
```bash
curl http://localhost:8080 -UseBasicParsing
```

Sẽ trả về fake HTML page từ INetSim.

### 4. Test DNS service của INetSim
```bash
# Windows PowerShell
nslookup google.com 127.0.0.1

```

## 📡 API Endpoints

### Health Check
```
GET http://localhost:5000/status
```

**Response:**
```json
{
  "service": "http-simulation",
  "status": "running",
  "version": "2.0",
  "features": [
    "http_analysis",
    "request_classification",
    "safe_executable_handling",
    "adaptive_response"
  ]
}
```

### Analyze Request
```
POST http://localhost:5000/analyze
```
Phân tích một HTTP request và trả về thông tin chi tiết về category, intent, risk level.

### Simulate Request
```
POST http://localhost:5000/simulate
```
Simulate một HTTP request và trả về response thực tế.

### View Executable Logs
```
GET http://localhost:5000/logs/executables
```
Xem danh sách tất cả executable download requests đã được log.

### Catch-All Route
```
ANY /*
```
Mọi request khác sẽ được phân tích tự động và trả về response phù hợp.

📖 **[Xem API Documentation Chi Tiết](HTTP_SIMULATION_GUIDE.md#-api-endpoints)**

## 🧪 Demo & Testing

### Quick Demo
Chạy script demo để xem các tính năng chính:

```bash
# Đảm bảo service đang chạy
docker-compose up -d

# Cài dependencies (nếu chưa có)
pip install requests

# Chạy demo
python demo_http_simulation.py
```

Demo sẽ showcase:
- ✅ Service status check
- ✅ Executable download analysis
- ✅ Safe executable handling (sandbox)
- ✅ Honeypot executable for suspicious requests
- ✅ Malicious request detection (XSS, path traversal)
- ✅ API simulation
- ✅ Authentication simulation
- ✅ Static content serving
- ✅ Logging and tracking

### Full Test Suite
Chạy comprehensive test suite:

```bash
# Chạy tất cả tests
python test_http_simulation.py

# Output:
# ============================================================
# HTTP SIMULATION SYSTEM - TEST SUITE
# ============================================================
# TEST: Status Check
# ✓ PASSED
# TEST: Static Content - CSS
# ✓ PASSED
# ...
# TEST SUMMARY
# Total Tests: 12
# Passed: 12 ✓
# Success Rate: 100.0%
```

### Manual Testing Examples

#### Test 1: Download Safe Executable
```bash
curl http://localhost:5000/tools/installer.exe -o installer.exe
# Headers sẽ chứa: X-Sandboxed: true
```

#### Test 2: Trigger Honeypot
```bash
curl http://localhost:5000/malware.exe -H "User-Agent: Suspicious" -o malware.exe
# Headers sẽ chứa: X-Honeypot: true
```

#### Test 3: Malicious Request
```bash
curl "http://localhost:5000/api?id=1' OR '1'='1"
# Risk level: high, có thể bị block
```

#### Test 4: Analyze Request
```bash
curl -X POST http://localhost:5000/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "url": "/download/suspicious.exe",
    "headers": {"User-Agent": "Bot"},
    "client_ip": "192.168.1.1"
  }'
```

## ⚙️ Cấu hình

### INetSim Configuration
File: `shared/config/etc/inetsim/inetsim.conf`

Cấu hình được tự động generate bởi `service-simulation/app/config/inetsim.py`

**Các service được bật:**
- HTTP Server (port 80 → 8080 trên host)
- DNS Server (port 53 → 53 trên host)

### Docker Network
- **Network name**: `simulation_network`
- **Driver**: bridge
- **DNS**: Containers giao tiếp qua tên (không dùng static IP)

### Volumes Mapping
| Host Path | Container Path | Mục đích |
|-----------|----------------|----------|
| `./shared/config/etc/inetsim` | `/etc/inetsim` | Config INetSim |
| `./shared/logs` | `/logs` | Logs từ cả 2 services |
| `./shared/logs/inetsim` | `/var/log/inetsim` | INetSim logs chi tiết |

## 🔧 Troubleshooting

### Lỗi: Port already in use
```
Error: bind: address already in use
```

**Giải pháp:**
```bash
# Kiểm tra process đang dùng port
netstat -ano | findstr :5000
netstat -ano | findstr :8080
netstat -ano | findstr :53

# Kill process hoặc đổi port trong docker-compose.yml
ports:
  - "5001:5000"  # Đổi port host
```

### Lỗi: Container không start được
```bash
# Xem logs chi tiết
docker-compose logs inetsim
docker-compose logs service-simulation

# Restart containers
docker-compose restart
```

### Lỗi: Permission denied (Linux/Mac)
```bash
# Thêm quyền cho shared folders
chmod -R 755 shared/
```

### Rebuild từ đầu
```bash
# Xóa tất cả và build lại
docker-compose down -v
docker-compose up --build
```

## 🛠️ Development

### Thêm dependencies cho Python
1. Cập nhật `service-simulation/Dockerfile`:
```dockerfile
RUN pip install flask requests  # Thêm package mới
```

2. Rebuild:
```bash
docker-compose up --build service-simulation
```

### Xem logs realtime khi develop
```bash
docker-compose logs -f
```

### Vào trong container để debug
```bash
# Vào container service-simulation
docker exec -it service-simulation bash

# Vào container inetsim
docker exec -it inetsim bash
```

## 📝 Notes

- Containers sử dụng Docker DNS để giao tiếp qua tên (`inetsim`, `service-simulation`)
- File config INetSim được auto-generate mỗi khi service-simulation khởi động
- Logs được lưu persistent trong folder `shared/logs/`
- Service-simulation đợi INetSim khởi động xong trước khi start (health check loop)

## 👥 Team Collaboration

### Clone và chạy lần đầu:
```bash
git clone <repository-url>
cd service-simulation-module
docker-compose up --build
```

### Khi có thay đổi code:
```bash
git pull
docker-compose up --build
```

### Best Practices:
- ✅ Commit code thường xuyên
- ✅ Không commit folder `shared/logs/` (add vào .gitignore)
- ✅ Document các thay đổi config trong README
- ✅ Test trước khi push

---

**Tác giả**: [Tên team của bạn]  
**Ngày tạo**: January 2026  
**License**: [License type]
