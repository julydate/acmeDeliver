package handler

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zekroTJA/timedmap"

	"github.com/thinkeridea/go-extend/exnet"

	"github.com/julydate/acmeDeliver/config"
)

func New(c *config.Config) *Handler {
	h := &Handler{
		key:      c.Key,
		timeDiff: c.TimeDiff,
		cache:    timedmap.New(time.Second),
	}

	return h
}

func (h *Handler) validateTimestamp(t string, ip string, response http.ResponseWriter) (int64, bool) {
	reqTime, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		handleErrorResponse(response, ip, 403, "Timestamp not allowed.", "Incoming illegal timestamp:"+t)
		return 0, false
	}

	expireTime := time.Now().Unix() - reqTime

	if expireTime < -h.timeDiff {
		handleErrorResponse(response, ip, 403, "Timestamp not allowed.", fmt.Sprintf("Incoming illegal timestamp: %v", expireTime))
		return 0, false
	} else if expireTime > h.timeDiff {
		handleErrorResponse(response, ip, 403, "Timestamp expired.", fmt.Sprintf("Incoming expired access: %v", expireTime))
		return 0, false
	}
	return reqTime, true
}

func (h *Handler) validateChecksumReplay(checksum string, ip string, reqTime int64, response http.ResponseWriter) bool {
	if checkTime, ok := h.cache.GetValue(checksum).(int64); ok {
		if checkTime > 0 && time.Now().Unix()-checkTime > h.timeDiff {
			h.cache.Remove(checkTime)
		} else {
			handleErrorResponse(response, ip, 403, "Repeat access.", fmt.Sprintf("Incoming repeat access: %v", checksum))
			return false
		}
	} else {
		h.cache.Set(checksum, reqTime, time.Duration(h.timeDiff)*time.Second)
	}
	return true
}

func (h *Handler) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		log.Fatal("ParseForm:", err)
		return
	}

	ip := exnet.ClientPublicIP(req)
	if ip == "" {
		ip = exnet.ClientIP(req)
	}

	domain, err := checkValue(response, req.Form, "domain")
	if err != nil {
		return
	}

	file, err := checkValue(response, req.Form, "file")
	if err != nil {
		return
	}

	sign, err := checkValue(response, req.Form, "sign")
	if err != nil {
		return
	}

	checksum, err := checkValue(response, req.Form, "checksum")
	if err != nil {
		return
	}

	t, err := checkValue(response, req.Form, "t")
	if err != nil {
		return
	}

	reqTime, ok := h.validateTimestamp(t, ip, response)
	if !ok {
		return
	}

	if !h.validateChecksumReplay(checksum, ip, reqTime, response) {
		return
	}

	if !validateFileAndDomain(ip, domain, file, response) {
		return
	}

	checkToken := fmt.Sprintf("%x", md5.Sum([]byte(domain+file+h.key+t+checksum)))
	if sign == checkToken {
		filepath := path.Join("./certs", domain, "certificates", file)
		log.Infof("Access from IP: %s", ip)
		log.Infof("Access file: %s -> %s", domain, file)
		http.ServeFile(response, req, filepath)
	} else {
		handleErrorResponse(response, ip, 401, "Unauthorized access.", fmt.Sprintf("Incoming unauthorized access: %v", sign))
	}
}
