from config.inetsim import generate_config
from api.server import app
import time, socket #import module time để tạm dừng, socket để kiểm tra kết nối mạng

generate_config()

while True:
    try:
        socket.create_connection(("inetsim", 80), timeout=5) #kiểm tra kết nối tới dịch vụ inetsim trên cổng 80
        break
    except OSError: #nếu không kết nối được thì tạm dừng 1 giây và thử lại
        time.sleep(1)

# Start Flask API server
if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000) #chạy server Flask trên tất cả các địa chỉ mạng với cổng 5000