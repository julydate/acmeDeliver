package config

import "github.com/julydate/acmeDeliver/app/mylego"

type Config struct {
	Bind     string `yaml:"Bind"`
	Port     int    `yaml:"Port"`
	Tls      bool   `yaml:"Tls"`
	TlsPort  int    `yaml:"TlsPort"`
	Key      string `yaml:"Key"`
	TimeDiff int64  `yaml:"TimeDiff"`
	Interval int    `yaml:"Interval"`

	CertConfig []*mylego.CertConfig `yaml:"CertConfig"`
}
