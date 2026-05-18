# Handoff: UBER-like Artifact Extension — Tiếp nối phần Linux

> **Mục tiêu bàn giao:** Tổng hợp những gì đã xây dựng (Windows-focused), làm rõ kiến trúc,
> và chỉ rõ phần việc tiếp theo là cài đặt/tích hợp/thử nghiệm các công cụ tạo fake artifacts
> cho môi trường **Linux**.

---

## 1. Bối cảnh & Mục tiêu

### Vấn đề đặt ra

Malware thông minh kiểm tra môi trường trước khi chạy payload thật. Nếu phát hiện sandbox,
nó sẽ **tự huỷ / không làm gì**. Các dấu hiệu sandbox phổ biến:

| Dấu hiệu bị kiểm tra | Ý nghĩa |
|---|---|
| DNS cache trống hoặc quá ít entry | Máy thật luôn có hàng trăm DNS queries |
| Browser history không tồn tại | Người dùng thật có lịch sử duyệt web |
| TTL của tất cả DNS entry đều nhau | Inject giả tạo theo batch |
| Không có file download nào | Không ai dùng máy mà chưa tải gì |
| Transition type trong browser history đồng nhất | Lịch sử bị sinh tự động |

### Giải pháp đã xây dựng

**UBER-like Artifact Extension** — một Python module chạy như **init container** trong Docker,
inject các persistent system artifacts (DNS cache, browser history, Zone.Identifier files)
vào một Docker volume **trước khi** malware/sample chạy, giúp sandbox trông như máy thật.

---

## 2. Cấu trúc thư mục

```
service-simulation-module/artifact-extension/
├── artifact_extension/              # Core Python package
│   ├── __init__.py                  # Exports: ArtifactExtension, ArtifactManager, UserProfile
│   ├── base.py                      # Abstract base class ArtifactExtension
│   ├── manager.py                   # ArtifactManager: điều phối tất cả extensions
│   ├── profile.py                   # Statistical profile dataclasses
│   └── extensions/
│       ├── __init__.py              # Exports: DNSArtifactExtension, HTTPArtifactExtension
│       ├── dns_extension.py         # Sinh DNS cache text file
│       └── http_extension.py        # Sinh Chrome History SQLite + Zone.Identifier
├── tests/
│   └── test_poc_detector.py         # 10 PoC test cases (ALL PASS)
├── profiles/
│   └── normal_user.json             # Sample profile JSON
├── Dockerfile                       # Init container image
├── entrypoint.py                    # Docker entrypoint
└── requirements.txt                 # pytest>=7.4 (sqlite3 là stdlib)
```

---

## 3. Kiến trúc Plugin

### Abstract Base Class (`base.py`)

```python
class ArtifactExtension(ABC):
    @abstractmethod
    def on_service_event(self, event: dict) -> None:
        """Nhận event từ service simulation (dns_query, http_request, file_download...)"""
        ...

    @abstractmethod
    def inject(self, artifacts_path: str) -> None:
        """Ghi artifact ra volume path (ví dụ: /artifacts)"""
        ...
```

**Mỗi extension mới** chỉ cần implement 2 method này. Plugin contract rõ ràng — extension
không biết sự tồn tại của nhau, ArtifactManager lo broadcast.

### ArtifactManager (`manager.py`)

```python
manager = ArtifactManager(artifacts_path="/artifacts", profile=profile)
manager.register(DNSArtifactExtension(profile))
manager.register(HTTPArtifactExtension(profile))
manager.dispatch({"type": "dns_query", "domain": "example.com", "timestamp": time.time()})
manager.inject_all()   # tạo thư mục + gọi inject() trên từng extension
```

Lỗi trong một extension **không block** các extension khác (try/except per extension).

### User Profile (`profile.py`)

```python
# Kiểm soát số lượng và độ phân tán của artifacts
UserProfile(
    user_type="normal",
    dns=DNSProfile(entries=150, negative_ratio=0.05, ttl_jitter_seconds=3600),
    browser=BrowserProfile(history_entries=400, days_of_history=14,
                           transition_weights={"LINK":0.45,"TYPED":0.20,...})
)
```

Profile được load từ JSON tại `/config/artifact_profile.json` trong Docker.

---

## 4. Các Extensions đã implement (Windows-focused)

### 4.1 DNS Extension (`dns_extension.py`)

**Output:** `<artifacts_path>/dns/cache.txt`

**Format:** Mimics `ipconfig /displaydns` trên Windows:
```
    Record Name . . . . . : google.com
    Record Type . . . . . : 1
    Time To Live  . . . . : 3247
    Data Length . . . . . : 4
    Section . . . . . . . : Answer
    A (Host) Record . . . : 142.250.80.46
```

**Kỹ thuật chống fingerprinting:**
- TTL được jitter ngẫu nhiên để các entry trông như được query ở thời điểm khác nhau
- Mix NXDOMAIN entries (negative cache) theo configurable ratio
- Event-driven entries từ DNS queries thực tế của service simulation

### 4.2 HTTP Extension (`http_extension.py`)

**Output 1:** `<artifacts_path>/browser/History` — SQLite, Chrome-compatible schema

**Schema Chrome:**
```sql
CREATE TABLE urls (id, url, title, visit_count, last_visit_time, ...);
CREATE TABLE visits (id, url, visit_time, transition, visit_duration, ...);
```

> ⚠️ **Chrome epoch = microseconds since 1601-01-01** (không phải Unix epoch)
> Conversion: `chrome_time = (unix_time + 11_644_473_600) * 1_000_000`
> Đây là lỗi phổ biến nhất trong fake history — dùng sai epoch bị detect ngay.

**Kỹ thuật chống fingerprinting:**
- Timestamps phân phối theo exponential distribution (skew toward recent)
- Visit duration ngẫu nhiên 5s–5min (tránh tell của duration=0)
- Transition types lấy từ weighted sampling: LINK(45%), TYPED(20%), FORM_SUBMIT(10%)...

**Output 2:** `<artifacts_path>/browser/downloads/<name>.Zone.Identifier`

**Format Zone.Identifier (Windows NTFS ADS):**
```ini
[ZoneTransfer]
ZoneId=3
ReferrerUrl=https://www.google.com/
HostUrl=https://cdn.example.com/file.zip
```

---

## 5. Docker Integration

### Init Container Pattern

```yaml
# docker-compose.network-sim.yml
services:
  artifact-extension:
    build: ../service-simulation-module/artifact-extension
    container_name: pack-a-mal-artifact-ext
    volumes:
      - artifacts_volume:/artifacts
      - ../service-simulation-module/shared/config:/config
    environment:
      - ARTIFACTS_PATH=/artifacts
      - PROFILE_PATH=/config/artifact_profile.json
    restart: "no"           # init-container: chạy một lần rồi thoát
    # KHÔNG có "networks:" — init container không cần network access

  service-simulation:
    depends_on:
      inetsim:
        condition: service_healthy
      artifact-extension:
        condition: service_completed_successfully   # chờ artifacts xong mới start
    volumes:
      - artifacts_volume:/artifacts    # thấy artifacts đã inject

volumes:
  artifacts_volume:
    name: pack-a-mal-artifacts
```

**Startup order:** `inetsim (healthy)` → `artifact-extension (exit 0)` → `service-simulation`

> ⚠️ **Lưu ý quan trọng:** Đừng thêm `networks:` vào `artifact-extension`.
> Nếu join network, init container chiếm địa chỉ IP gây conflict với `inetsim`
> → "Address already in use" error.

### Re-inject on demand (Web Demo)

```bash
docker-compose -f docker-compose.network-sim.yml run --rm artifact-extension
```

---

## 6. PoC Test Suite — Kết quả

Chạy: `python -m pytest tests/ -v`

| Test Case | Mô tả | Kết quả |
|---|---|---|
| `test_case_a_clean_sandbox_detected` | Sandbox sạch → bị detect | ✅ PASS |
| `test_case_b_service_only_still_detected` | Có service events nhưng chưa inject → vẫn bị detect | ✅ PASS |
| `test_case_c_artifact_injection_bypasses_detection` | Sau inject → không bị detect | ✅ PASS |
| `test_case_c_dns_checks_pass` | DNS cache đủ entry, TTL jitter đúng | ✅ PASS |
| `test_case_c_browser_checks_pass` | Browser history đủ URL, transition varied | ✅ PASS |
| `test_artifact_freshness_dns` | Hai lần inject sinh TTL khác nhau | ✅ PASS |
| `test_zone_identifier_created_for_download` | file_download → tạo .Zone.Identifier | ✅ PASS |
| `test_zone_identifier_fields_populated` | ZoneId, ReferrerUrl, HostUrl đều có | ✅ PASS |
| `test_chrome_timestamps_in_correct_epoch` | Chrome epoch đúng (1601-01-01) | ✅ PASS |
| `test_multiple_extensions_independent` | Các extensions hoạt động độc lập | ✅ PASS |

**10/10 PASS**

---

## 7. Web Demo Integration

**File:** `web_demo/app.py` + `web_demo/templates/index.html`

| Route | Chức năng |
|---|---|
| `GET /stream/artifact/docker-inject` | SSE stream: chạy `docker-compose run --rm artifact-extension` |
| `GET /api/artifact/poc` | Chạy pytest, parse kết quả, trả JSON `{case_a, case_b, case_c}` |
| `GET /api/artifact/summary` | Đọc `demo_artifacts/`, đếm DNS entries, query History SQLite |

UI card có 2 button: **Inject Docker** và **Summary** + 4 PoC indicators (A/B/C/extra).

---

## 8. PHẦN VIỆC TIẾP THEO — Linux Artifacts

### 8.1 Bối cảnh

Các artifacts đã implement mang màu **Windows** (Zone.Identifier là NTFS ADS, `ipconfig /displaydns` format).
Malware trên Linux kiểm tra các artifact khác hoàn toàn. Cần một bộ extensions Linux song song.

### 8.2 Artifacts quan trọng cần implement cho Linux

#### Tier 1 — Ưu tiên cao (hay bị kiểm tra nhất)

| Artifact | Đường dẫn | Công cụ/thư viện gợi ý |
|---|---|---|
| **nscd DNS cache** | `/var/cache/nscd/hosts` | Binary format — parse/generate thủ công hoặc dùng `nscd` thật |
| **systemd-resolved cache** | `resolvectl statistics` output hoặc `/run/systemd/resolve/` | Fake bằng file text |
| **Bash history** | `~/.bash_history` | Sinh chuỗi command realistic theo profile |
| **Firefox SQLite History** | `~/.mozilla/firefox/*.default/places.sqlite` | Schema khác Chrome — `moz_places` + `moz_historyvisits` |
| **Recently used files** | `~/.local/share/recently-used.xbel` | XML format (GTK bookmark file) |

#### Tier 2 — Ưu tiên trung bình

| Artifact | Đường dẫn | Ghi chú |
|---|---|---|
| **Chromium History** (Linux) | `~/.config/chromium/Default/History` | Cùng schema với Chrome Windows — **có thể tái dùng `http_extension.py`** |
| **apt/dpkg logs** | `/var/log/apt/history.log`, `/var/log/dpkg.log` | Text format, sinh lịch sử install packages |
| **systemd journal** | `/run/log/journal/` | Binary format — phức tạp, có thể skip |
| **utmp/wtmp** (login history) | `/var/log/wtmp` | Binary struct format — dùng `struct` module |
| **Thumbnail cache** | `~/.cache/thumbnails/` | PNG files + `.thumbnails/` |

#### Tier 3 — Nâng cao (nếu có thời gian)

| Artifact | Ghi chú |
|---|---|
| **NetworkManager connection profiles** | `/etc/NetworkManager/system-connections/` |
| **SSH known_hosts** | `~/.ssh/known_hosts` |
| **Xfce/GNOME recent documents** | Desktop environment specific |

### 8.3 Cách thêm extension mới (ví dụ: BashHistoryExtension)

```python
# artifact_extension/extensions/bash_history_extension.py

from artifact_extension.base import ArtifactExtension
from artifact_extension.profile import UserProfile

_COMMON_COMMANDS = [
    "ls -la", "cd ~", "git status", "python3 --version",
    "sudo apt update", "curl https://example.com", ...
]

class BashHistoryExtension(ArtifactExtension):
    def __init__(self, profile: UserProfile):
        self._profile = profile
        self._events: list[dict] = []

    def on_service_event(self, event: dict) -> None:
        if event.get("type") == "shell_command":
            self._events.append(event)

    def inject(self, artifacts_path: str) -> None:
        out_dir = Path(artifacts_path) / "home" / ".bash_history"
        out_dir.parent.mkdir(parents=True, exist_ok=True)
        lines = random.choices(_COMMON_COMMANDS, k=self._profile.dns.entries)
        out_dir.write_text("\n".join(lines) + "\n")
```

Sau đó đăng ký trong `entrypoint.py`:
```python
from artifact_extension.extensions.bash_history_extension import BashHistoryExtension
manager.register(BashHistoryExtension(profile))
```

### 8.4 Firefox places.sqlite — Schema cần implement

```sql
-- Khác với Chrome, Firefox dùng schema này:
CREATE TABLE moz_places (
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL,
    title TEXT,
    visit_count INTEGER DEFAULT 0,
    last_visit_date INTEGER,   -- microseconds since Unix epoch (KHÔNG phải Chrome epoch!)
    frecency INTEGER DEFAULT -1
);

CREATE TABLE moz_historyvisits (
    id INTEGER PRIMARY KEY,
    place_id INTEGER REFERENCES moz_places(id),
    visit_date INTEGER,        -- microseconds since Unix epoch
    visit_type INTEGER         -- 1=LINK, 2=TYPED, 3=BOOKMARK, 4=EMBED, 5=REDIRECT_PERMANENT...
);
```

> Firefox dùng **microseconds since Unix epoch (1970-01-01)** — khác Chrome.

### 8.5 nscd DNS cache format

nscd lưu DNS cache ở binary format. Hai hướng tiếp cận:

**Hướng A (đơn giản):** Fake file text `/etc/hosts` với nhiều entries, không cần binary format.

**Hướng B (thực tế hơn):** Start container với `nscd` thật và feed DNS queries qua `getaddrinfo()`.

```dockerfile
# Trong Dockerfile của Linux artifact extension
RUN apt-get install -y nscd
COPY seed_dns.sh /app/
# seed_dns.sh dùng getent hosts để populate nscd cache
```

### 8.6 Kiểm tra file `recently-used.xbel`

```xml
<?xml version="1.0" encoding="UTF-8"?>
<xbel version="1.0"
      xmlns:bookmark="http://www.freedesktop.org/standards/desktop-bookmarks"
      xmlns:mime="http://www.freedesktop.org/standards/shared-mime-info">
  <bookmark href="file:///home/user/Documents/report.pdf"
            added="2026-05-10T09:23:00Z"
            modified="2026-05-10T09:23:00Z"
            visited="2026-05-15T14:10:00Z">
    <info>
      <metadata owner="http://freedesktop.org">
        <mime:mime-type type="application/pdf"/>
        <bookmark:applications>
          <bookmark:application name="Evince" exec="evince %u" count="1"
                                modified="2026-05-15T14:10:00Z"/>
        </bookmark:applications>
      </metadata>
    </info>
  </bookmark>
</xbel>
```

Dùng `xml.etree.ElementTree` (stdlib) để sinh file này.

---

## 9. Cách chạy & kiểm tra

### Chạy tests hiện tại

```bash
cd d:\Project-EIU\pack-a-mal\service-simulation-module\artifact-extension
python -m pytest tests/ -v
# Expected: 10/10 PASS
```

### Chạy Docker stack

```bash
cd d:\Project-EIU\pack-a-mal\dynamic-analysis

# Lần đầu hoặc sau lỗi network:
docker-compose -f docker-compose.network-sim.yml down --remove-orphans
docker-compose -f docker-compose.network-sim.yml up -d

# Xem logs của init container:
docker logs pack-a-mal-artifact-ext

# Kiểm tra artifacts đã được inject vào volume:
docker run --rm -v pack-a-mal-artifacts:/art alpine ls -la /art/
```

### Chạy Web Demo

```bash
cd d:\Project-EIU\pack-a-mal\web_demo
.\venv\Scripts\python.exe app.py
# Mở http://127.0.0.1:5500
# → Card "UBER-like Artifact Extension" → Inject Docker + Chạy PoC Tests
```

---

## 10. Điểm cần lưu ý khi phát triển tiếp

1. **Epoch timestamps**: Chrome = microseconds since 1601-01-01. Firefox & Linux = microseconds since 1970-01-01 (Unix). Nhầm epoch là lỗi detect thường gặp nhất.

2. **Volume mounting trong Docker**: Artifacts Linux phải được mount vào đúng path mà sample chạy (ví dụ `/root/.bash_history`, `/home/user/.mozilla/`). Cần thêm mount points trong `docker-compose.network-sim.yml`.

3. **Extension không cần network**: Giữ init container **không có `networks:`** — tránh IP conflict với inetsim.

4. **Profile-driven**: Mọi số lượng (bao nhiêu command trong bash history, bao nhiêu Firefox URLs) đều phải đi qua `UserProfile` — không hardcode số.

5. **Freshness test**: Viết thêm test kiểm tra hai lần inject sinh timestamps khác nhau (đã có cho DNS, cần thêm cho Linux artifacts).
