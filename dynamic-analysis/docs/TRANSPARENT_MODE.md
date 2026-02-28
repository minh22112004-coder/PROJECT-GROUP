# Transparent Mode - Tích hợp từ siemens/sparring

> **Nguồn cảm hứng:** [siemens/sparring](https://github.com/siemens/sparring)  
> **Ngôn ngữ gốc:** Python (sparring) → Go (project này)  
> **Ngày tích hợp:** 26/02/2026

---

## 1. Giới thiệu

**Transparent Mode** là chế độ vận hành thứ ba của Network Mode Controller, được tích hợp dựa trên triết lý của dự án [siemens/sparring](https://github.com/siemens/sparring):

> *"While working transparently, sparring will NOT alter any transmitted data and only log connections and try to extract interesting data for supported protocols."*

Trái ngược với **Full Mode** (cô lập hoàn toàn) và **Half Mode** (proxy có thể chặn/chỉnh sửa), Transparent Mode **không bao giờ can thiệp vào traffic** — chỉ quan sát, theo dõi và ghi log.

### So sánh 3 chế độ

| Tính năng | Full Mode | Half Mode | Transparent Mode |
|       --- |       --- |       --- |           ---    |
| Chặn traffic | ✅ Tất cả | ✅ Có thể | ❌ Không bao giờ |
| Sửa đổi traffic | ❌ | ✅ Có thể | ❌ Không bao giờ |
| Chuyển tiếp ra ngoài | ❌ | ✅ Có thể | ✅ Luôn luôn |
| Ghi log connections | ✅ | ✅ | ✅ Chi tiết hơn |
| Trích xuất payload | ❌ | ❌ | ✅ HTTP/DNS/SMTP/FTP |
| Theo dõi ICMP | ❌ | ❌ | ✅ |
| An toàn với mẫu | ✅ Cao nhất | ⚠️ Trung bình | ✅ Cao (không làm thay đổi hành vi) |

---

## 2. Các file đã thay đổi / tạo mới

### 2.1 File tạo mới

#### `internal/networkmode/transparent.go`
File core chứa toàn bộ logic Transparent Mode.

**Các thành phần chính:**

| Struct/Func | Mô tả | Tương đương trong sparring |
|---|---|---|
| `ConnectionInfo` | Lưu thông tin một kết nối TCP/UDP/ICMP | `sparring's connection dict` |
| `ExtractedPayload` | Dữ liệu payload đã trích xuất | `application.get_stats()` |
| `TransparentModeHandler` | Handler chính, không sửa traffic | `class Sparring (TRANSPARENT mode)` |
| `HandleRequest()` | Điểm vào chính - pass-through | `cb()` callback với `NF_STOP` |
| `trackConnection()` | Theo dõi bảng kết nối | `tcp.connections dict` |
| `identifyAppProtocol()` | Nhận dạng giao thức từ port/payload | `classify()` |
| `extractAndLogPayload()` | Trích xuất data từ giao thức được hỗ trợ | `application modules (http, smtp, ftp, dns)` |
| `PrintSummary()` | In tóm tắt traffic | `print_connections() + print_stats()` |

### 2.2 File đã sửa đổi

#### `internal/networkmode/mode.go`

**Thêm:**
```go
// Constant mới
ModeTransparent Mode = "transparent"

// Struct cấu hình mới
type TransparentModeConfig struct {
    Enabled            bool
    ExtractPayloads    bool
    LogConnections     bool
    LogICMP            bool
    SupportedProtocols []string
    ConnectionLogFile  string
    PayloadLogFile     string
    MaxPayloadSize     int64
}
```

**Cập nhật:**
- `IsValid()` → chấp nhận `ModeTransparent`
- `Config` struct → thêm field `TransparentMode *TransparentModeConfig`
- `DefaultConfig()` → thêm default `TransparentModeConfig` (disabled)
- `Validate()` → thêm validation cho Transparent Mode

---

#### `internal/networkmode/errors.go`

**Thêm:**
```go
ErrTransparentModeNotEnabled = errors.New("transparent mode is not enabled")
```

---

#### `internal/networkmode/controller.go`

**Thêm field:**
```go
type Controller struct {
    // ...
    transparentHandler *TransparentModeHandler  // MỚI
}
```

**Thêm hàm:**
```go
// Xử lý request trong Transparent Mode
// Decision luôn là ActionForward - KHÔNG BAO GIỜ block
func (c *Controller) handleTransparentMode(ctx, req) (*Response, *Decision, error)

// Lấy thống kê Transparent Mode
func (c *Controller) GetTransparentStats() (map[string]interface{}, error)

// In tóm tắt traffic dạng text
func (c *Controller) GetTransparentSummary() (string, error)
```

**Cập nhật:**
- `NewController()` → khởi tạo `transparentHandler` nếu `mode == ModeTransparent`
- `HandleRequest()` → thêm `case ModeTransparent` trong switch
- `SwitchMode()` → hỗ trợ chuyển sang/từ `ModeTransparent`, lazy-init handler
- `Health()` → kiểm tra `transparentHandler != nil` khi ở Transparent Mode
- `Close()` → gọi `transparentHandler.Close()` để flush log files

---

#### `internal/networkmode/router.go`

**Thêm:**
```go
// Trong RouteRequest() switch
case ModeTransparent:
    return r.routeTransparentMode(ctx, req)

// Hàm mới
func (r *Router) routeTransparentMode(ctx, req) (*Response, error)
// → Trả response pass-through, không forward đến đâu cả
// → Set header: X-Pack-A-Mal-Mode: transparent
```

---

#### `internal/networkmode/controller_test.go`

**Thêm test cases:**

| Test | Mô tả |
|---|---|
| `TestMode_IsValid` | Thêm case `ModeTransparent` → expected `true` |
| `TestConfig_Validate` | Thêm 3 case: valid, not-enabled, missing-config |
| `TestController_TransparentMode` | Test đầy đủ: HTTP + DNS request, kiểm tra response passthrough, action = Forward, thống kê |
| `TestController_SwitchToTransparentMode` | Test chuyển Full → Transparent → Full |

---

#### `config/network-mode.yaml`

**Thêm section:**
```yaml
transparent_mode:
  enabled: false
  extract_payloads: true
  log_connections: true
  log_icmp: true
  supported_protocols: [http, https, dns, smtp, ftp]
  connection_log_file: "/logs/transparent_connections.log"
  payload_log_file: "/logs/transparent_payloads.log"
  max_payload_size: 1048576  # 1MB
```

---

#### `examples/networkmode/main.go`

**Thêm Example 5:** `runTransparentModeExample()` minh hoạ:
- Quan sát HTTP C2 beacon từ malware
- Quan sát DNS lookup đến C2 domain
- Quan sát SMTP exfiltration attempt
- In thống kê và connection summary

---

## 3. Cách hoạt động

### 3.1 Luồng xử lý request trong Transparent Mode

```
Malware gửi packet
        │
        ▼
Controller.HandleRequest()
        │
        ├─── LogRequest() → ghi log request
        │
        ├─── case ModeTransparent:
        │         │
        │         ▼
        │    handleTransparentMode()
        │         │
        │         ▼
        │    TransparentModeHandler.HandleRequest()
        │         │
        │         ├── trackConnection()
        │         │       → Tạo/cập nhật ConnectionInfo trong bảng connections
        │         │       → Nhận dạng AppProtocol (HTTP/DNS/SMTP/FTP)
        │         │       → Cập nhật thống kê (TCP/UDP counter, ProtocolBreakdown)
        │         │
        │         ├── extractAndLogPayload()  [nếu ExtractPayloads=true]
        │         │       → Chỉ chạy cho giao thức trong SupportedProtocols
        │         │       → Truncate nếu vượt MaxPayloadSize
        │         │       → Parse HTTP headers, DNS query, SMTP commands, FTP commands
        │         │       → Ghi JSON vào payload_log_file
        │         │
        │         ├── writeConnectionLog()  [nếu LogConnections=true]
        │         │       → Ghi JSON vào connection_log_file
        │         │
        │         └── Return Response{Source: "transparent_passthrough"}
        │                        ← KHÔNG SỬA ĐỔI GÌ
        │
        ├─── Decision = {Action: ActionForward}
        │                ← Luôn luôn Forward (không block, không modify)
        │
        └── LogTraffic() → ghi traffic log
```

### 3.2 Tracking kết nối

```
Connection key = "src_ip:src_port->dst_ip:dst_port/protocol"

Lần đầu thấy key:   → Tạo ConnectionInfo mới, thêm vào map
Lần tiếp theo:      → Cập nhật BytesSent, AppProtocol
```

Giao thức được nhận dạng theo thứ tự ưu tiên:
1. Protocol đã biết từ request (HTTP, HTTPS, DNS, SMTP, FTP)
2. Port mapping (80→HTTP, 443→HTTPS, 53→DNS, 25→SMTP, 21→FTP, ...)
3. Payload prefix inspection (`GET `, `POST`, `EHLO`, `USER`, ...)

### 3.3 Format log file

**connection_log_file** (JSONL - 1 JSON object mỗi dòng):
```json
{"timestamp":"2026-02-26T10:00:00Z","event":"observed","connection_id":"abc123","protocol":"TCP","app_protocol":"HTTP","src":"192.168.1.50:55001","dst":"203.0.113.42:80","domain":"malware-c2.example.com","bytes_sent":42}
```

**payload_log_file** (JSONL):
```json
{"connection_id":"abc123","timestamp":"2026-02-26T10:00:01Z","protocol":"HTTP","direction":"outgoing","parsed_data":{"method":"POST","path":"/beacon","host":"malware-c2.example.com","user_agent":"Mozilla/5.0 (compatible; bot/1.0)","full_url":"http://malware-c2.example.com/beacon"},"size":42,"truncated":false}
```

### 3.4 Payload extraction theo giao thức

| Giao thức | Dữ liệu được trích xuất |
|---|---|
| **HTTP/HTTPS** | Method, Path, Host, User-Agent, full URL, Content-Type, sensitive headers (Authorization, Cookie, ...) |
| **DNS** | Queried domain, DNS port type (standard/DoT) |
| **SMTP** | EHLO/HELO/MAIL FROM/RCPT TO commands (không lấy PASS) |
| **FTP** | USER, RETR/STOR/LIST/CWD commands (không lấy PASS) |

---

## 4. Cách sử dụng

### 4.1 Sử dụng trực tiếp trong code Go

```go
import "github.com/ossf/package-analysis/internal/networkmode"

// Tạo cấu hình Transparent Mode
config := networkmode.DefaultConfig()
config.Mode = networkmode.ModeTransparent
config.TransparentMode = &networkmode.TransparentModeConfig{
    Enabled:            true,
    ExtractPayloads:    true,
    LogConnections:     true,
    LogICMP:            true,
    SupportedProtocols: []string{"http", "https", "dns", "smtp", "ftp"},
    ConnectionLogFile:  "/logs/transparent_connections.log",
    PayloadLogFile:     "/logs/transparent_payloads.log",
    MaxPayloadSize:     1 * 1024 * 1024, // 1MB
}

// Tạo controller
controller, err := networkmode.NewController(config, slog.Default())
if err != nil {
    log.Fatal(err)
}
defer controller.Close()

ctx := context.Background()

// Xử lý requests - traffic KHÔNG bị sửa đổi
resp, err := controller.HandleRequest(ctx, req)
// resp.Source == "transparent_passthrough"
// resp.Decision.Action == ActionForward  (luôn luôn)

// Lấy thống kê
stats, _ := controller.GetTransparentStats()
fmt.Println(stats["total_connections"])
fmt.Println(stats["protocol_breakdown"])

// In connection summary
summary, _ := controller.GetTransparentSummary()
fmt.Print(summary)
```

### 4.2 Cấu hình qua YAML

Sửa file `config/network-mode.yaml`:

```yaml
network_mode:
  mode: "transparent"   # Đổi từ "full" hoặc "half"

  transparent_mode:
    enabled: true        # PHẢI là true
    extract_payloads: true
    log_connections: true
    log_icmp: true
    supported_protocols:
      - "http"
      - "https"
      - "dns"
      - "smtp"
      - "ftp"
    connection_log_file: "/logs/transparent_connections.log"
    payload_log_file: "/logs/transparent_payloads.log"
    max_payload_size: 1048576
```

### 4.3 Chuyển chế độ động (runtime mode switching)

```go
ctx := context.Background()

// Full Mode → Transparent Mode
err = controller.SwitchMode(ctx, networkmode.ModeTransparent)

// Transparent Mode → Half Mode
err = controller.SwitchMode(ctx, networkmode.ModeHalf)

// Transparent Mode → Full Mode (an toàn nhất)
err = controller.SwitchMode(ctx, networkmode.ModeFull)
```

---

## 5. Use Cases điển hình

### 5.1 Phân tích pháp y bị động (Passive Forensic Analysis)
Khi cần quan sát hành vi thực của malware mà không làm nó thay đổi hành vi do bị phát hiện:
```
Full Mode → Transparent Mode → quan sát → Full Mode (lock down)
```

### 5.2 Baseline profiling
Chạy mẫu trong Transparent Mode trước để hiểu các domain/IP nó liên lạc, sau đó cấu hình rules cho Half Mode:
```
Transparent Mode (quan sát 5 phút)
         ↓
Đọc connection log → tạo blacklist/whitelist
         ↓
Half Mode với custom rules
```

### 5.3 Phát hiện C2 channels
Theo dõi bất thường trong User-Agent, DNS queries bất thường, SMTP exfiltration:
```go
stats, _ := controller.GetTransparentStats()
breakdown := stats["protocol_breakdown"].(map[string]int64)
// Nếu SMTP có nhiều kết nối → nghi ngờ exfiltration
// Nếu DNS có nhiều unique domains → nghi ngờ DGA
```

---

## 6. Bảo mật và hạn chế

| Vấn đề | Xử lý |
|---|---|
| Handler chưa init | Fallback tự động về Full Mode (fail-safe) |
| Log file không mở được | Fallback về slogger, tiếp tục hoạt động |
| Payload quá lớn | Truncate theo `MaxPayloadSize`, đánh dấu `truncated: true` |
| SMTP/FTP password | **Không trích xuất** lệnh PASS để tránh lưu credentials |
| Mode phải enabled | `Validate()` trả lỗi nếu `enabled: false` khi `mode: transparent` |

---

## 7. Cấu trúc file sau khi tích hợp

```
dynamic-analysis/internal/networkmode/
├── mode.go              ← Thêm ModeTransparent, TransparentModeConfig
├── transparent.go       ← MỚI - Core logic Transparent Mode
├── controller.go        ← Thêm transparentHandler, handleTransparentMode()
├── router.go            ← Thêm routeTransparentMode()
├── errors.go            ← Thêm ErrTransparentModeNotEnabled
├── controller_test.go   ← Thêm 4 test cases mới
├── interceptor.go       ← Không thay đổi
├── decision.go          ← Không thay đổi
├── modifier.go          ← Không thay đổi
├── logger.go            ← Không thay đổi
├── request.go           ← Không thay đổi
└── README.md

dynamic-analysis/config/
└── network-mode.yaml    ← Thêm section transparent_mode

dynamic-analysis/examples/networkmode/
└── main.go              ← Thêm Example 5: runTransparentModeExample()

dynamic-analysis/docs/
└── TRANSPARENT_MODE.md  ← File này
```
