# Cấu trúc Folder - Service Simulation Module

```
service-simulation-module/
├── docker-compose.yml              ← khởi chạy toàn bộ stack
├── README.md
├── QUICKSTART.md
├── QUICK_REFERENCE.md
│
├── inetsim/                        ← container giả lập dịch vụ mạng
│   ├── Dockerfile
│   └── entrypoint.sh               ← script khởi động INetSim
│
├── service-simulation/             ← container phân tích HTTP
│   ├── Dockerfile
│   └── app/
│       ├── main.py                 ← entry point ứng dụng
│       ├── analyzer/               ← phân tích request
│       ├── api/                    ← REST API server
│       ├── collector/              ← thu thập log
│       ├── config/                 ← cấu hình INetSim
│       └── handler/                ← xử lý response & executable
│
└── shared/                         ← volume dùng chung 2 container
    ├── config/
    │   └── etc/
    │       └── inetsim/
    │           └── inetsim.conf    ← cấu hình dịch vụ giả lập
    └── logs/
        ├── executables/            ← log file thực thi bắt được
        └── inetsim/                ← log debug/main/service
```

## Luồng hoạt động

- `docker-compose.yml` khởi động 2 container: `inetsim` (giả lập HTTP/DNS/FTP...) + `service-simulation` (proxy phân tích)
- Volume `shared/` kết nối cả hai: config được đọc chung, log được ghi chung
- `inetsim/entrypoint.sh` → khởi INetSim → `service-simulation/app/main.py` → lắng nghe request → ghi log vào `shared/logs/`
