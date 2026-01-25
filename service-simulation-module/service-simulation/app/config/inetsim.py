import os #import module os để thao tác với hệ thống file

def generate_config():
    config = """service_bind_address 0.0.0.0
dns_default_ip 127.0.0.1          

start_service http
start_service dns

http_default_fakefile   sample.html text/html

dns_bind_port 53                 
"""

    config_dir = "/etc/inetsim"
    config_file_path = os.path.join(config_dir, "inetsim.conf") #đường dẫn tới file cấu hình inetsim.conf trong docker
    
    
    if os.path.exists(config_file_path): #nếu file cấu hình đã tồn tại thì xóa đi để tạo mới
        os.remove(config_file_path)
    
    os.makedirs(config_dir, exist_ok=True) #tạo thư mục cấu hình nếu chưa tồn tại, exist_ok là không báo lỗi nếu thư mục đã tồn tại
    with open(config_file_path, "w") as config_file: #mở file cấu hình để ghi
        config_file.write(config)



