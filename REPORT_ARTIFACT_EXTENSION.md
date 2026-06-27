**BÁO CÁO TUẦN — TÍCH HỢP FAKE ARTIFACTS CHO MÔI TRƯỜNG LINUX**

**1. Mục tiêu thực hiện**
- Trong tuần này, nhóm tập trung vào việc cài đặt, tích hợp và kiểm thử cơ chế tạo fake artifacts cho môi trường Linux.
- Mục tiêu chính là xây dựng được một luồng chạy ổn định trong repo hiện tại, tạo ra các artifact giả phục vụ mô phỏng môi trường hệ thống, đồng thời đảm bảo các file sinh ra từ quá trình demo không bị đưa vào quản lý mã nguồn.

**2. Phạm vi đã thực hiện**
- Đã sử dụng module sẵn có trong repo là [service-simulation-module/artifact-extension](service-simulation-module/artifact-extension) làm nền tảng triển khai.
- Đã kiểm tra cách hoạt động của các thành phần chính: `entrypoint.py`, `ArtifactManager` và các extension trong [artifact_extension/extensions/](artifact_extension/extensions/).
- Đã chạy kiểm thử để xác nhận các artifact được sinh ra đúng định dạng và đúng hành vi mong đợi.
- Đã cập nhật cấu trúc repo để loại bỏ các file demo khỏi git và tránh phát sinh rác trong quá trình làm việc.

**3. Cách hoạt động của hệ thống**
- `entrypoint.py` là điểm khởi chạy của init-container. Tệp này tạo luồng thực thi ban đầu, kích hoạt quá trình sinh dữ liệu giả và gọi phần điều phối chính.
- `ArtifactManager` trong [artifact_extension/manager.py](artifact_extension/manager.py) đóng vai trò điều phối toàn bộ quá trình. Thành phần này tạo thư mục đầu ra, lần lượt gọi từng extension và tách lỗi theo từng phần để một extension lỗi không làm hỏng toàn bộ quá trình.
- Mỗi extension trong [artifact_extension/extensions/](artifact_extension/extensions/) phụ trách một loại artifact riêng. Ví dụ:
  - `dns_extension.py` tạo file DNS cache giả theo kiểu đầu ra gần giống `ipconfig /displaydns`.
  - `http_extension.py` tạo lịch sử trình duyệt SQLite và các file phụ trợ như `.Zone.Identifier`.
- Luồng này phù hợp với mô hình init-container: artifact được tạo trước khi container chính chạy, giúp mô phỏng môi trường có dấu vết hệ thống giống thật hơn.

**4. Cấu trúc những thay đổi đã thêm hoặc chỉnh sửa**
- **Tầng khởi động:** `entrypoint.py` để kích hoạt chuỗi sinh artifact.
- **Tầng điều phối:** `ArtifactManager` để gom và gọi các extension theo thứ tự.
- **Tầng chức năng:** các file extension trong [artifact_extension/extensions/](artifact_extension/extensions/) để sinh từng loại artifact cụ thể.
- **Tầng kiểm thử:** bộ test PoC trong `tests/` để xác minh kết quả sinh artifact, tính hợp lệ của dữ liệu và độ ổn định của hành vi.
- **Tầng cấu hình repo:** đã cập nhật [PROJECT-GROUP/.gitignore](PROJECT-GROUP/.gitignore) để bỏ qua output demo và cache của pytest.
- **Tầng dữ liệu phát sinh:** đã untrack thư mục [service-simulation-module/artifact-extension/demo_artifacts/](service-simulation-module/artifact-extension/demo_artifacts/) khỏi git để tránh ghi nhận dữ liệu sinh tự động.

**5. Kết quả kiểm thử**
- Môi trường chạy test sử dụng Python 3.14 và `pytest>=7.4`.
- Lệnh kiểm thử đã dùng:
```bash
python -m pytest tests -v
```
- Kết quả kiểm thử cho thấy các kịch bản PoC đều chạy thành công, tổng cộng `10 passed`.
- Các nội dung được kiểm tra gồm: artifact được sinh ra đúng vị trí, file lịch sử trình duyệt có định dạng hợp lệ, DNS cache có nội dung phù hợp và file `.Zone.Identifier` được tạo đúng hành vi mong đợi.

**6. Hướng dẫn chạy sample**
- Hướng dẫn dưới đây ưu tiên chạy bằng Docker để đúng với mô hình init-container của module.
- Tại thư mục gốc `PROJECT-GROUP`, chạy sample bằng Docker Compose:
```powershell
Set-Location 'c:\Users\dangt\OneDrive\Desktop\Do an 22\PROJECT-GROUP'
docker-compose -f docker-compose.network-sim.yml run --rm artifact-extension
```
- Lệnh trên sẽ khởi chạy container `artifact-extension`, sinh các fake artifacts vào volume dùng chung của hệ thống mô phỏng, rồi tự dừng khi hoàn tất.
- Nếu muốn kiểm thử lại module theo dạng test bên trong môi trường Docker, có thể chạy thêm:
```powershell
docker-compose -f docker-compose.network-sim.yml run --rm artifact-extension python -m pytest tests -v
```
- Khi chạy xong, các file demo sẽ xuất hiện trong thư mục output của module, gồm DNS cache, lịch sử trình duyệt và các file phụ trợ liên quan.
- Trường hợp cần debug hoặc chạy local ngoài Docker, có thể kích hoạt virtual environment và chạy pytest trực tiếp, nhưng đây chỉ là phương án phụ.

**7. Nhận xét về hướng tích hợp**
- Trong quá trình tìm hiểu, nhóm có đối chiếu thêm một số repo bên ngoài như `fake-sandbox`, `pafish` và `al-khaser`.
- Tuy nhiên, theo đánh giá hiện tại, việc tích hợp trực tiếp ba repo đó là **có thể về mặt kỹ thuật**, nhưng **lệch hướng so với mục tiêu** của đề tài.
- Lý do là các repo này thiên về **phát hiện sandbox / anti-analysis**, trong khi nhu cầu hiện tại là **tạo fake artifacts cho môi trường Linux**.
- Vì vậy, lựa chọn hợp lý hơn là giữ hướng triển khai trong repo hiện tại, tận dụng `artifact-extension` làm lõi sinh dữ liệu giả. Cách làm này vừa bám sát mục tiêu đồ án, vừa dễ kiểm soát, dễ kiểm thử và dễ báo cáo kết quả.

**8. Đánh giá chung**
- Về mặt kỹ thuật, bài toán đã được triển khai theo hướng khả thi trong repo hiện tại.
- Về mặt định hướng, giải pháp hiện tại phù hợp hơn so với việc cố tích hợp nguyên trạng các repo tham khảo bên ngoài.
- Hệ thống đã có đủ các thành phần cơ bản: luồng khởi tạo, điều phối, sinh artifact, kiểm thử và cấu hình loại trừ file phát sinh.

**9. Kết luận**
- Nhóm đã hoàn thành phần chính của yêu cầu liên quan đến việc tạo fake artifacts cho Linux trong repo hiện tại.
- Các thay đổi đã được kiểm thử và ghi nhận rõ trong code cũng như trong cấu trúc repository.
- Hướng tích hợp thêm các repo tham khảo bên ngoài là khả thi nhưng không cần thiết cho mục tiêu đồ án hiện tại vì sẽ làm lệch trọng tâm triển khai.

**Tài liệu liên quan**
- [PROJECT-GROUP/.gitignore](PROJECT-GROUP/.gitignore)
- [PROJECT-GROUP/REPORT_ARTIFACT_EXTENSION.md](PROJECT-GROUP/REPORT_ARTIFACT_EXTENSION.md)
- [service-simulation-module/artifact-extension/entrypoint.py](service-simulation-module/artifact-extension/entrypoint.py)
- [artifact_extension/manager.py](artifact_extension/manager.py)
