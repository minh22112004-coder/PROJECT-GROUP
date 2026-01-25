
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    inetsim \
    net-tools \
    curl \
    dnsmasq 
    

COPY entrypoint.sh /entrypoint.sh
#thêm quyền execute cho file entrypoint.sh, chmod (change mod) r,w,x
RUN chmod +x /entrypoint.sh   
#khi khởi động container sẽ chạy entrypoint.sh 
CMD ["/entrypoint.sh"]