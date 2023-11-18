package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"github.com/julydate/acmeDeliver/app/handler"
	"github.com/julydate/acmeDeliver/app/mylego"
	"github.com/julydate/acmeDeliver/config"
)

func New(c *config.Config) *Controller {
	var legos []*mylego.LegoCMD
	for i := range c.CertConfig {
		lego, err := mylego.New(c.CertConfig[i])
		if err != nil {
			log.Error(err)
			continue
		}
		legos = append(legos, lego)
	}
	return &Controller{
		httpServe: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", c.Bind, c.Port),
			Handler: handler.New(c),
		},
		myLego:    legos,
		cronJob:   cron.New(),
		interval:  c.Interval,
		tlsConfig: &c.TlsConfig,
	}
}

func (c *Controller) Start() error {
	var certPath, keyPath string

	// Apply certs on start
	for i := range c.myLego {

		l := c.myLego[i]
		switch l.Conf.CertMode {
		case "dns":
			cert, key, err := l.DNSCert()
			if err != nil {
				log.Error(err)
			}
			if l.Conf.CertDomain == c.tlsConfig.Domain {
				certPath = cert
				keyPath = key
			}
		case "http", "tls":
			cert, key, err := l.HTTPCert()
			if err != nil {
				log.Error(err)
			}
			if l.Conf.CertDomain == c.tlsConfig.Domain {
				certPath = cert
				keyPath = key
			}
		default:
			log.Errorf("unsupported certmode: %s", l.Conf.CertMode)
		}
	}

	// cron job
	if _, err := c.cronJob.AddJob(fmt.Sprintf("@every %ds", c.interval),
		cron.NewChain(cron.SkipIfStillRunning(cron.DefaultLogger)).Then(c)); err != nil {
		log.Error(err)
	}

	if c.tlsConfig.Enable {
		if certPath == "" && keyPath == "" {
			return errors.New("cert file is not exist")
		}

		c.httpServe.Addr = fmt.Sprintf("%s:%d", c.tlsConfig.Bind, c.tlsConfig.Port)
		log.Infof("Start tls server on: \033[32m%s\033[0m (%s)", c.httpServe.Addr, c.tlsConfig.Domain)
		return c.httpServe.ListenAndServeTLS(certPath, keyPath)
	}

	log.Infof("Start server on: \033[32m%s\033[0m", c.httpServe.Addr)
	return c.httpServe.ListenAndServe()
}

func (c *Controller) Stop() error {
	log.Info("Stop server..")
	return c.httpServe.Shutdown(c.cronJob.Stop())
}

// Run  cron job
func (c *Controller) Run() {
	log.Info("Cron job of renew certs")
	for i := range c.myLego {
		l := c.myLego[i]
		_, _, _, err := l.RenewCert()
		if err != nil {
			log.Error(err)
		}
	}
}
