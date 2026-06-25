# Báo Cáo Tích Hợp Fake Artifacts Cho Môi Trường Linux

## 1. Mục tiêu

Mục tiêu của phần này là xây dựng một cơ chế tạo fake artifacts nội bộ để làm môi trường phân tích trông giống máy thật hơn, đặc biệt cho Linux. Hệ thống cần tạo được các dấu vết hợp lý như hostname, machine-id, os-release, DMI, lịch sử shell, log hệ thống và các artifact liên quan đến trình duyệt/danh tính hệ thống.

Thay vì phụ thuộc hoàn toàn vào các công cụ ngoài, hướng triển khai được chọn là tạo extension nội bộ trong repository, sau đó dùng test tự động để xác nhận artifact sinh ra là hợp lý, đủ đa dạng và có tính "freshness" giữa các lần chạy.

## 2. Hướng triển khai đã chọn

Sau khi cân nhắc các công cụ tham chiếu bên ngoài như `fake-sandbox`, `pafish`, và `al-khaser`, hướng phù hợp nhất là:

- Dùng các công cụ đó như nguồn tham chiếu để xác định loại dấu vết mà sandbox/malware thường kiểm tra.
- Xây dựng một Linux artifact extension nội bộ trong `artifact-extension/`.
- Giữ toàn bộ luồng test và inject artifact trong repo để dễ bảo trì, dễ chạy trong Docker, và không phụ thuộc vào môi trường ngoài.

Lý do chọn hướng này:

- Ổn định hơn so với việc phụ thuộc vào tool ngoài.
- Dễ kiểm thử tự động.
- Dễ tích hợp vào pipeline hiện có.
- Phù hợp với mục tiêu của đồ án là mô phỏng môi trường phân tích, không phải đóng gói nguyên bản các tool anti-analysis bên ngoài.

## 3. Các bước đã thực hiện

### Bước 1: Khảo sát kiến trúc hiện có

Đầu tiên, kiểm tra cấu trúc của module `artifact-extension` để xác định điểm tích hợp phù hợp. Kết quả cho thấy hệ thống đã có sẵn:

- `ArtifactManager` để điều phối các extension.
- `DNSArtifactExtension` để tạo cache DNS giả.
- `HTTPArtifactExtension` để tạo browser history và file `Zone.Identifier`.
- `entrypoint.py` để inject artifacts trước khi sample chạy.
- `tests/test_poc_detector.py` để kiểm tra hành vi của artifact injection.

### Bước 2: Thiết kế extension Linux

Đã thêm một extension mới tên `LinuxArtifactExtension` trong:

- `artifact-extension/artifact_extension/extensions/linux_extension.py`

Extension này sinh ra các file và dấu vết Linux phổ biến, gồm:

- `linux/etc/hostname`
- `linux/etc/machine-id`
- `linux/var/lib/dbus/machine-id`
- `linux/etc/os-release`
- `linux/etc/issue`
- `linux/etc/hosts`
- `linux/sys/class/dmi/id/sys_vendor`
- `linux/sys/class/dmi/id/product_name`
- `linux/sys/class/dmi/id/bios_vendor`
- `linux/sys/class/dmi/id/bios_version`
- `linux/proc/cpuinfo`
- `linux/home/<user>/.bash_history`
- `linux/home/<user>/.config/gtk-3.0/settings.ini`
- `linux/home/<user>/.local/share/recently-used.xbel`
- `linux/var/log/syslog`

Các dữ liệu này được tạo theo hướng thống kê và plausibly-real, không cố mô phỏng một máy thật cụ thể.

### Bước 3: Gắn extension Linux vào luồng inject

Đã cập nhật `artifact-extension/entrypoint.py` để register thêm `LinuxArtifactExtension(profile)` bên cạnh DNS và HTTP extension.

Kết quả là khi chạy entrypoint, hệ thống sẽ inject đồng thời:

- DNS artifacts
- HTTP/browser artifacts
- Linux host artifacts

### Bước 4: Xuất extension trong package

Đã cập nhật `artifact_extension/extensions/__init__.py` để export `LinuxArtifactExtension`, giúp việc import và sử dụng nhất quán hơn trong toàn bộ module.

### Bước 5: Bổ sung test xác nhận Linux artifacts

Đã mở rộng `artifact-extension/tests/test_poc_detector.py` để kiểm tra thêm:

- Linux artifact được tạo đúng cấu trúc.
- Machine ID thay đổi giữa hai lần chạy độc lập.

Cụ thể, test mới xác nhận:

- Tồn tại `linux/etc/hostname`
- Tồn tại `linux/etc/machine-id`
- Tồn tại `linux/etc/os-release`
- Tồn tại thư mục `linux/home`
- `machine-id` khác nhau giữa hai run riêng biệt, thể hiện tính freshness

### Bước 6: Cập nhật tài liệu

Đã cập nhật:

- `README.md`
- `QUICKSTART.md`

Mục đích là để mô tả đúng kiến trúc hiện tại, nhấn mạnh rằng phần fake artifacts Linux được tạo nội bộ trong `artifact-extension` và có thể test bằng pytest.

## 4. Kết quả kiểm thử đã xác nhận

Sau khi hoàn thiện thay đổi, đã chạy test trực tiếp trong thư mục `artifact-extension`:

```powershell
Set-Location 'C:\Users\dangt\OneDrive\Desktop\Do an 22\PROJECT-GROUP\service-simulation-module\artifact-extension'; python -m pytest tests -v
```

Kết quả:

- 12/12 test passed
- Bao gồm các test cũ cho DNS, HTTP, Zone.Identifier, Chrome epoch
- Bao gồm 2 test mới cho Linux artifacts

Điểm đáng chú ý:

- Test cho Linux artifacts xác nhận file được sinh đúng.
- Test freshness xác nhận `machine-id` thay đổi giữa các lần chạy.
- Điều này giúp giảm nguy cơ tạo ra artifact quá đồng nhất, vốn là dấu hiệu dễ bị sandbox phát hiện.

## 5. Hướng dẫn test chi tiết

### 5.1. Test theo cách khuyến nghị

Đây là cách test đã được xác nhận hoạt động ổn định:

```powershell
cd "C:\Users\dangt\OneDrive\Desktop\Do an 22\PROJECT-GROUP\service-simulation-module\artifact-extension"
python -m pytest tests -v
```

Nếu toàn bộ hệ thống hoạt động đúng, kết quả mong đợi là:

- `12 passed`
- Không có lỗi import
- Không có lỗi tạo file hoặc lỗi sqlite

Quy trình kiểm tra nên làm theo đúng thứ tự sau:

1. Mở terminal PowerShell.
2. Chuyển vào thư mục `artifact-extension`.
3. Chạy `python -m pytest tests -v`.
4. Xác nhận tất cả test đều `PASSED`.
5. Nếu cần, mở file test hoặc thư mục tạm để kiểm tra artifact sinh ra.

### 5.2. Test riêng phần Linux artifacts

Nếu chỉ muốn kiểm tra phần Linux fake artifacts, có thể chạy toàn bộ suite vì test Linux đã được tích hợp trong cùng bộ test:

```powershell
cd "C:\Users\dangt\OneDrive\Desktop\Do an 22\PROJECT-GROUP\service-simulation-module\artifact-extension"
python -m pytest tests -v -k linux
```

Lệnh này sẽ chạy các test liên quan đến Linux artifacts.

Khi test Linux pass, cần thấy các kiểm tra sau cũng pass:

- Tạo được `linux/etc/hostname`
- Tạo được `linux/etc/machine-id`
- Tạo được `linux/etc/os-release`
- Tạo được `linux/sys/class/dmi/id/product_name`
- `machine-id` thay đổi giữa hai lần chạy độc lập

### 5.3. Test theo từng nhóm chức năng

Có thể lọc theo tên test để kiểm tra từng phần:

- DNS:
```powershell
python -m pytest tests -v -k dns
```

- HTTP / browser history:
```powershell
python -m pytest tests -v -k browser
```

- Zone.Identifier:
```powershell
python -m pytest tests -v -k zone
```

- Linux artifacts:
```powershell
python -m pytest tests -v -k linux
```

Nếu muốn kiểm tra toàn bộ suite nhưng chỉ xem nhanh phần nào, có thể dùng:

```powershell
python -m pytest tests -v -k "dns or browser or zone or linux"
```

Lệnh này hữu ích khi bạn muốn trình bày từng nhóm artifact trong buổi báo cáo hoặc demo.

### 5.4. Test thủ công bằng cách tạo artifact trực tiếp

Nếu muốn kiểm tra artifact sinh ra ở mức file hệ thống, có thể dùng `pytest` chạy xong rồi mở thư mục tạm do test tạo, hoặc chạy extension theo luồng entrypoint trong môi trường phù hợp.

Sau khi inject, cần kiểm tra các đường dẫn sau:

- `artifacts/dns/cache.txt`
- `artifacts/browser/History`
- `artifacts/browser/downloads/*.Zone.Identifier`
- `artifacts/linux/etc/hostname`
- `artifacts/linux/etc/machine-id`
- `artifacts/linux/etc/os-release`
- `artifacts/linux/sys/class/dmi/id/*`
- `artifacts/linux/home/<user>/.bash_history`

Nội dung cần quan sát khi kiểm tra thủ công:

- `hostname` phải là một tên máy hợp lý, không phải giá trị rỗng.
- `machine-id` phải là chuỗi hex dài và khác giữa hai lần inject.
- `os-release` phải có các trường như `PRETTY_NAME`, `ID`, `VERSION_ID`.
- `cpuinfo` phải trông giống máy thật, không được quá trống hoặc quá ngắn.
- `.bash_history` nên chứa các lệnh quản trị hoặc lệnh người dùng phổ biến.
- `syslog` nên có ít nhất vài dòng log hệ thống thông thường.

### 5.5. Nếu test fail thì kiểm tra gì trước

- Kiểm tra đường dẫn đang đứng có đúng là `artifact-extension` hay không.
- Kiểm tra Python đang dùng có cài `pytest` hay chưa.
- Kiểm tra package import có bị lỗi do chạy sai working directory.
- Kiểm tra xem file mới `linux_extension.py` đã được import trong `entrypoint.py` và `extensions/__init__.py` chưa.

### 5.6. Gợi ý khi trình bày trong demo

Nếu cần trình bày trực tiếp trên lớp, nên demo theo ba bước ngắn:

1. Chạy test bằng `python -m pytest tests -v`.
2. Chỉ ra dòng `12 passed` để chứng minh hệ thống đã ổn định.
3. Mở một số file artifact như `linux/etc/os-release` hoặc `browser/History` để cho thấy dữ liệu đã được sinh ra thực tế.

Thông điệp chính khi demo là: hệ thống không chỉ pass test, mà còn sinh ra các dấu vết đủ thuyết phục để mô phỏng môi trường Linux có hoạt động bình thường.

## 6. Kết luận

Đồ án hiện đã có một cơ chế fake artifacts nội bộ cho Linux, được kiểm thử tự động và có tài liệu hướng dẫn chạy rõ ràng. Hướng triển khai này là lựa chọn hợp lý nhất vì:

- Dễ kiểm soát
- Dễ test
- Dễ trình bày trong báo cáo
- Không phụ thuộc vào việc các tool ngoài có chạy được trên Linux hay không

Nếu cần trình bày ngắn gọn trong báo cáo đồ án, có thể tóm tắt như sau:

> Nhóm đã xây dựng một extension nội bộ để sinh fake artifacts cho môi trường Linux, bao gồm hostname, machine-id, os-release, DMI, lịch sử shell và log hệ thống. Hệ thống được kiểm thử bằng pytest và đã xác nhận tạo artifact ổn định, có tính đa dạng giữa các lần chạy, từ đó giúp môi trường phân tích trông giống máy thật hơn và giảm dấu hiệu sandbox.
