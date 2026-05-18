# Cách Beaconing Detection Hoạt Động

## 1. Vấn đề cần giải quyết

Malware hiện đại (APT, botnet, spyware) **không tự hành động độc lập** — chúng phải duy trì kết nối với C2 server để nhận lệnh. Kênh giao tiếp này gọi là **beaconing**: malware định kỳ gửi tín hiệu "tôi vẫn đang sống" về C2.

> **Vấn đề:** Traffic này nhỏ, trông giống HTTP thường, và thường được mã hóa bằng HTTPS.  
> **Giải pháp:** Không cần đọc nội dung — chỉ cần phân tích **thời điểm gửi**.

---

## 2. Thu thập dữ liệu đầu vào

Bộ detector chỉ cần **metadata** của mỗi HTTP request:

```
{ src_ip, dst_ip, url, timestamp }
```

Không cần: payload, headers, nội dung mã hóa.

---

## 3. So sánh hành vi người dùng vs malware

```
TRAFFIC NGƯỜI DÙNG THẬT
  t=0s    ──► GET google.com
  t=47s   ──► GET youtube.com
  t=312s  ──► POST api/login
  t=891s  ──► GET news.com
               ↑ không đều, nhiều endpoint khác nhau → NORMAL

TRAFFIC MALWARE BEACONING
  t=0s    ──► POST /jquery-3.3.1.min.js  ──► C2 Server
  t=61s   ──► POST /jquery-3.3.1.min.js  ──►
  t=119s  ──► POST /jquery-3.3.1.min.js  ──►  ← interval ~60s
  t=182s  ──► POST /jquery-3.3.1.min.js  ──►
  t=243s  ──► POST /jquery-3.3.1.min.js  ──►
               ↑ đều đặn, cùng 1 endpoint, lặp nhiều lần → BEACON
```

---

## 4. Thuật toán phát hiện

### Bước 1 — Group traffic theo (src_ip, dst_ip, url)
Gom tất cả request từ cùng một nguồn đến cùng một endpoint.

### Bước 2 — Tính các khoảng thời gian giữa request
```
timestamps = [0, 61, 119, 182, 243]
Δt = [61, 58, 63, 61, 59]  ← delta giữa các lần
```

### Bước 3 — Tính jitter (độ không đều)

$$\text{jitter} = \frac{\sigma(\Delta t)}{\mu(\Delta t)} \times 100\%$$

```
mean(Δt)  = 60.4s
stdev(Δt) = 1.8s
jitter    = 1.8 / 60.4 × 100% = 2.98%
```

### Bước 4 — Ra quyết định

| Điều kiện | Ngưỡng |
|---|---|
| Jitter thấp | < 15% |
| Số lần lặp | ≥ 4 requests |
| Cùng endpoint | một URL duy nhất |

Nếu **cả 3 điều kiện cùng thoả** → **BEACON ALERT**

---

## 5. Sơ đồ tổng quan pipeline

```
Network Traffic (raw)
        │
        ▼
  [Packet Capture]
        │
        ▼ (chỉ lấy metadata)
  { ip, url, timestamp }
        │
        ▼
  [Session Tracker]
  Group by (src_ip, dst_ip, url)
        │
        ▼
  [Beacon Detector]
  Tính Δt → mean, stdev, jitter
        │
     jitter < 15%
     count  ≥ 4
        │
        ▼
  [Alert Engine]
  Ghi alert → c2_alerts.json
        │
        ▼
  [Dashboard / SIEM]
  Hiển thị cảnh báo
```

---

## 6. Ví dụ thực tế — Cobalt Strike

Cobalt Strike là C2 framework phổ biến nhất trong red team và APT thực tế:

```
Cấu hình mặc định:
  Beacon interval: 60 seconds
  Jitter:          0% (không thêm nhiễu)

Traffic quan sát được:
  10.0.0.5 → 185.220.1.47 → POST /jquery-3.3.1.min.js
  Timestamps: 0s, 60s, 120s, 180s, 240s...
  Jitter tính được: ~0.5% → ALERT ngay sau 4 requests (4 phút)
```

---

## 7. Tại sao attacker không thể né được?

| Kỹ thuật attacker dùng | Có né được không? | Lý do |
|---|---|---|
| Mã hóa HTTPS | Không | Detector chỉ dùng timestamp, không đọc payload |
| Giả User-Agent (Chrome/Firefox) | Không | Detector không nhìn headers |
| Đổi URL liên tục (DGA) | Một phần | Nếu vẫn cùng IP đích, vẫn bị detect theo (src_ip, dst_ip) |
| Tăng jitter lớn | Một phần | Nếu jitter > 40% thì C2 mất ổn định, malware hoạt động kém |

> **Ràng buộc cốt lõi:** Malware **bắt buộc** phải beacon đủ đều để C2 phản hồi kịp thời. Đây là trade-off không thể tránh — nếu jitter quá lớn, C2 mất kiểm soát.

---

## 8. Độ chính xác trong thực tế

Các hệ thống enterprise dùng beaconing detection:

| Hệ thống | Phương pháp |
|---|---|
| **Cisco Cognitive Intelligence** | Statistical analysis trên NetFlow metadata |
| **Splunk SIEM** | SPL queries tính stddev của inter-arrival time |
| **Elastic Security** | ML model trên ECS network events |
| **Darktrace** | Unsupervised learning phát hiện periodic anomaly |

Tham khảo:
- [Cisco Cognitive Threat Analytics](https://blogs.cisco.com/security/cognitive-threat-analytics-transparency-in-advanced-threat-research)
- [Elastic Security Labs — Identifying Beaconing Malware](https://www.elastic.co/security-labs/identifying-beaconing-malware-using-elastic)
- [NCC Group — Detecting Beaconing via Mathematical Analysis](https://research.nccgroup.com/2021/06/04/detecting-beaconing-activity-from-malware-using-mathematical-analysis/)

---

## 9. Tóm tắt một câu

> Beaconing detection phát hiện malware **không qua nội dung gói tin**, mà qua **nhịp điệu thời gian** — thứ mà malware không thể che giấu nếu vẫn muốn duy trì kết nối C2.
