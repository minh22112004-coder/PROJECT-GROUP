#!/bin/bash 

mkdir -p /var/log/inetsim   #tạo thư mục log nếu chưa tồn tại , p là parent ( thư mục cha)
chown -R inetsim:inetsim /var/log/inetsim #thay đổi chủ sở hữu của thư mục log thành user inetsim, r là recursive ( đệ quy)
chmod -R 775 /var/log/inetsim #thay đổi quyền truy cập của thư mục log, 7 cho chủ sở hữu (r,w,x), 7 cho nhóm (r,w,x), 5 cho người khác(r,x)

touch /var/log/inetsim/main.log #tạo file log chính nếu chưa tồn tại
chown inetsim:inetsim /var/log/inetsim/main.log #thay đổi chủ sở hữu của file log thành user inetsim
chmod 664 /var/log/inetsim/main.log #thay đổi quyền truy cập của file log, 6 cho chủ sở hữu (r,w), 6 cho nhóm (r,w), 4 cho người khác (r)

exec /usr/bin/inetsim --data /var/lib/inetsim --conf /etc/inetsim/inetsim.conf --log-dir /var/log/inetsim #chạy inetsim với các tham số chỉ định thư mục dữ liệu, file cấu hình và thư mục log