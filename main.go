package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/thinkeridea/go-extend/exnet"
	"github.com/zekroTJA/timedmap"
)

// PORT 服务端口
const PORT string = "9090"

// DIR 证书文件所在目录
const DIR string = "./"

// KEY 密码
const KEY string = "passwd"

// EXPTIME 时间戳误差，单位秒
const EXPTIME int64 = 86400

// 初始化全局变量
var domain, file, t, checksum, sign string

// Creates a new timed map which scans for expired keys every 1 second
var tm = timedmap.New(1 * time.Second)

func main() {
	// 设置访问的路由
	http.HandleFunc("/", check)
	// 设置监听的端口
	err := http.ListenAndServe(":"+PORT, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
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
		fmt.Fprintf(response, "No domain specified.")
		return
	}
	domain = req.Form.Get("domain")
	// 获取传入文件名
	if len(req.Form.Get("file")) == 0 {
		fmt.Fprintf(response, "No file specified.")
		return
	}
	file = req.Form.Get("file")
	// 获取传入签名
	if len(req.Form.Get("sign")) == 0 {
		fmt.Fprintf(response, "No sign specified.")
		return
	}
	sign = req.Form.Get("sign")
	// 获取传入验证码
	if len(req.Form.Get("checksum")) == 0 {
		fmt.Fprintf(response, "No checksum specified.")
		return
	}
	checksum = req.Form.Get("checksum")
	// 获取传入时间戳
	if len(req.Form.Get("t")) == 0 {
		fmt.Fprintf(response, "No timestamp specified.")
		return
	}
	t = req.Form.Get("t")

	// 校验时间戳是否合法
	reqTime, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		fmt.Println("Access from IP:", ip)
		fmt.Println("Incoming illegal timestamp:", t)
		fmt.Fprintf(response, "Timestamp not allowed.")
		return
	}
	expireTime := time.Now().Unix() - reqTime
	// 时间戳太超前可以判定为异常访问
	if expireTime < -EXPTIME {
		fmt.Println("Access from IP:", ip)
		fmt.Println("Incoming illegal timestamp:", expireTime)
		fmt.Fprintf(response, "Timestamp not allowed.")
		return
	}
	// 校验时间戳是否过期
	if expireTime > EXPTIME {
		fmt.Println("Access from IP:", ip)
		fmt.Println("Incoming expired access:", expireTime)
		fmt.Fprintf(response, "Timestamp expired.")
		return
	}

	// 计算 token
	token := md5.New()
	io.WriteString(token, domain)
	io.WriteString(token, file)
	io.WriteString(token, KEY)
	io.WriteString(token, t)
	io.WriteString(token, checksum)
	checkToken := fmt.Sprintf("%x", token.Sum(nil))

	// 校验签名
	if sign == checkToken {
		// 检测验证码是否重复
		if checkTime, ok := tm.GetValue(checksum).(int64); ok {
			if checkTime > 0 && time.Now().Unix()-checkTime > EXPTIME {
				tm.Remove(checkTime)
			} else {
				// 检测到重放请求
				fmt.Println("Access from IP:", ip)
				fmt.Println("Incoming repeat access:", checksum)
				fmt.Fprintf(response, "Repeat access.")
				return
			}
		} else {
			tm.Set(checksum, reqTime, time.Duration(EXPTIME)*time.Second)
		}

		// 校验域名是否在指定文件夹内
		var checkDomain, checkFile bool = false, false
		files, _ := ioutil.ReadDir(DIR)
		for _, f := range files {
			if domain == f.Name() {
				checkDomain = true
			}
		}
		if checkDomain {
			// 对应域名的文件夹存在，校验内部文件是否存在
			files, _ := ioutil.ReadDir(DIR + domain)
			for _, f := range files {
				if file == f.Name() {
					checkFile = true
				}
			}
		} else {
			// 获取的域名不存在
			fmt.Println("Access from IP:", ip)
			fmt.Println("Incoming illegal domain:", domain)
			fmt.Fprintf(response, "Domain not exist.")
			return
		}
		if !checkFile {
			// 获取的文件不存在
			fmt.Println("Access from IP:", ip)
			fmt.Println("Incoming illegal filename:", file)
			fmt.Fprintf(response, "File not exist.")
			return
		}
		// 全部校验通过，放行文件
		filepath := DIR + domain + "/" + file
		fmt.Println("Access from IP:", ip)
		fmt.Println("Access file:", filepath)
		http.ServeFile(response, req, filepath)
	} else {
		// 签名错误
		fmt.Println("Access from IP:", ip)
		fmt.Println("Incoming unauthorized access:", sign)
		fmt.Fprintf(response, "Unauthorized access.")
	}
}
