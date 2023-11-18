package config

func DefaultConfig() *Config {
	return &Config{
		Bind:     "",
		Port:     9090,
		Key:      "passwd",
		TimeDiff: 60,
		Interval: 3600,
		TlsConfig: TlsConfig{
			Enable: false,
		},
		CertConfig: nil,
	}
}
