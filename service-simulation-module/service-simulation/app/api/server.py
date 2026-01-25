from flask import Flask 

app = Flask(__name__) #khởi tạo ứng dụng Flask

@app.route("/status") #định nghĩa route /status
def status():
    return {"service": "simulation", "status": "running"} #trả về trạng thái dịch vụ dưới dạng JSON