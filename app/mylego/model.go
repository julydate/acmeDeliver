package mylego

type CertConfig struct {
	CertMode   string            `yaml:"CertMode"`
	CertDomain string            `yaml:"CertDomain"`
	Provider   string            `yaml:"Provider"`
	Email      string            `yaml:"Email"`
	DNSEnv     map[string]string `yaml:"DNSEnv"`
}

type LegoCMD struct {
	Conf *CertConfig
	path string
}
