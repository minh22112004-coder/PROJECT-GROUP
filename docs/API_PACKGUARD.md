# PackGuard Analysis API 1.0.0 - Tổng hợp

## 🔑 Authentication
- API key auth: `X-API-Key: <API_KEY>`
- Bearer auth: `Authorization: Bearer <API_KEY>`

## ⚙️ Engines
- **dynamic-analysis**: mặc định cho npm, PyPI, VS Code extensions
- **lastpymile**: cho npm và PyPI

## 📤 Submission Modes
1. **Live scan (registry)**
   - `POST /api/v1/analyze/purl/`
   - Phân tích trực tiếp từ registry

2. **Local archive scan (3 bước)**
   - `POST /api/v1/archives/upload-url/` → nhận signed upload URL + `gcs_path`
   - `PUT` file lên signed URL (Google Cloud Storage)
   - `POST /api/v1/analyze/archive/` với `purl` + `gcs_path` (+ tùy chọn GitHub URL)

## 📂 Nhóm API chính

### 1. Analysis
- `POST /api/v1/analyze/purl/` → phân tích gói từ registry
- `POST /api/v1/analyze/archive/` → phân tích gói từ file nén đã upload

### 2. Archives
- `POST /api/v1/archives/upload-url/` → tạo signed URL để upload
- `PUT <signed_url>` → upload file nén
- Dùng `gcs_path` để tham chiếu khi gọi analyze/archive

### 3. Tasks
- `GET /api/v1/tasks/{task_id}/` → xem trạng thái một task
- `GET /api/v1/tasks/` → liệt kê tất cả task

### 4. Reports
- `GET /api/v1/reports/{report_id}/` → lấy báo cáo chi tiết
- `GET /api/v1/reports/` → liệt kê các báo cáo

## 📌 Ví dụ request

### Live scan với purl
```bash
curl -X POST "https://api.packguard.io/api/v1/analyze/purl/" \
  -H "Authorization: Bearer <API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"purl":"pkg:npm/express@4.17.1"}'

# 1. Lấy upload URL
curl -X POST "https://api.packguard.io/api/v1/archives/upload-url/" \
  -H "Authorization: Bearer <API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"filename":"express-4.17.1.tgz"}'

# 2. Upload file lên signed URL (PUT)
curl -X PUT "<SIGNED_URL>" \
  -H "Content-Type: application/gzip" \
  --data-binary "@express-4.17.1.tgz"

# 3. Gọi analyze/archive
curl -X POST "https://api.packguard.io/api/v1/analyze/archive/" \
  -H "Authorization: Bearer <API_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"purl":"pkg:npm/express@4.17.1","gcs_path":"<GCS_PATH>"}'
```


# 📦 PackGuard API v1

## 1. Analyze Archive

### **POST** `/api/v1/analyze/archive/`

**Request Body**

```json
{
  "purl": "pkg:npm/axios@1.14.1",
  "gcs_path": "uploads/uuid/axios-1.14.1.tgz"
}
```

### ✅ Responses

#### **201 Created**

```json
{
  "success": true,
  "request_id": "string",
  "data": {
    "task_id": 42,
    "status_url": "string",
    "message": "string"
  }
}
```

#### **200 Active**

```json
{
  "data": {
    "status": "processing",
    "message": "Analysis already processing",
    "task_id": 456,
    "status_url": "https://packguard.dev/api/v1/task/456"
  },
  "success": true,
  "request_id": "d96c0c9a-6e3c-44cb-bfb1-1b1c5b3c50b0"
}
```

#### ❌ Errors

* `400 Bad Request`
* `401 Unauthorized`
* `405 Method Not Allowed`
* `500 Internal Server Error`

```json
{
  "success": false
}
```

---

## 2. Archives

### **POST** `/api/v1/archives/upload-url/`

**Request Body**

```json
{
  "filename": "axios-1.14.1.tgz"
}
```

### ✅ Response

#### **200 OK**

```json
{
  "data": {
    "url": "https://storage.googleapis.com/...",
    "message": "Upload URL generated successfully",
    "gcs_path": "uploads/uuid/axios-1.14.1.tgz"
  },
  "success": true,
  "request_id": "5a7ffdf8-44a9-4c7f-84b5-68e5a38cb7fb"
}
```

#### ❌ Errors

* `400 / 401 / 405 / 500`

```json
{
  "success": false
}
```

---

## 3. Tasks

### **GET** `/api/v1/task/{task_id}/`

### ✅ Response

#### **200 Processing**

```json
{
  "data": {
    "purl": "pkg:pypi/requests@2.25.1",
    "status": "processing",
    "task_id": 123,
    "created_at": "2026-04-09T10:20:30.123456+00:00",
    "engine_type": "dynamic-analysis"
  },
  "success": true,
  "request_id": "0b7d1d7f-6b9c-4e14-9a67-f5f0d2d6b9bf"
}
```

#### ❌ Errors

* `404 Not Found`
* `500 Internal Server Error`

```json
{
  "success": false
}
```

---

## 4. Reports

### **GET** `/api/v1/reports/`

**Query Parameters**

| Name      | Type   | Description       |
| --------- | ------ | ----------------- |
| page      | int    | Trang hiện tại    |
| page_size | int    | Số item mỗi trang |
| status    | string | Trạng thái task   |

### ✅ Response

#### **200 OK**

```json
{
  "success": true,
  "request_id": "string",
  "data": {
    "items": [
      {
        "task_id": 42,
        "purl": "string",
        "engine_type": "dynamic-analysis",
        "status": "queued",
        "created_at": "string",
        "status_url": "string",
        "error_message": "string"
      }
    ],
    "page": 42,
    "page_size": 42,
    "total": 42
  }
}
```

#### ❌ Errors

* `400 / 401 / 405 / 500`

```json
{
  "success": false
}
```
