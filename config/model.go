package config

import "github.com/julydate/acmeDeliver/app/mylego"

type Config struct {
	Bind      string    `yaml:"Bind"`
	Port      int       `yaml:"Port"`
	Key       string    `yaml:"Key"`
	TimeDiff  int64     `yaml:"TimeDiff"`
	Interval  int       `yaml:"Interval"`
	TlsConfig TlsConfig `yaml:"TlsConfig"`

	CertConfig []*mylego.CertConfig `yaml:"CertConfig"`
}

type TlsConfig struct {
	Enable bool   `yaml:"Enable"`
	Domain string `yaml:"Domain"`
	Bind   string `yaml:"Bind"`
	Port   int    `yaml:"Port"`
}
