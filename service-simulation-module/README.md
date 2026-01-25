# Service Simulation Module

Dự án mô phỏng các Internet services (HTTP, DNS) sử dụng INetSim và Flask API để giám sát và thu thập logs.

## 📋 Mục lục
- [Giới thiệu](#giới-thiệu)
- [Cấu trúc dự án](#cấu-trúc-dự-án)
- [Yêu cầu hệ thống](#yêu-cầu-hệ-thống)
- [Cài đặt và chạy](#cài-đặt-và-chạy)
- [Kiểm tra hoạt động](#kiểm-tra-hoạt-động)
- [API Endpoints](#api-endpoints)
- [Cấu hình](#cấu-hình)
- [Troubleshooting](#troubleshooting)

## 🎯 Giới thiệu

Project bao gồm 2 services chính:
1. **INetSim**: Mô phỏng các internet services (HTTP, DNS)
2. **Service-Simulation**: Flask API để quản lý và giám sát INetSim

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
│       ├── api/
│       │   └── server.py      # Flask API endpoints
│       ├── collector/
│       │   └── logs.py        # Log collector
│       └── config/
│           └── inetsim.py     # Generator config INetSim
└── shared/                    # Data được share giữa containers
    ├── config/
    │   └── etc/inetsim/
    │       └── inetsim.conf   # Config file của INetSim
    └── logs/                  # Log files từ cả 2 services
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
  "service": "simulation",
  "status": "running"
}
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
