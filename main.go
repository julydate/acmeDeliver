package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/thinkeridea/go-extend/exnet"
	"github.com/zekroTJA/timedmap"
)

// VERSION 版本号
const VERSION = "1.1"

// h 帮助信息
var h bool

// bind 绑定监听地址
var bind string

// port 服务端口
var port string

// tls TLS 监听开关
var tls bool

// tlsport TLS 服务端口
var tlsport string

// certFile TLS 服务证书文件
var certFile string

// keyFile TLS 服务私钥文件
var keyFile string

// baseDir 证书文件所在目录
var baseDir string

// key 密码
var key string

// timeRange 时间戳误差，单位秒
var timeRange int64

// 初始化从客户端获取的全局变量
var domain, file, t, checksum, sign string

// Creates a new timed map which scans for expired keys every 1 second
var tm = timedmap.New(1 * time.Second)

func init() {
	// 初始化从命令行获取参数
	flag.BoolVar(&h, "h", false, "显示帮助信息")
	flag.StringVar(&bind, "b", "", "绑定监听地址,默认绑定所有接口")
	flag.StringVar(&port, "p", "9090", "服务端口,默认 9090")
	flag.BoolVar(&tls, "tls", false, "是否监听 TLS,默认关闭")
	flag.StringVar(&tlsport, "tlsport", "9443", "TLS 服务端口,默认 9443")
	flag.StringVar(&certFile, "cert", "cert.pem", "TLS 服务证书文件,默认 cert.pem")
	flag.StringVar(&keyFile, "key", "key.pem", "TLS 服务私钥文件,默认 key.pem")
	flag.StringVar(&baseDir, "d", "./", "证书文件所在目录,默认当前目录")
	flag.StringVar(&key, "k", "passwd", "密码,默认 passwd")
	flag.Int64Var(&timeRange, "t", 60, "时间戳误差,默认 60 秒")
	// 修改默认 Usage
	flag.Usage = usage
}

func main() {
	// 从 arguments 中解析注册的 flag
	// 必须在所有 flag 都注册好而未访问其值时执行
	// 未注册却使用 flag -help 时，会返回 ErrHelp
	flag.Parse()

	// 显示帮助信息
	if h {
		flag.Usage()
		return
	}

	// 设置访问的路由
	http.HandleFunc("/", check)

	// 启动 TLS 端口监听
	if tls {
		// 使用 goroutine 进行监听防阻塞
		go func() {
			if err := http.ListenAndServeTLS(bind+":"+tlsport, certFile, keyFile, nil); err != nil {
				log.Fatal("ListenAndServeTLS:", err)
			}
		}()
	}

	// 启动端口监听
	if err := http.ListenAndServe(bind+":"+port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// 自定义帮助信息
func usage() {
	fmt.Fprintf(os.Stderr, `acmeDeliver version: `+VERSION+`
Usage: acmeDeliver [-h] [-p port] [-d dirname] [-k password] [-t time] [-b address] [-tls] [-tlsport port] [-cert filename] [-key filename]

Options:
`)
	flag.PrintDefaults()
}

func check(response http.ResponseWriter, req *http.Request) {
	// 解析 url 传递的参数，对于 POST 则解析响应包的主体（request body）
	err := req.ParseForm()
	if err != nil {
		log.Fatal("ParseForm:", err)
		return
	}

	// 获取来访 IP 地址
	var ip = exnet.ClientPublicIP(req)
	if ip == "" {
		ip = exnet.ClientIP(req)
	}

	// 获取传入域名
	if len(req.Form.Get("domain")) == 0 {
		response.WriteHeader(400)
		fmt.Fprintf(response, "No domain specified.")
		return
	}
	domain = req.Form.Get("domain")
	// 获取传入文件名
	if len(req.Form.Get("file")) == 0 {
		response.WriteHeader(400)
		fmt.Fprintf(response, "No file specified.")
		return
	}
	file = req.Form.Get("file")
	// 获取传入签名
	if len(req.Form.Get("sign")) == 0 {
		response.WriteHeader(400)
		fmt.Fprintf(response, "No sign specified.")
		return
	}
	sign = req.Form.Get("sign")
	// 获取传入验证码
	if len(req.Form.Get("checksum")) == 0 {
		response.WriteHeader(400)
		fmt.Fprintf(response, "No checksum specified.")
		return
	}
	checksum = req.Form.Get("checksum")
	// 获取传入时间戳
	if len(req.Form.Get("t")) == 0 {
		response.WriteHeader(400)
		fmt.Fprintf(response, "No timestamp specified.")
		return
	}
	t = req.Form.Get("t")

	// 校验时间戳是否合法
	reqTime, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		fmt.Println("Access from IP:", ip)
		fmt.Println("Incoming illegal timestamp:", t)
		response.WriteHeader(403)
		fmt.Fprintf(response, "Timestamp not allowed.")
		return
	}
	expireTime := time.Now().Unix() - reqTime
	// 时间戳太超前可以判定为异常访问
	if expireTime < -timeRange {
		fmt.Println("Access from IP:", ip)
		fmt.Println("Incoming illegal timestamp:", expireTime)
		response.WriteHeader(403)
		fmt.Fprintf(response, "Timestamp not allowed.")
		return
	}
	// 校验时间戳是否过期
	if expireTime > timeRange {
		fmt.Println("Access from IP:", ip)
		fmt.Println("Incoming expired access:", expireTime)
		response.WriteHeader(403)
		fmt.Fprintf(response, "Timestamp expired.")
		return
	}

	// 计算 token
	token := md5.New()
	io.WriteString(token, domain)
	io.WriteString(token, file)
	io.WriteString(token, key)
	io.WriteString(token, t)
	io.WriteString(token, checksum)
	checkToken := fmt.Sprintf("%x", token.Sum(nil))

	// 校验签名
	if sign == checkToken {
		// 检测验证码是否重复
		if checkTime, ok := tm.GetValue(checksum).(int64); ok {
			if checkTime > 0 && time.Now().Unix()-checkTime > timeRange {
				tm.Remove(checkTime)
			} else {
				// 检测到重放请求
				fmt.Println("Access from IP:", ip)
				fmt.Println("Incoming repeat access:", checksum)
				response.WriteHeader(403)
				fmt.Fprintf(response, "Repeat access.")
				return
			}
		} else {
			tm.Set(checksum, reqTime, time.Duration(timeRange)*time.Second)
		}

		// 校验域名是否在指定文件夹内
		dirInfo, err := os.Stat(path.Join(baseDir, domain))
		if err != nil || !dirInfo.IsDir() {
			fmt.Println("Access from IP:", ip)
			fmt.Println("Incoming illegal domain:", domain)
			response.WriteHeader(404)
			fmt.Fprintln(response, "Domain not exist.")
			return
		}
		//
		fileInfo, err := os.Stat(path.Join(baseDir, domain, file))
		if err != nil || fileInfo.IsDir() {
			fmt.Println("Access from IP:", ip)
			fmt.Println("Incoming illegal domain:", domain)
			response.WriteHeader(404)
			fmt.Fprintf(response, "Certificate not found.")
			return
		}
		// 全部校验通过，放行文件
		filepath := path.Join(baseDir, domain, file)
		fmt.Println("Access from IP:", ip)
		fmt.Println("Access file:", filepath)
		http.ServeFile(response, req, filepath)
	} else {
		// 签名错误
		fmt.Println("Access from IP:", ip)
		fmt.Println("Incoming unauthorized access:", sign)
		response.WriteHeader(401)
		fmt.Fprintf(response, "Unauthorized access.")
	}
}
