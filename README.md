# acmeDeliver

![GitHub](https://img.shields.io/github/license/julydate/acmeDeliver?style=flat-square)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/julydate/acmeDeliver?style=flat-square)
![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/julydate/acmeDeliver?include_prereleases&style=flat-square)

 acme.sh 证书分发服务

将 acme.sh 获取的证书通过 web 服务分发到多台服务器

本分支为服务端源码，客户端源码在 [client](https://github.com/julydate/acmeDeliver/tree/client) 分支下

## Usage

### Server

```bash
$ ./acmeDeliver -h
acmeDeliver version: 1.1
Usage: acmeDeliver [-h] [-p port] [-d dirname] [-k password] [-t time] [-b address] [-tls] [-tlsport port] [-cert filename] [-key filename]

Options:
  -h    
        显示帮助信息
  -p string
        服务端口,默认 9090 (default "9090")
  -d string
        证书文件所在目录,默认当前目录 (default "./")
  -k string
        密码,默认 passwd (default "passwd")
  -t int
        时间戳误差,默认 60 秒 (default 60)
  -b string
        绑定监听地址,默认绑定所有接口
  -tls
        是否监听 TLS,默认关闭
  -tlsport string
        TLS 服务端口,默认 9443 (default "9443")
  -cert string
        TLS 服务证书文件,默认 cert.pem
  -key string
        TLS 服务私钥文件,默认 key.pem (default "key.pem")

$ ./acmeDeliver -p 8080 -d "/tmp/acme" -k "passcode" -t 600 -b 0.0.0.0 -tls -tlsport 8443 -cert server.pem -key server.key
```

### Client

切换到 [client](https://github.com/julydate/acmeDeliver/tree/client) 分支

Download `client.sh` to your machine(下载`client.sh`到你的机器上)

```bash
wget https://raw.githubusercontent.com/julydate/acmeDeliver/client/client.sh
```

```bash
# Get single file `mydomain.net.key` to current work folder
# 单独下载'mydomain.net.key'文件到当前工作目录
./client.sh -d "mydomain.net" -p "passcode" -s "myacmedeliverserver.net:8080" -n "mydomain.net.key"


# Automatically download certs only when server's certs' timestamp updates (Only download and do not deploy)
# 仅在服务端证书的时间戳更新时自动下载证书密钥(仅下载不部署)
./client.sh -d "mydomain.net" -p "passcode" -s "myacmedeliverserver.net:8080" -c "0"


# Automatically download certs only when server's certs' timestamp updates and deploy to apache
# 仅在服务端证书的时间戳更新时自动下载证书密钥并部署到apache
#
# !CAUTION! MUST SET apache_* vars before execute this script!
# !注意! 运行脚本前必须设置`apache_*`相关变量
# Example
# apache_cert_file="/path/to/certfile/in/apache/cert.pem"
# apache_key_file="/path/to/keyfile/in/apache/key.pem"
# apache_fullchain_file="/path/to/fullchain/certfile/apache/fullchain.pem"
#
# To execute commands after updating the certificate, uncomment and configure `apache_reloadcmd` the content yourself 
# 若要更新证书后执行命令，请取消注释并自行配置`apache_reloadcmd`内容
# 
./client.sh -d "mydomain.net" -p "passcode" -s "myacmedeliverserver.net:8080" -c "a"
#
# The configurations of nginx are the same, except for the prefix of the variable
# nginx除了变量的前缀的配置相同

```

## Document

详细教程：[使用 acme.sh 部署通配符证书申请与分发服务](https://www.julydate.com/post/462996681/)

简明教程如下，以 Debian 和当前版本，使用 CloudFlare 为例。

### 服务端

```bash
# 安装环境
apt-get install openssl cron socat curl -y
apt-get update ca-certificates
systemctl enable cron
systemctl start cron

# 创建工作目录
mkdir -p /home/acme

# 安装 acme.sh 脚本
curl https://get.acme.sh | sh
source ~/.bashrc
source ~/.bash_profile
acme.sh  --upgrade  --auto-upgrade --log  "/home/acme/acme.log"

# 定义临时变量
# example.com 修改成你的域名
export DOMAIN="example.com"
# 下面的内容根据所使用的 DNS 服务商更改
export CF_Key="b8e8fff91ff445a1a238fc080797910b"
export CF_Email="admin@example.com"

# 设置 CA
acme.sh --set-default-ca --server letsencrypt

# 签发证书
mkdir -p /home/acme/${DOMAIN}
acme.sh --issue --dns dns_cf -d ${DOMAIN} -d *.${DOMAIN}

# 移动证书
acme.sh --install-cert -d ${DOMAIN} \
--cert-file      /home/acme/${DOMAIN}/cert.pem  \
--key-file       /home/acme/${DOMAIN}/key.pem  \
--fullchain-file /home/acme/${DOMAIN}/fullchain.pem \
--reloadcmd     "echo \$(date -d \"\$current\" +%s) > /home/acme/${DOMAIN}/time.log"

# 下载 acmeDeliver
curl -sLo /home/acme/acmeDeliver https://github.com/julydate/acmeDeliver/releases/download/v1.1/acmeDeliver_1.1_Linux_x86_64
chmod +x /home/acme/acmeDeliver


# 运行 acmeDeliver, -p 指定端口 -k 指定同步密码（请不要用此处的密码）
/home/acme/acmeDeliver -p 9929 -d "/home/acme/" -k 9bff385c71d051c3e81af2bb6950b3e4

# 上一步没有问题则后台运行
nohup /home/acme/acmeDeliver -p 9929 -d "/home/acme/" -k 9bff385c71d051c3e81af2bb6950b3e4 > /home/acme/acmeDeliver.log 2>&1 &

# 防火墙放行指定端口
iptables -I INPUT -m state --state NEW -m tcp -p tcp --dport 9929 -j ACCEPT

# 设置进程守护
cat > /etc/systemd/system/acmeDeliver.service << EOF
[Unit]
Description=acmeDeliver
After=network-online.target
Wants=network-online.target systemd-networkd-wait-online.service
[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
DynamicUser=true
ExecStart=/home/acme/acmeDeliver -p 9929 -d "/home/acme/" -k 9bff385c71d051c3e81af2bb6950b3e4 > /home/acme/acmeDeliver.log 2>&1 &
[Install]
WantedBy=multi-user.target
EOF

# 设置开机启动
systemctl enable --now acmeDeliver
```

### 客户端

```bash
# 安装环境
apt-get install openssl cron curl -y
apt-get update ca-certificates
systemctl enable cron
systemctl start cron

# 下载客户端
curl -sLo /root/acmeDeliverClient.sh https://raw.githubusercontent.com/julydate/acmeDeliver/client/client.sh
chmod +x /root/acmeDeliverClient.sh

# 更改客户端的工作目录
sed -i 's|\/tmp\/acme|\/home\/acme/|g' /root/acmeDeliverClient.sh

# 测试运行客户端
# 其中 -p 指定的密码就是前面你部署服务端的时候设置的密码
# 233.233.233.233:9929 改为你服务器的 IP 和前面设置的服务端口
/root/acmeDeliverClient.sh  -d "example.com" -p "9bff385c71d051c3e81af2bb6950b3e4" -s "http://233.233.233.233:9929" -c "0"

# 设置客户端定时同步
crontab -e
# 最后一行添加以下内容并保存
0 0 * * * /root/acmeDeliverClient.sh  -d "example.com" -p "9bff385c71d051c3e81af2bb6950b3e4" -s "http://233.233.233.233:9929" -c "0" > /dev/null 2>&1 &

```

证书在 `/home/acme` 目录下

## Contributors

[![Moe](https://avatars.githubusercontent.com/u/25688691?v=4&s=48)](https://github.com/MoeMegu)
[@Moe](https://github.com/MoeMegu)

[![Raoby](https://avatars.githubusercontent.com/u/56875134?v=4&s=48)](https://github.com/Raobee)
[@Raoby](https://github.com/Raobee)

## Thanks

[acme.sh](https://github.com/acmesh-official/acme.sh)