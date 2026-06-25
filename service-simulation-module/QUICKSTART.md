# 🚀 Hướng Dẫn Chạy Nhanh - HTTP Simulation System

## ⚡ Chạy Nhanh (Quick Start)

### Bước 1: Khởi động Services

```powershell
# Di chuyển vào thư mục
cd D:\PROJECT\Project\pack-a-mal\service-simulation-module

# Build và khởi động containers
docker-compose up -d
```

**Đợi khoảng 10-15 giây để services khởi động hoàn toàn.**

### Bước 2: Kiểm tra hoạt động

```powershell
# Test API
curl http://localhost:5000/status -UseBasicParsing | ConvertFrom-Json
```

**Kết quả mong đợi:**
```json
{
  "service": "http-simulation",
  "status": "running",
  "version": "2.0"
}
```

### Bước 3: Test các tính năng

#### ✅ Test 1: Download file thực thi (Safe)
```powershell
curl http://localhost:5000/tools/installer.exe -OutFile test.exe
```
→ File được sandbox an toàn, không có code thực thi thật

#### ✅ Test 2: Phân tích request
```powershell
$body = @{
    method = "GET"
    url = "/download/malware.exe"
    headers = @{"User-Agent" = "Python/3.9"}
    client_ip = "192.168.1.100"
} | ConvertTo-Json

curl http://localhost:5000/analyze -Method Post -Body $body -ContentType "application/json" | ConvertFrom-Json
```

#### ✅ Test 3: Xem logs executable
```powershell
curl http://localhost:5000/logs/executables | ConvertFrom-Json
```

## 📊 API Endpoints Chính

| Endpoint | Method | Mô tả |
|----------|--------|-------|
| `/status` | GET | Kiểm tra service status |
| `/analyze` | POST | Phân tích HTTP request |
| `/simulate` | POST | Simulate request và trả về response |
| `/logs/executables` | GET | Xem log các executable downloads |
| `/*` | ANY | Catch-all - xử lý mọi request |

## 🎯 Ví Dụ Sử Dụng

### Download Executable (sẽ được sandbox)
```powershell
# Low risk - trả về safe fake file
curl http://localhost:5000/installer.exe -OutFile installer.exe

# Medium risk - trả về honeypot file
curl "http://localhost:5000/backdoor.exe" -Headers @{"User-Agent"="Malware/1.0"} -OutFile backdoor.exe
```

### Test Attack Detection
```powershell
# XSS Attack
curl "http://localhost:5000/search?q=<script>alert('xss')</script>"

# Path Traversal
curl "http://localhost:5000/download/../../../etc/passwd"

# SQL Injection
curl "http://localhost:5000/api?id=1' OR '1'='1"
```

### API Simulation
```powershell
# API request sẽ trả về JSON giả
curl http://localhost:5000/api/v1/users | ConvertFrom-Json
```

### Static Content
```powershell
# CSS, JS, Images
curl http://localhost:5000/styles/main.css
curl http://localhost:5000/scripts/app.js
curl http://localhost:5000/images/logo.png
```

## 🔍 Kiểm Tra Sandbox Files

```powershell
# Xem files trong sandbox
docker exec service-simulation ls -la /logs/executables/

# Xem nội dung file sandbox
docker exec service-simulation cat /logs/executables/*.exe

# Xem metadata
docker exec service-simulation cat /logs/executables/*.metadata.json
```

## 📝 Xem Logs

```powershell
# Xem logs từ service-simulation
docker-compose logs -f service-simulation

# Xem logs từ inetsim
docker-compose logs -f inetsim

# Xem logs cả hai
docker-compose logs -f
```

## 🛑 Dừng Services

```powershell
# Dừng containers (giữ data)
docker-compose stop

# Dừng và xóa containers (giữ images)
docker-compose down

# Xóa hoàn toàn (bao gồm volumes)
docker-compose down -v
```

## 🔄 Restart Services

```powershell
# Restart nhanh
docker-compose restart

# Rebuild và restart
docker-compose up -d --build
```

## 🧪 Chạy Demo Script

```powershell
# Cài đặt requests (nếu chưa có)
pip install requests

# Chạy demo đầy đủ (9 scenarios)
python demo_http_simulation.py

# Chạy test suite (12 tests)
python test_http_simulation.py
```

## 🐛 Troubleshooting

### Lỗi: Port đang được sử dụng
```powershell
# Kiểm tra port
netstat -ano | findstr :5000
netstat -ano | findstr :8080

# Dừng containers cũ
docker-compose down

# Hoặc đổi port trong docker-compose.yml
# Sửa: "5001:5000" thay vì "5000:5000"
```

### Lỗi: Container không start
```powershell
# Xem logs chi tiết
docker-compose logs

# Xóa và rebuild
docker-compose down
docker-compose up --build
```

### Lỗi: Module import error
```powershell
# Rebuild container
docker-compose build service-simulation --no-cache
docker-compose up -d
```

## 📚 Tài Liệu Chi Tiết

- **Comprehensive Guide:** [HTTP_SIMULATION_GUIDE.md](HTTP_SIMULATION_GUIDE.md)
- **Quick Reference:** [QUICK_REFERENCE.md](QUICK_REFERENCE.md)
- **Technical Report:** [../Reports/REPORT_HTTP_EXTENSION.md](../Reports/REPORT_HTTP_EXTENSION.md)
- **Main README:** [README.md](README.md)

## 🎓 Các Tính Năng Chính

### 1. Phân Tích Request
- Trích xuất method, URL, headers, body
- Phát hiện file executable
- Check security threats (XSS, SQLi, etc.)

### 2. Phân Loại Request
- 9 categories tự động
- Confidence scoring
- Risk assessment (low/medium/high)

### 3. Safe Executable Handling
- **Sandbox Fake**: File giả an toàn (low risk)
- **Honeypot**: File tracking (medium risk)
- **Block**: Chặn hoàn toàn (high risk)

### 4. Response Generation
- Static content (CSS, JS, images)
- API responses (JSON)
- Authentication simulation
- File downloads

### 5. Linux Fake Artifacts
- Tạo footprint hệ thống Linux ở `artifact-extension/linux/`
- Sinh `hostname`, `machine-id`, `os-release`, DMI, bash history, syslog
- Kiểm thử bằng `cd artifact-extension; python -m pytest tests -v`

## 💡 Tips

1. **Luôn check status trước:** `curl http://localhost:5000/status`
2. **Xem logs realtime:** `docker-compose logs -f service-simulation`
3. **Test với curl trước khi code:** Dễ debug hơn
4. **Check sandbox files:** Để xem request đã được log chưa
5. **Dùng demo script:** Để hiểu được flow hoàn chỉnh

## 🎯 Use Cases

- ✅ Phân tích hành vi malware
- ✅ Honeypot deployment
- ✅ Security research
- ✅ Package analysis
- ✅ Training & education

## 📞 Support

Nếu gặp vấn đề:
1. Check logs: `docker-compose logs`
2. Xem troubleshooting section ở trên
3. Đọc [HTTP_SIMULATION_GUIDE.md](HTTP_SIMULATION_GUIDE.md)
4. Check [REPORT_HTTP_EXTENSION.md](../Reports/REPORT_HTTP_EXTENSION.md)

---

**Version:** 2.0  
**Last Updated:** February 8, 2026  
**Status:** ✅ Production Ready
