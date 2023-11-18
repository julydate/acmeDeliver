package controller

import (
	"net/http"

	"github.com/robfig/cron/v3"

	"github.com/julydate/acmeDeliver/app/mylego"
	"github.com/julydate/acmeDeliver/config"
)

type Controller struct {
	httpServe *http.Server
	myLego    []*mylego.LegoCMD
	cronJob   *cron.Cron
	interval  int
	tlsConfig *config.TlsConfig
}
