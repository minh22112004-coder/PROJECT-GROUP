# Cấu Trúc Project - Network Simulation Integration

## Tổng Quan Thay Đổi

Tài liệu này hiển thị cấu trúc project Pack-A-Mal và đánh dấu các file/folder được **thêm mới** hoặc **chỉnh sửa** để tích hợp tính năng Network Simulation.

### Ký Hiệu
- ✨ **NEW** - File/folder mới được tạo
- 🔧 **MODIFIED** - File đã tồn tại, được chỉnh sửa
- 📁 - Folder gốc không thay đổi
- 📄 - File gốc không thay đổi

---

## Cấu Trúc Project

```
pack-a-mal/
├── 📄 README.md
│
├── 📁 dynamic-analysis/
│   ├── 📄 README.md
│   ├── 📄 go.mod
│   ├── 📄 Makefile
│   ├── ✨ HUONG_DAN_CHAY.md                    # Hướng dẫn chạy Network Simulation
│   ├── ✨ .env.example                         # Environment variables mẫu
│   ├── ✨ docker-compose.network-sim.yml       # Docker Compose cho INetSim
│   │
│   ├── 📁 cmd/
│   │   ├── 📁 worker/
│   │   │   ├── 🔧 main.go                      # MODIFIED: Tích hợp NetworkSimulator
│   │   │   ├── 🔧 config.go                    # MODIFIED: Thêm network sim config
│   │   │   └── 📄 pubsubextender/...
│   │   ├── 📁 analyze/...
│   │   ├── 📁 scheduler/...
│   │   └── 📁 downloader/...
│   │
│   ├── 📁 internal/
│   │   ├── ✨ networksim/                      # NEW: Module Network Simulation
│   │   │   ├── ✨ networksim.go               # Core logic: URL liveness, INetSim redirect
│   │   │   └── ✨ networksim_test.go          # Unit tests (20 test cases)
│   │   │
│   │   ├── 📁 sandbox/
│   │   │   └── 🔧 sandbox.go                   # MODIFIED: Thêm custom DNS server support
│   │   │
│   │   ├── 📁 analysis/...
│   │   ├── 📁 dynamicanalysis/...
│   │   ├── 📁 log/...
│   │   ├── 📁 worker/...
│   │   └── 📁 utils/...
│   │
│   ├── 📁 sample_packages/
│   │   ├── ✨ malicious_network_package/       # NEW: Sample malicious package
│   │   │   ├── ✨ README.md                    # Mô tả package
│   │   │   ├── ✨ setup.py                     # Python package setup
│   │   │   ├── ✨ test_network.py              # Test script
│   │   │   └── ✨ malicious_network_package/
│   │   │       └── ✨ __init__.py             # Package code với dead URLs
│   │   │
│   │   └── 📁 sample_python_package/...
│   │
│   ├── 📁 scripts/
│   │   ├── ✨ setup_network_simulation.sh      # Setup automation script
│   │   ├── ✨ test_network_simulation.sh       # Test automation script
│   │   ├── ✨ test_inetsim_integration.py      # Integration test script (Python)
│   │   ├── 📄 analyse-tarballs.sh
│   │   ├── 📄 deploy.sh
│   │   └── 📄 run_analysis.sh
│   │
│   ├── 📁 examples/
│   │   ├── 🔧 README.md                        # MODIFIED: Added network-simulation link
│   │   ├── ✨ network-simulation/              # NEW: Network Simulation demo & docs
│   │   │   ├── ✨ README.md                    # Hướng dẫn sử dụng demo
│   │   │   └── ✨ demo_network_simulation.py   # Demo script
│   │   │
│   │   ├── 📁 custom-sandbox/...
│   │   └── 📁 e2e/...
│   │
│   ├── 📁 sandboxes/...
│   ├── 📁 function/...
│   ├── 📁 infra/...
│   ├── 📁 pkg/...
│   └── 📁 tools/...
│
├── ✨ service-simulation-module/               # NEW: INetSim Docker services
│   ├── ✨ README.md
│   ├── ✨ docker-compose.yml
│   │
│   ├── ✨ inetsim/                             # INetSim container
│   │   ├── ✨ Dockerfile
│   │   └── ✨ entrypoint.sh
│   │
│   ├── ✨ service-simulation/                  # Service Simulation API
│   │   ├── ✨ Dockerfile
│   │   └── ✨ app/
│   │       ├── ✨ main.py
│   │       ├── ✨ api/server.py
│   │       ├── ✨ collector/logs.py
│   │       └── ✨ config/inetsim.py
│   │
│   └── ✨ shared/                              # Shared configs & logs
│       ├── ✨ config/etc/inetsim/
│       │   └── ✨ inetsim.conf
│       └── ✨ logs/inetsim/
│           ├── ✨ debug.log
│           ├── ✨ main.log
│           └── ✨ service.log
│
└── 📁 web/...
```

---

## Chi Tiết Các Thay Đổi

### 1️⃣ Core Network Simulation Module

**Folder:** `dynamic-analysis/internal/networksim/` ✨

Chứa logic chính:
- **`networksim.go`** (~120 lines):
  - `IsURLAlive()` - Kiểm tra URL có alive không (HEAD request)
  - `ShouldRedirectToINetSim()` - Quyết định redirect (nếu URL không alive)
  - `GetDNSServers()` - Trả về DNS servers cho sandbox
  - `ValidateINetSimConnection()` - Validate INetSim (đơn giản)

- **`networksim_test.go`** (~80 lines):
  - 4 unit test cases chính
  - Test coverage: URL liveness, redirection, DNS config

**Mục đích:** Thực hiện yêu cầu *"kiểm tra xem URL có alive hay không, nếu không alive thì điều hướng tới dịch vụ Inetsim"*

---

### 2️⃣ Worker Integration

**Files:**
- `cmd/worker/main.go` 🔧
- `cmd/worker/config.go` 🔧

**Thay đổi:**
```go
// config.go - Thêm NetworkSimConfig
type config struct {
    // ... existing fields
    networkSimConfig  *networksim.Config  // NEW
}

// main.go - Validate INetSim và configure sandbox
if config.networkSimConfig.IsEnabled {
    networksim.ValidateINetSimConnection(...)
    dnsServers := networkSim.GetDNSServers()
    sandbox.DNSServers(dnsServers)  // Configure sandbox DNS
}
```

**Mục đích:** Tích hợp network simulation vào worker flow

---

### 3️⃣ Sandbox DNS Configuration

**File:** `internal/sandbox/sandbox.go` 🔧

**Thay đổi:**
- Thêm field `dnsServers []string`
- Thêm function `DNSServers(servers []string)` option
- Modify `createContainer()` để dùng custom DNS thay vì hardcode 8.8.8.8

**Mục đích:** Cho phép sandbox sử dụng INetSim DNS server (172.20.0.2:53)

---

### 4️⃣ Sample Malicious Package

**Folder:** `sample_packages/malicious_network_package/` ✨

**Files:**
- `__init__.py` - Package code với các functions:
  - `check_network_connectivity()`
  - `attempt_http_requests()`
  - `exfiltrate_data()`
  - `download_payload()`

- `test_network.py` - Test script

**Dead URLs used:**
- `malicious-c2-server.example.com`
- `expired-malware-repo.net`
- `dead-phishing-site.org`
- `fake-cdn.badsite.com`

**Mục đích:** Thực hiện yêu cầu *"Tạo một package mẫu có kết nối tới một URL (không còn alive)"*

---

### 5️⃣ INetSim Services

**Folder:** `service-simulation-module/` ✨

**Cấu trúc:**
- **inetsim/** - INetSim 1.3.2 Docker container
  - Dockerfile (Ubuntu 22.04 + INetSim)
  - entrypoint.sh

- **service-simulation/** - Flask API để quản lý
  - API endpoints: `/status`, `/logs`, `/stats`
  - Log collector
  - Config management

- **shared/** - Shared resources
  - Config files: `inetsim.conf`
  - Logs: `service.log`, `debug.log`, `main.log`

**Services:**
- DNS (port 53) → 172.20.0.2:53
- HTTP (port 80) → localhost:8080
- HTTPS (port 443) → localhost:8443
- FTP (port 21) → localhost:8021
- SMTP (port 25) → localhost:8025

**Mục đích:** Cung cấp fake network services cho phân tích malware an toàn

---

### 6️⃣ Docker Compose Configuration

**File:** `docker-compose.network-sim.yml` ✨

**Services:**
```yaml
inetsim:
  - Network: pack-a-mal-network (172.20.0.0/24)
  - IP: 172.20.0.2
  - Ports: 53, 80, 443, 21, 25

service-simulation:
  - API port: 5000
  - Depends on: inetsim
```

**Mục đích:** Orchestration cho INetSim services

---

### 7️⃣ Documentation & Scripts

**Files:**
- ✨ `HUONG_DAN_CHAY.md` - Hướng dẫn chạy bằng tiếng Việt
- ✨ `.env.example` - Environment variables mẫu
- ✨ `scripts/setup_network_simulation.sh` - Setup script
- ✨ `scripts/test_network_simulation.sh` - Test script

**Mục đích:** Hướng dẫn sử dụng và automation

---

## Environment Variables

```bash
# Network Simulation
OSSF_NETWORK_SIMULATION_ENABLED=true
OSSF_INETSIM_DNS_ADDR=17 files**
- Go code: 2 files (networksim.go, networksim_test.go)
- Python: 3 files (sample package + tests)
- Docker: 3 files (Dockerfiles, docker-compose)
- Config: 4 files (.env.example, inetsim.conf, entrypoint.sh, etc.)
- Documentation: 2 files (HUONG_DAN_CHAY.md, examples/network-simulation/README.md)
- Scripts: 2 files (.sh automation)
- Demo/Test: 2 files (demo_network_simulation.py, test_inetsim_integration.py)
- Service API: 4 files (Flask app)

### Files Đã Sửa: **5 files**
- cmd/worker/main.go
- cmd/worker/config.go
- internal/sandbox/sandbox.go
- README.md
- examples/README.md

### Files Đã Di Chuyển: **2 files**
- demo_network_simulation.py → examples/network-simulation/
- test_inetsim_integration.py → scripts/ files (.env.example, inetsim.conf, entrypoint.sh, etc.)
- Documentation: 1 file (HUONG_DAN_CHAY.md)
- Scripts: 2 files (.sh automation)
- Service API: 4 files (Flask app)

### Files Đã Sửa: **4 files**
- cmd/worker/main.go
- cmd/worker/config.go
- internal/sandbox/sandbox.go
- README.md

### Tính Năng Hoàn Thành ✅
1. ✅ Network simulation module với URL liveness checking
2. ✅ INetSim integration để redirect dead URLs
3. ✅ Sample malicious package với dead URLs
4. ✅ Sandbox DNS configuration tự động
5. ✅ Docker services (INetSim + API)
6. ✅ Full documentation

---

**Tác giả:** GitHub Copilot  
**Ngày tạo:** 2026-01-25  
**Project:** Pack-A-Mal Network Simulation Integration
