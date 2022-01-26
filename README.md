# acmeDeliver

![GitHub](https://img.shields.io/github/license/julydate/acmeDeliver?style=flat-square)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/julydate/acmeDeliver?style=flat-square)
![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/julydate/acmeDeliver?include_prereleases&style=flat-square)

 acme.sh 证书分发服务

将 acme.sh 获取的证书通过 http 服务分发到多台服务器

## Server Usage

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
```

## Server Example

```bash
./acmeDeliver -p 8080 -d "/tmp/acme" -k "passcode" -t 600 -b 0.0.0.0 -tls -tlsport 8443 -cert server.pem -key server.key
```

## Client Usage

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

## Client Example

### Download `client.sh` to your machine(下载`client.sh`到你的机器上)

```bash
wget https://raw.githubusercontent.com/julydate/acmeDeliver/client/client.sh
```

## Document

待更新，配套 bash 客户端开发中

## Contributors

[![Moe](https://avatars.githubusercontent.com/u/25688691?v=4&s=48)](https://github.com/MoeMegu)
[@Moe](https://github.com/MoeMegu)

[![Raoby](https://avatars.githubusercontent.com/u/56875134?v=4&s=48)](https://github.com/Raobee)
[@Raoby](https://github.com/Raobee)
