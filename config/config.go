package config

func DefaultConfig() *Config {
	return &Config{
		Bind:     "",
		Port:     9090,
		Tls:      false,
		TlsPort:  9443,
		Key:      "passwd",
		TimeDiff: 60,
	}
}
