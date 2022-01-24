# acmeDeliver

![GitHub](https://img.shields.io/github/license/julydate/acmeDeliver?style=flat-square)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/julydate/acmeDeliver?style=flat-square)

 acme.sh 证书分发服务

将 acme.sh 获取的证书通过 http 服务分发到多台服务器

## Usage

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

## Client Example
```bash
# Get single file `mydomain.net.key` to current work folder
# 单独下载'mydomain.net.key'文件到当前工作目录
./client.sh -d "mydomain.net" -p "passcode" -s "myacmedeliverserver.net:8080" -n "mydomain.net.key"

# Automatically download certs only when server's certs' timestamp updates
# 仅在服务端证书的时间戳更新时自动下载证书密钥
./client.sh -d "mydomain.net" -p "passcode" -s "myacmedeliverserver.net:8080" -c
```

## Document

待更新，配套 bash 客户端开发中
