FROM mysql/mysql-server:5.7
RUN yum install openssl -y
RUN mkdir -p /etc/mysql/mysql.conf.d/

COPY etc/toggle_ssl.sh /docker-entrypoint-initdb.d/
COPY etc/test.sql /docker-entrypoint-initdb.d/

COPY ./ssl /ssl
