# Demo Script — UBER-like Artifact Extension

> Đối tượng: Đồng nghiệp kỹ thuật  
> Thời gian ước tính: 15–20 phút  
> Yêu cầu: Docker Desktop đang chạy, terminal tại `d:\Project-EIU\pack-a-mal`

---

## TRƯỚC KHI BẮT ĐẦU — Checklist chuẩn bị

```powershell
# 1. Khởi động Docker stack
cd d:\Project-EIU\pack-a-mal\dynamic-analysis
docker-compose -f docker-compose.network-sim.yml up -d

# 2. Khởi động web demo (terminal riêng)
cd d:\Project-EIU\pack-a-mal\web_demo
.\venv\Scripts\python.exe app.py

# 3. Mở browser: http://127.0.0.1:5500
```

Kiểm tra nhanh:
- [ ] `docker ps` thấy `pack-a-mal-inetsim` (Up) và `pack-a-mal-sim-api` (Up)
- [ ] `pack-a-mal-artifact-ext` ở trạng thái `Exited (0)` — bình thường
- [ ] Browser mở http://127.0.0.1:5500 hiện dashboard

---

## PHẦN 1 — Vấn đề & Bối cảnh (2 phút)

**Nói:**
> "Malware thông minh không chạy ngay — nó kiểm tra môi trường trước.
> Nếu thấy dấu hiệu sandbox, nó tự huỷ. Kết quả là sandbox thu được không có gì."

**Cho xem:**  
Mở `tests/test_poc_detector.py`, chỉ vào class `SandboxDetector` và giải thích từng check:

```python
def check_dns_cache_missing(self)    # DNS cache không tồn tại?
def check_dns_cache_sparse(self)     # Ít hơn 10 entries?
def check_dns_ttl_uniform(self)      # Tất cả TTL đều nhau? (bulk inject)
def check_browser_history_missing(self)  # Không có browser history?
def check_browser_history_sparse(self)   # Ít hơn 20 URLs?
def check_transition_uniform(self)   # Tất cả visit cùng 1 loại transition?
```

**Nói:**
> "Đây là 6 kiểm tra mà malware (hoặc sandbox fingerprinter) thực hiện.
> Sandbox sạch sẽ fail hết 6 cái này."

---

## PHẦN 2 — Kiến trúc Plugin (3 phút)

**Cho xem cấu trúc thư mục:**

```
artifact_extension/
├── base.py      ← Abstract base: 2 method cần implement
├── manager.py   ← Điều phối: broadcast events, gọi inject()
├── profile.py   ← Cấu hình thống kê (bao nhiêu entries, phân phối gì)
└── extensions/
    ├── dns_extension.py   ← Sinh DNS cache
    └── http_extension.py  ← Sinh Browser History + Zone.Identifier
```

**Mở `base.py`, giải thích plugin contract:**

```python
class ArtifactExtension(ABC):
    def on_service_event(self, event: dict) -> None:
        # Nhận event từ service simulation
        # Extension tự quyết định xử lý hay bỏ qua

    def inject(self, artifacts_path: str) -> None:
        # Ghi artifact vào Docker volume
        # Mỗi extension tạo subdirectory riêng
```

**Nói:**
> "Muốn thêm artifact mới — ví dụ bash history cho Linux — chỉ cần tạo class
> kế thừa ArtifactExtension, implement 2 method này, rồi register với manager.
> Không cần chạm vào code hiện có."

**Mở `manager.py`, chỉ vào `dispatch()` và `inject_all()`:**

```python
manager.register(DNSArtifactExtension(profile))
manager.register(HTTPArtifactExtension(profile))
manager.dispatch(event)   # broadcast đến tất cả extensions
manager.inject_all()      # tạo thư mục, gọi inject() từng cái
```

---

## PHẦN 3 — Demo PoC Test Cases (3 phút)

**Trên web demo, cuộn đến card "UBER-like Artifact Extension"**

**Click "Chạy PoC Tests"** và giải thích trong khi chạy:

> "Chúng ta có 3 test case kiểm tra toàn bộ flow:"

Chờ kết quả hiện ra, giải thích từng case:

| Case | Setup | Kết quả mong đợi |
|---|---|---|
| **Case A** | Sandbox hoàn toàn sạch | Detector nói "ĐÂY LÀ SANDBOX" ✅ |
| **Case B** | Có service events nhưng chưa inject | Detector vẫn detect được ✅ |
| **Case C** | Đã inject đầy đủ artifacts | Detector bị qua mặt — "Không phải sandbox" ✅ |

**Nói:**
> "Case C là điểm mấu chốt — sau khi inject xong,
> sandbox trông như máy thật và malware sẽ chạy payload thật."

---

## PHẦN 4 — Live Injection Demo (5 phút)

### 4a. Cho xem artifacts hiện tại trong Docker volume

Mở terminal, chạy:

```powershell
docker run --rm -v pack-a-mal-artifacts:/art alpine sh -c "find /art -type f"
```

Kết quả expected:
```
/art/dns/cache.txt
/art/browser/History
/art/browser/downloads/setup.exe.Zone.Identifier
/art/browser/downloads/ChromeSetup.exe.Zone.Identifier
/art/browser/downloads/DiscordSetup.exe.Zone.Identifier
/art/browser/downloads/VSCodeSetup-x64.exe.Zone.Identifier
/art/browser/downloads/requests-2.31.0.tar.gz.Zone.Identifier
```

### 4b. Inject Docker button trên web UI

**Click "Inject Docker"** và giải thích trong khi stream output chạy:

> "Container `artifact-extension` được spin up, ghi artifacts vào volume,
> rồi tự exit. Service simulation container mount cùng volume đó
> nên thấy đầy đủ artifacts."

Chờ output xuất hiện trong terminal panel.

### 4c. Click "Summary"

Sau khi inject xong, click **"Summary"** để xem stats:

| Metric | Giá trị expected |
|---|---|
| DNS Entries | ~157 |
| History URLs | ~25–30 |
| Zone.Identifier files | 5 |

### 4d. Chỉ vào DNS cache file (giải thích chống fingerprinting)

Mở terminal:

```powershell
docker run --rm -v pack-a-mal-artifacts:/art alpine grep "Time To Live" /art/dns/cache.txt | Select-Object -First 10
```

**Nói:**
> "Chú ý TTL của từng entry khác nhau — 847, 2341, 156, 3012...
> Nếu tất cả TTL đều nhau, đó là dấu hiệu inject hàng loạt.
> Malware check đúng điều này."

---

## PHẦN 5 — Chi tiết Kỹ thuật (nếu được hỏi)

### Chrome Epoch Bug (hay gặp nhất)

```python
# SAI — dùng Unix epoch
timestamp = int(time.time() * 1_000_000)

# ĐÚNG — Chrome epoch bắt đầu từ 1601-01-01
_CHROME_EPOCH_OFFSET_US = 11_644_473_600 * 1_000_000
timestamp = int(time.time() * 1_000_000) + _CHROME_EPOCH_OFFSET_US
```

> "Đây là lỗi phổ biến nhất trong fake browser history.
> Chrome lưu timestamp tính từ năm 1601, không phải 1970.
> Nếu dùng Unix epoch, timestamp nhỏ hơn thực tế 116 năm — detect ngay."

### Zone.Identifier (Windows NTFS ADS)

```ini
[ZoneTransfer]
ZoneId=3          # 3 = Internet zone (file tải từ internet)
ReferrerUrl=https://www.google.com/
HostUrl=https://cdn.discordapp.com/apps/DiscordSetup.exe
```

> "Trên Windows, mọi file tải từ internet đều có stream NTFS này.
> Nếu sandbox không có — đó là dấu hiệu rõ ràng."

### User Profile (tránh hardcode)

```python
UserProfile(
    dns=DNSProfile(entries=150, ttl_jitter_seconds=3600),
    browser=BrowserProfile(history_entries=400, days_of_history=14)
)
```

> "Tất cả số lượng đều cấu hình qua profile JSON — không hardcode.
> Mount `/config/artifact_profile.json` vào container để thay đổi."

---

## PHẦN 6 — Init Container Pattern (Docker) (2 phút)

**Mở `docker-compose.network-sim.yml`, chỉ vào phần `artifact-extension`:**

```yaml
artifact-extension:
  restart: "no"           # chạy một lần rồi thoát — KHÔNG restart

service-simulation:
  depends_on:
    artifact-extension:
      condition: service_completed_successfully  # chờ exit 0 mới start
```

**Vẽ sơ đồ trên whiteboard hoặc giải thích:**

```
[inetsim]        → start → healthy
[artifact-ext]   → start → generate artifacts → EXIT(0)
[service-sim]    → start (sau khi 2 cái trên xong) → running
                          ↑
                  thấy artifacts trong volume /artifacts
```

**Nói:**
> "artifact-extension không phải service thông thường — nó là init container.
> Chạy một lần, ghi xong file, thoát. Giống setup script chạy trước khi app start.
> Và quan trọng: không có network access — không cần, và bỏ network còn tránh IP conflict."

---

## Câu hỏi thường gặp

**Q: Tại sao cần Zone.Identifier nếu sandbox chạy trên Linux?**  
A: Một số malware Windows được phân tích trong Wine hoặc Windows guest VM — chúng vẫn check NTFS ADS. Ngoài ra, đây là template — Linux artifacts (bash history, Firefox places.sqlite, recently-used.xbel) sẽ là phần mở rộng tiếp theo.

**Q: Artifacts có thực sự bị malware detect không?**  
A: PoC Detector trong test suite mô phỏng các check thật. Case C chứng minh artifact injection bypass được. Với malware thực tế, hiệu quả phụ thuộc vào mức độ tinh vi của fingerprinting routine.

**Q: Làm sao thêm artifact type mới?**  
A: Tạo class kế thừa `ArtifactExtension`, implement `on_service_event()` và `inject()`, rồi gọi `manager.register(NewExtension(profile))` trong `entrypoint.py`.

**Q: Profile JSON để ở đâu?**  
A: Mount vào `/config/artifact_profile.json` trong Docker. Xem `service-simulation-module/shared/config/artifact_profile.json`.
