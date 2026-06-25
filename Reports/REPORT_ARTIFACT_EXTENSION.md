# BÁO CÁO: UBER-LIKE ARTIFACT EXTENSION — KỸ THUẬT GIẢ LẬP PERSISTENT SYSTEM ARTIFACTS

**Ngày hoàn thành:** 18/05/2026  
**Tác giả:** GitHub Copilot (Claude Sonnet 4.6)  
**Dự án:** Pack-A-Mal — Dynamic Malware Analysis Framework  
**Module:** `service-simulation-module/artifact-extension`

---

## 1. TỔNG QUAN

### 1.1. Vấn đề nghiên cứu

Một trong những thách thức lớn nhất trong phân tích malware động là hiện tượng **sandbox evasion** (né tránh sandbox). Malware thông minh thực hiện một loạt kiểm tra môi trường trước khi chạy payload thật — nếu phát hiện sandbox, nó tự huỷ, ngủ đông, hoặc thực thi hành vi vô hại, khiến sandbox không thu được IOC có giá trị.

Các dấu hiệu sandbox phổ biến bị kiểm tra bao gồm:

| Dấu hiệu | Lý do bị detect |
|---|---|
| DNS resolver cache trống | Máy thật luôn tích luỹ hàng trăm DNS query |
| DNS TTL đồng nhất | Inject hàng loạt không tự nhiên |
| Browser history không tồn tại | Người dùng thật có lịch sử duyệt web |
| Visit transition type đồng nhất | Lịch sử được sinh tự động |
| Không có file download | Máy thật luôn có file đã tải |
| Chrome epoch sai | Timestamp tính từ 1970 thay vì 1601 |

### 1.2. Giải pháp đề xuất

Xây dựng **UBER-like Artifact Extension** — module Python chạy theo mô hình **init container** trong Docker, tự động sinh và inject các *persistent system artifacts* có độ thực tế cao vào Docker volume **trước khi** malware/sample được thực thi.

Thuật ngữ "UBER-like" phản ánh cách tiếp cận: giống như Uber sinh dữ liệu người dùng giả lập để test, module này sinh dữ liệu hệ thống giả lập để che giấu sandbox.

### 1.3. Phạm vi thực hiện

| Hạng mục | Trạng thái |
|---|---|
| Kiến trúc plugin (ABC + Manager) | ✅ Hoàn thành |
| DNS Artifact Extension | ✅ Hoàn thành |
| HTTP/Browser Artifact Extension | ✅ Hoàn thành |
| User Profile system | ✅ Hoàn thành |
| Docker init container | ✅ Hoàn thành |
| PoC Test Suite (10 cases) | ✅ 10/10 PASS |
| Web Demo Integration | ✅ Hoàn thành |
| Docker Compose Integration | ✅ Hoàn thành |

---

## 2. KIẾN TRÚC HỆ THỐNG

### 2.1. Tổng quan kiến trúc

```
┌─────────────────────────────────────────────────────────────────┐
│                    DOCKER COMPOSE STACK                         │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  PHASE 1: Pre-execution (Init Container)                │    │
│  │                                                         │    │
│  │  ┌──────────────────────────────────────────────────┐   │    │
│  │  │           artifact-extension container           │   │    │
│  │  │                                                  │   │    │
│  │  │  entrypoint.py                                   │   │    │
│  │  │       │                                          │   │    │
│  │  │       ├── UserProfile (from JSON config)         │   │    │
│  │  │       ├── ArtifactManager                        │   │    │
│  │  │       │       ├── DNSArtifactExtension           │   │    │
│  │  │       │       └── HTTPArtifactExtension          │   │    │
│  │  │       ├── _emit_background_events()              │   │    │
│  │  │       └── inject_all() ──► /artifacts/ volume    │   │    │
│  │  │                                                  │   │    │
│  │  │  EXIT(0) ─────────────────────────────────────►  │   │    │
│  │  └──────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                           │                                     │
│                    artifacts_volume                             │
│                    /artifacts/                                  │
│                    ├── dns/cache.txt          (~157 entries)    │
│                    ├── browser/History        (SQLite, ~400v)   │
│                    └── browser/downloads/     (5 Zone.Id files) │
│                           │                                     │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  PHASE 2: Execution (Service containers see artifacts)  │    │
│  │                                                         │    │
│  │   inetsim ──► service-simulation ──► [sample/malware]   │    │
│  │                       │                                 │    │
│  │               mount /artifacts/ ◄── thấy artifacts      │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2. Cấu trúc thư mục module

```
service-simulation-module/artifact-extension/
├── artifact_extension/              # Core Python package
│   ├── __init__.py                  # Public exports
│   ├── base.py                      # ArtifactExtension ABC
│   ├── manager.py                   # ArtifactManager orchestrator
│   ├── profile.py                   # Statistical profile dataclasses
│   └── extensions/
│       ├── __init__.py
│       ├── dns_extension.py         # DNS resolver cache generator
│       └── http_extension.py        # Browser History + Zone.Identifier generator
├── tests/
│   └── test_poc_detector.py         # 10 PoC test cases
├── profiles/
│   └── normal_user.json             # Sample profile JSON
├── Dockerfile                       # Init container image
├── entrypoint.py                    # Docker entrypoint
└── requirements.txt                 # pytest>=7.4
```

### 2.3. Plugin Contract

Module được thiết kế theo **Plugin Pattern** — mỗi artifact type là một plugin độc lập:

```python
# base.py
class ArtifactExtension(ABC):

    @abstractmethod
    def on_service_event(self, event: dict) -> None:
        """
        Nhận event từ service simulation.
        Extension tự quyết định xử lý hay bỏ qua dựa trên event["type"].
        Event schema tối thiểu: {"type": str, "timestamp": float}
        """
        ...

    @abstractmethod
    def inject(self, artifacts_path: str) -> None:
        """
        Ghi tất cả artifacts đã tích luỹ ra artifacts_path.
        Mỗi extension tự tạo subdirectory riêng.
        """
        ...
```

**Lợi ích của thiết kế này:**
- Thêm artifact type mới không cần chạm vào code hiện có
- Extension không biết sự tồn tại của nhau
- Lỗi trong một extension không block các extension khác
- Dễ test đơn lẻ (unit test từng extension)

---

## 3. CHI TIẾT IMPLEMENTATION

### 3.1. ArtifactManager

**File:** `artifact_extension/manager.py`

```python
class ArtifactManager:
    def register(self, ext: ArtifactExtension) -> None:
        self._extensions.append(ext)

    def dispatch(self, event: dict) -> None:
        """Broadcast event đến tất cả extensions."""
        for ext in self._extensions:
            try:
                ext.on_service_event(event)
            except Exception as exc:
                logger.error("[%s] on_service_event failed: %s",
                             type(ext).__name__, exc)

    def inject_all(self) -> None:
        """Tạo thư mục và gọi inject() trên từng extension."""
        Path(self._artifacts_path).mkdir(parents=True, exist_ok=True)
        for ext in self._extensions:
            try:
                ext.inject(self._artifacts_path)
                logger.info("  [OK] %s", type(ext).__name__)
            except Exception as exc:
                logger.error("  [FAIL] %s: %s", type(ext).__name__, exc)
```

### 3.2. User Profile System

**File:** `artifact_extension/profile.py`

Toàn bộ số lượng và phân phối artifacts được kiểm soát qua profile — không hardcode:

```python
@dataclass
class DNSProfile:
    entries: int = 150             # Tổng số DNS cache entries
    negative_ratio: float = 0.05   # Tỷ lệ NXDOMAIN entries
    ttl_jitter_seconds: int = 3600 # Phạm vi jitter cho TTL

@dataclass
class BrowserProfile:
    history_entries: int = 400     # Số visits trong SQLite
    days_of_history: int = 14      # Trải đều timestamps qua bao nhiêu ngày
    transition_weights: dict = {   # Xác suất từng loại visit
        "LINK": 0.45,
        "TYPED": 0.20,
        "AUTO_BOOKMARK": 0.05,
        "AUTO_SUBFRAME": 0.15,
        "FORM_SUBMIT": 0.10,
        "RELOAD": 0.05,
    }
```

Profile được load từ JSON tại `/config/artifact_profile.json` trong Docker:

```json
{
    "user_type": "normal",
    "dns": { "entries": 150, "negative_ratio": 0.05, "ttl_jitter_seconds": 3600 },
    "browser": { "history_entries": 400, "days_of_history": 14 }
}
```

### 3.3. DNS Artifact Extension

**File:** `artifact_extension/extensions/dns_extension.py`  
**Output:** `<artifacts_path>/dns/cache.txt`

**Kỹ thuật chống fingerprinting:**

**a) TTL jitter — tránh "bulk injection" tell:**
```python
# TTL của từng entry được chọn ngẫu nhiên trong khoảng rộng
# → entries trông như được query ở các thời điểm khác nhau
ttl = random.randint(1, profile.dns.ttl_jitter_seconds)  # 1 đến 3600 giây
```

**b) NXDOMAIN entries — negative cache:**
```python
# 5% entries là NXDOMAIN (stale ad/tracker domains)
# → hành vi tự nhiên của máy thật đã truy cập nhiều website
neg_count = max(1, int(p.entries * p.negative_ratio))
```

**c) random.choices với replacement — vượt quá pool size:**
```python
# Pool chỉ có 30 domains nhưng profile yêu cầu 150 entries
# → sampling with replacement tạo ra entries hợp lý
sampled = random.choices(pool, k=positive_count)
```

**Format output** (mimics `ipconfig /displaydns` trên Windows):
```
Windows IP Configuration

    DNS Resolver Cache
    -------------------

    Record Name . . . . . : google.com
    Record Type . . . . . : 1
    Time To Live  . . . . : 2847
    Data Length . . . . . : 4
    Section . . . . . . . : Answer
    A (Host) Record . . . : 142.250.80.46

    Record Name . . . . . : tracker.old-analytics.invalid
    Record Type . . . . . : 0
    Time To Live  . . . . : 347
    Data Length . . . . . : 0
    Section . . . . . . . : Answer
    (No records)
```

**Kết quả thực tế:** 157 DNS entries sau khi inject.

### 3.4. HTTP Artifact Extension

**File:** `artifact_extension/extensions/http_extension.py`

#### 3.4.1. Browser History SQLite Database

**Output:** `<artifacts_path>/browser/History`  
**Schema:** Tương thích Chrome (Chrome sử dụng định dạng này trên mọi OS)

```sql
CREATE TABLE urls (
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL,
    title TEXT DEFAULT '',
    visit_count INTEGER DEFAULT 0,
    last_visit_time INTEGER NOT NULL,  -- Chrome epoch (microseconds từ 1601-01-01)
    hidden INTEGER DEFAULT 0
);

CREATE TABLE visits (
    id INTEGER PRIMARY KEY,
    url INTEGER NOT NULL REFERENCES urls(id),
    visit_time INTEGER NOT NULL,       -- Chrome epoch
    transition INTEGER DEFAULT 0,      -- LINK=0, TYPED=1, AUTO_BOOKMARK=2...
    visit_duration INTEGER DEFAULT 0   -- microseconds
);
```

**Kỹ thuật quan trọng — Chrome Epoch:**

```python
# Microseconds giữa 1601-01-01 (Chrome epoch) và 1970-01-01 (Unix epoch)
_CHROME_EPOCH_OFFSET_US: int = 11_644_473_600 * 1_000_000

def _unix_to_chrome_time(unix_ts: float) -> int:
    return int(unix_ts * 1_000_000) + _CHROME_EPOCH_OFFSET_US
```

> **Lỗi phổ biến:** Dùng Unix epoch cho Chrome History. Timestamp sẽ nhỏ hơn giá trị
> đúng khoảng 116 năm — malware kiểm tra range timestamp sẽ detect ngay.

**Kỹ thuật chống fingerprinting:**

```python
# Timestamps phân phối theo exponential (skew toward recent)
# → mô phỏng hành vi browse tự nhiên: nhiều activity gần đây, ít hơn xa hơn
interval = random.expovariate(1.0 / avg_gap)

# Visit duration ngẫu nhiên 5s–5min (tránh duration=0 tell)
duration_us = random.randint(5_000_000, 300_000_000)

# Transition types theo weighted sampling từ profile
transition_name = random.choices(
    list(weights.keys()),
    weights=list(weights.values())
)[0]
```

#### 3.4.2. Zone.Identifier Files (Mark-of-the-Web)

**Output:** `<artifacts_path>/browser/downloads/<filename>.Zone.Identifier`

Trên Windows, mọi file tải từ Internet đều nhận NTFS Alternate Data Stream này:

```ini
[ZoneTransfer]
ZoneId=3
ReferrerUrl=https://www.google.com/
HostUrl=https://cdn.discordapp.com/apps/DiscordSetup.exe
```

| Field | Ý nghĩa |
|---|---|
| `ZoneId=3` | Internet Zone — file từ internet |
| `ReferrerUrl` | Trang web dẫn đến download |
| `HostUrl` | URL download trực tiếp |

**Background downloads được inject:**
- `setup.exe` — GitHub release
- `ChromeSetup.exe` — Google Chrome
- `DiscordSetup.exe` — Discord
- `VSCodeSetup-x64.exe` — VS Code
- `requests-2.31.0.tar.gz` — Python package

**Kết quả thực tế:** 5 Zone.Identifier files sau khi inject.

### 3.5. Docker Entrypoint & Background Events

**File:** `entrypoint.py`

```python
def _emit_background_events(manager: ArtifactManager) -> None:
    now = time.time()
    hour = 3600.0

    # DNS queries ~4 giờ trước
    for domain, ip in background_dns:
        manager.dispatch({
            "type": "dns_query",
            "domain": domain,
            "resolved_ip": ip,
            "timestamp": now - (4 * hour),
        })

    # HTTP browsing ~2 giờ trước
    for url, title, transition in background_http:
        manager.dispatch({
            "type": "http_request",
            "url": url,
            "title": title,
            "transition": transition,
            "timestamp": now - (2 * hour),
        })

    # File downloads ~3 giờ trước
    for download_url, referrer, filename in background_downloads:
        manager.dispatch({
            "type": "file_download",
            "url": download_url,
            "referrer_url": f"https://{referrer}/",
            "filename": filename,
            "timestamp": now - (3 * hour),
        })
```

**Lý do offset timestamps về quá khứ:** Artifacts cần trông như tích luỹ tự nhiên
qua thời gian — không phải tất cả được tạo đúng lúc container start.

---

## 4. DOCKER INTEGRATION

### 4.1. Init Container Pattern

```yaml
# docker-compose.network-sim.yml

services:
  # PHASE 1: Init container
  artifact-extension:
    build: ../service-simulation-module/artifact-extension
    container_name: pack-a-mal-artifact-ext
    volumes:
      - artifacts_volume:/artifacts
      - ../service-simulation-module/shared/config:/config
    environment:
      - ARTIFACTS_PATH=/artifacts
      - PROFILE_PATH=/config/artifact_profile.json
    restart: "no"     # ← chạy một lần, không restart

  # PHASE 2: Service containers
  service-simulation:
    depends_on:
      inetsim:
        condition: service_healthy
      artifact-extension:
        condition: service_completed_successfully  # ← chờ exit(0)
    volumes:
      - artifacts_volume:/artifacts   # ← thấy artifacts đã inject

volumes:
  artifacts_volume:
    name: pack-a-mal-artifacts
```

### 4.2. Startup Order

```
t=0s   inetsim          → Starting
t=5s   inetsim          → Healthy (healthcheck pass)
t=5s   artifact-ext     → Starting
t=6s   artifact-ext     → Generating DNS cache (157 entries)...
t=7s   artifact-ext     → Generating Browser History (400 visits)...
t=7s   artifact-ext     → Generating Zone.Identifier (5 files)...
t=7s   artifact-ext     → EXIT(0) ✅
t=7s   service-sim      → Starting (artifact-ext completed_successfully)
t=10s  service-sim      → Running, /artifacts/ populated
```

### 4.3. Bài học quan trọng — Network Conflict

**Lỗi xảy ra ban đầu:**
```
Error response from daemon: failed to set up container networking:
Address already in use
```

**Nguyên nhân:** `artifact-extension` được khai báo join vào `pack_a_mal_network`.
Init container chiếm địa chỉ IP trong subnet `172.20.0.0/24`, gây conflict
với `inetsim` đang dùng static IP `172.20.0.2`.

**Fix:** Xóa hoàn toàn `networks:` khỏi `artifact-extension`. Init container
không cần network access — nó chỉ ghi file vào volume. Đây cũng là
best practice bảo mật: init container với quyền tối thiểu.

---

## 5. POC SANDBOX DETECTOR & TEST SUITE

### 5.1. Mô hình kiểm chứng

Để xác minh artifacts đủ thuyết phục, module bao gồm một **PoC Sandbox Detector**
mô phỏng các kiểm tra mà malware fingerprinting routine thực hiện:

```python
class SandboxDetector:
    """Safe PoC — chỉ kiểm tra file artifacts, không có network/privilege."""

    def check_dns_cache_missing(self) -> bool:
        return not (self._root / "dns" / "cache.txt").exists()

    def check_dns_cache_sparse(self) -> bool:
        """< 10 entries — suspiciously bare"""
        return cache.read_text().count("Record Name") < 10

    def check_dns_ttl_uniform(self) -> bool:
        """TTL spread < 60s — bulk injection fingerprint"""
        return (max(ttls) - min(ttls)) < 60

    def check_browser_history_missing(self) -> bool:
        return not (self._root / "browser" / "History").exists()

    def check_browser_history_sparse(self) -> bool:
        """< 20 URLs — nearly empty"""
        return url_count < 20

    def check_transition_uniform(self) -> bool:
        """< 2 distinct transition types — synthetic generation tell"""
        return distinct_transitions < 2

    def is_sandbox(self) -> bool:
        return any([
            self.check_dns_cache_missing(),
            self.check_dns_cache_sparse(),
            self.check_dns_ttl_uniform(),
            self.check_browser_history_missing(),
            self.check_browser_history_sparse(),
            self.check_transition_uniform(),
        ])
```

### 5.2. Ba kịch bản kiểm thử

| Case | Setup | Expected |
|---|---|---|
| **A** | Sandbox sạch — không có artifact nào | `is_sandbox() == True` |
| **B** | Có service events nhưng chưa gọi `inject_all()` | `is_sandbox() == True` |
| **C** | Service events + `inject_all()` hoàn tất | `is_sandbox() == False` |

**Case C** là mục tiêu chính — chứng minh rằng sau khi inject, sandbox
**không còn bị detect** bởi các kiểm tra fingerprinting.

### 5.3. Kết quả Test Suite

```
============================= test session starts =============================
collected 10 items

tests/test_poc_detector.py::test_case_a_clean_sandbox_detected          PASSED
tests/test_poc_detector.py::test_case_a_individual_checks               PASSED
tests/test_poc_detector.py::test_case_b_service_only_still_detected     PASSED
tests/test_poc_detector.py::test_case_c_artifact_injection_bypasses_detection PASSED
tests/test_poc_detector.py::test_case_c_dns_checks_pass                 PASSED
tests/test_poc_detector.py::test_case_c_browser_checks_pass             PASSED
tests/test_poc_detector.py::test_artifact_freshness_dns                 PASSED
tests/test_poc_detector.py::test_zone_identifier_created_for_download   PASSED
tests/test_poc_detector.py::test_zone_identifier_fields_populated       PASSED
tests/test_poc_detector.py::test_chrome_timestamps_in_correct_epoch     PASSED

============================= 10 passed in 0.55s ==============================
```

**10/10 tests PASS.**

### 5.4. Phân tích từng nhóm test

| Test case | Mục tiêu test | Test như thế nào | Kết quả mong đợi |
|---|---|---|---|
| `test_case_a_clean_sandbox_detected` | Xác nhận detector phát hiện sandbox khi chưa có artifact | Chạy `is_sandbox()` trên thư mục rỗng | Trả về `True` |
| `test_case_a_individual_checks` | Kiểm tra từng tín hiệu sandbox hoạt động độc lập | Gọi 6 individual checks trên môi trường sạch | Tất cả đều trả về `True` |
| `test_case_b_service_only_still_detected` | Chứng minh service events thôi chưa đủ | Dispatch events nhưng không gọi `inject_all()` | Vẫn bị detect |
| `test_case_c_artifact_injection_bypasses_detection` | Xác minh artifact đầy đủ có thể bypass detection | Chạy inject đầy đủ rồi kiểm tra lại detector | Trả về `False` |
| `test_case_c_dns_checks_pass` | Kiểm tra DNS artifact được sinh đúng mức tối thiểu | Đọc DNS cache sau khi inject | Có cache, ≥ 10 entries, TTL spread ≥ 60s |
| `test_case_c_browser_checks_pass` | Kiểm tra browser artifact hợp lệ | Đọc History SQLite sau khi inject | Có history, ≥ 20 URLs, ≥ 2 transition types |
| `test_artifact_freshness_dns` | Đảm bảo mỗi lần inject không bị deterministic | So sánh TTL giữa hai lần inject | TTL khác nhau |
| `test_zone_identifier_created_for_download` | Xác minh file download sinh Zone.Identifier | Trigger `file_download` event và kiểm tra output | Mỗi event tạo 1 file |
| `test_zone_identifier_fields_populated` | Kiểm tra metadata download được điền đủ | Mở file Zone.Identifier sau inject | Có `ZoneId`, `ReferrerUrl`, `HostUrl` |
| `test_chrome_timestamps_in_correct_epoch` | Xác minh Chrome timestamp đúng epoch | Chuyển timestamp Unix sang Chrome format | Offset được áp dụng đúng |

### 5.5. Test Samples and Baseline (English addendum)

The evaluation uses **three representative test samples**, all created in-house and stored under `dynamic-analysis/sample_packages/malicious_network_package/`:

- [test_network.py](https://github.com/DangTheNhan/EIU-Chat-Zone/blob/main/dynamic-analysis/sample_packages/malicious_network_package/test_network.py)
- [test_with_inetsim.py](https://github.com/DangTheNhan/EIU-Chat-Zone/blob/main/dynamic-analysis/sample_packages/malicious_network_package/test_with_inetsim.py)
- [test_full_mode.py](https://github.com/DangTheNhan/EIU-Chat-Zone/blob/main/dynamic-analysis/sample_packages/malicious_network_package/test_full_mode.py)

These samples were **synthetic test scripts**, not real malware samples collected from the wild. They were authored to emulate common malicious network behaviors in a controlled and reproducible way, so that the impact of network simulation and artifact emulation could be observed under consistent conditions.

The samples were selected to cover three execution scenarios:

1. **Direct execution without simulation**
2. **Execution with INetSim redirection**
3. **Execution with the stricter Full Mode interception policy**

The baseline for comparison is the **non-simulated environment**, meaning:

- no network simulation
- no artifact injection
- direct execution only

Under this baseline, the samples are expected to fail early because DNS resolution and HTTP requests are not intercepted, and no persistent host artifacts are available.

For the enhanced configurations, the samples are executed with:

- **Network simulation only**: traffic is redirected through INetSim, but the host environment remains minimal
- **Network simulation + artifact emulation**: INetSim is combined with the artifact extension so that the environment also contains realistic persistent artifacts

This sample design allows the evaluation to answer three practical questions:

- How many execution paths become active when simulation is enabled?
- Does the resulting trace become more coherent and easier to interpret?
- Are repeated runs stable under the same configuration?

In short, the sample set is intentionally small, controlled, and reproducible. Its purpose is to provide a reliable benchmark for evaluating sandbox realism rather than to represent a large malware corpus.

---

## 6. WEB DEMO INTEGRATION

### 6.1. Routes bổ sung vào Flask app

**File:** `web_demo/app.py`

| Route | Method | Chức năng |
|---|---|---|
| `/stream/artifact/docker-inject` | GET (SSE) | Stream output của `docker-compose run --rm artifact-extension` |
| `/api/artifact/poc` | GET | Chạy pytest, parse kết quả theo 3 case, trả JSON |
| `/api/artifact/summary` | GET | Đọc `demo_artifacts/`, trả thống kê DNS/Browser/Downloads |

**SSE Stream route:**
```python
@app.route("/stream/artifact/docker-inject")
def artifact_docker_inject():
    compose = DYNAMIC_ANALYSIS_DIR / "docker-compose.network-sim.yml"
    def gen():
        yield from stream_cmd(
            f'docker-compose -f "{compose}" run --rm artifact-extension',
            env_extra={"COMPOSE_PROGRESS": "plain", "NO_COLOR": "1"},
        )
        r = run_cmd('docker volume inspect pack-a-mal-artifacts')
        if r["ok"]:
            yield f"data: {json.dumps('✅ Volume pack-a-mal-artifacts tồn tại')}\n\n"
    return make_sse_response(gen)
```

**PoC Test route** (parse pytest output):
```python
@app.route("/api/artifact/poc")
def artifact_poc():
    result = run_cmd(
        f'"{VENV_PYTHON}" -m pytest tests/test_poc_detector.py -v --tb=short',
        cwd=str(ARTIFACT_EXT_DIR),
        timeout=60,
    )
    # Parse: "test_poc_detector.py::test_name PASSED/FAILED"
    # Nhóm theo case_a / case_b / case_c / extra
    ...
```

### 6.2. UI Card

Card mới "UBER-like Artifact Extension" trên dashboard gồm:

- **Inject Docker** (btn-primary) — trigger injection, stream log trực tiếp
- **Summary** — hiển thị stats: DNS entries, History URLs, Zone.Identifier count
- **Chạy PoC Tests** — 4 indicator (Case A / B / C / Extra) cập nhật realtime

---

## 7. ĐÁNH GIÁ KỸ THUẬT

### 7.1. Điểm mạnh

| Điểm mạnh | Chi tiết |
|---|---|
| **Kiến trúc mở rộng** | Plugin pattern — thêm artifact mới không đụng code cũ |
| **Không phụ thuộc ngoài** | Chỉ dùng stdlib Python (sqlite3, pathlib, random, struct) |
| **Profile-driven** | Mọi thông số cấu hình qua JSON — không hardcode |
| **Randomisation** | TTL jitter, timestamp exponential distribution, transition weights |
| **Chrome epoch đúng** | Offset 11,644,473,600 giây từ 1601-01-01 được xử lý chính xác |
| **Init container pattern** | Tách biệt hoàn toàn với service runtime, minimal privilege |
| **Test coverage** | 10 PoC cases kiểm tra end-to-end behavior |

### 7.2. Hạn chế hiện tại

| Hạn chế | Tác động | Hướng giải quyết |
|---|---|---|
| **Windows-centric artifacts** | Zone.Identifier là NTFS ADS, DNS format là `ipconfig /displaydns` | Xây dựng Linux extension set (xem Section 8) |
| **Static domain pool** | 30 domains hardcode trong `_COMMON_DOMAINS` | Mở rộng pool hoặc load từ file external |
| **Không có Prefetch** | Windows Prefetch files bị một số fingerprinters kiểm tra | Tier 2 extension |
| **Chromium on Linux path** | `~/.config/chromium/Default/History` chưa được map đúng | Mount path cần cấu hình theo OS |
| **Summary đọc local** | `/api/artifact/summary` đọc `demo_artifacts/` — không phản ánh Docker volume | Cần strategy đọc volume hoặc dùng local inject cho demo |

### 7.3. Kết quả đo lường

| Metric | Giá trị |
|---|---|
| Thời gian inject (Docker container) | < 2 giây |
| DNS entries sinh ra | 157 (profile = 150, overhead từ event entries) |
| Browser visits | 400 (đúng profile) |
| Distinct URLs | ~25 (từ background + event pool) |
| Zone.Identifier files | 5 (đúng số background_downloads) |
| TTL spread | 1 – 3600 giây (uniform random) |
| Test execution time | 0.55 giây (10 cases) |
| Docker image size | ~150MB (python:3.11-slim base) |

---

## 8. HƯỚNG PHÁT TRIỂN TIẾP THEO — LINUX ARTIFACTS

Phần việc tiếp theo (bàn giao cho đồng nghiệp) là xây dựng bộ extensions tương đương cho Linux.

### 8.1. Tier 1 — Ưu tiên cao

| Extension | Artifact | Path |
|---|---|---|
| `BashHistoryExtension` | `~/.bash_history` | Text, mỗi dòng 1 command |
| `FirefoxHistoryExtension` | `places.sqlite` | SQLite, schema `moz_places` + `moz_historyvisits` |
| `RecentlyUsedExtension` | `recently-used.xbel` | XML (GTK bookmark format) |
| `LinuxDNSExtension` | `/etc/hosts` hoặc nscd cache | Text hoặc binary |

### 8.2. Lưu ý quan trọng cho Linux

```
Firefox epoch  ≠  Chrome epoch
Firefox dùng microseconds từ Unix epoch (1970-01-01)
Chrome dùng microseconds từ Windows epoch (1601-01-01)
```

### 8.3. Cách thêm extension mới

```python
# Tạo file mới: artifact_extension/extensions/bash_history_extension.py

class BashHistoryExtension(ArtifactExtension):
    def on_service_event(self, event: dict) -> None:
        if event.get("type") == "shell_command":
            self._events.append(event)

    def inject(self, artifacts_path: str) -> None:
        out = Path(artifacts_path) / "home" / ".bash_history"
        out.parent.mkdir(parents=True, exist_ok=True)
        lines = random.choices(_COMMON_COMMANDS, k=self._profile.dns.entries)
        out.write_text("\n".join(lines) + "\n")
```

Sau đó thêm vào `entrypoint.py`:
```python
from artifact_extension.extensions.bash_history_extension import BashHistoryExtension
manager.register(BashHistoryExtension(profile))
```

---

## 9. HƯỚNG DẪN SỬ DỤNG

### 9.1. Chạy toàn bộ stack

```powershell
# Start Docker services
cd d:\Project-EIU\pack-a-mal\dynamic-analysis
docker-compose -f docker-compose.network-sim.yml up -d

# Start web demo
cd d:\Project-EIU\pack-a-mal\web_demo
.\venv\Scripts\python.exe app.py
# → http://127.0.0.1:5500
```

### 9.2. Chạy PoC tests

```powershell
cd d:\Project-EIU\pack-a-mal\service-simulation-module\artifact-extension
python -m pytest tests/ -v
# Expected: 10/10 PASS
```

### 9.3. Re-inject sau khi thay đổi code

```powershell
# Rebuild image
cd d:\Project-EIU\pack-a-mal\dynamic-analysis
docker-compose -f docker-compose.network-sim.yml build artifact-extension

# Re-inject vào volume
docker-compose -f docker-compose.network-sim.yml run --rm artifact-extension

# Kiểm tra kết quả
docker run --rm -v pack-a-mal-artifacts:/art alpine sh -c ^
  "grep -c 'Record Name' /art/dns/cache.txt; find /art -name '*.Zone.Identifier'"
```

### 9.4. Xử lý lỗi "Address already in use"

```powershell
# Dọn stale containers và network
docker-compose -f docker-compose.network-sim.yml down --remove-orphans

# Khởi động lại
docker-compose -f docker-compose.network-sim.yml up -d
```

---

## 10. KẾT LUẬN

Module **UBER-like Artifact Extension** đã được implement thành công với:

- **Kiến trúc plugin** rõ ràng, dễ mở rộng
- **2 artifact extensions** (DNS + HTTP/Browser) hoạt động đúng
- **10/10 PoC test cases** pass, xác minh artifact đủ thuyết phục để bypass sandbox detection
- **Docker init container** tích hợp vào compose stack không xung đột
- **Web demo** với SSE streaming và PoC indicators realtime

Kết quả thực đo trên Docker volume sau khi inject:
```
✅  DNS cache:           157 entries  (TTL spread: 1–3600s)
✅  Browser History:     400 visits   (25 distinct URLs, 6 transition types)
✅  Zone.Identifier:     5 files      (ZoneId=3, ReferrerUrl, HostUrl)
✅  PoC Detector:        is_sandbox() == False  (Case C)
```

Sandbox fingerprinting bị bypass hoàn toàn theo các tiêu chí đã định nghĩa.  
Phần tiếp theo: xây dựng bộ Linux artifacts tương đương (chi tiết tại `HANDOFF_LINUX_ARTIFACTS.md`).
