package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/idna"
)

func checkValue(response http.ResponseWriter, form url.Values, key string) (string, error) {
	if len(form.Get(key)) == 0 {
		response.WriteHeader(400)
		fmt.Fprintf(response, "no %s specified.", key)
		return "", fmt.Errorf("no %s specified", key)
	}

	return form.Get(key), nil
}

func handleErrorResponse(response http.ResponseWriter, ip string, statusCode int, errorMessage, printMessage string) {
	log.Infof("Access from IP: %s", ip)
	log.Infof(printMessage)
	response.WriteHeader(statusCode)
	fmt.Fprintf(response, errorMessage)
}

func validateFileAndDomain(ip string, domain string, file string, response http.ResponseWriter) bool {
	baseDir := "certs"
	domain = sanitizedDomain(domain)
	if _, err := os.Stat(path.Join(baseDir, domain)); os.IsNotExist(err) {
		handleErrorResponse(response, ip, 404, "Domain not exist.", fmt.Sprintf("Incoming illegal domain: %s", domain))
		return false
	}
	if _, err := os.Stat(path.Join(baseDir, domain, "certificates", file)); os.IsNotExist(err) {
		handleErrorResponse(response, ip, 404, "File not exist.", fmt.Sprintf("Incoming illegal filename: %s", file))
		return false
	}
	return true
}

func sanitizedDomain(domain string) string {
	safe, err := idna.ToASCII(strings.ReplaceAll(domain, "*", "_"))
	if err != nil {
		log.Error(err)
	}
	return safe
}
