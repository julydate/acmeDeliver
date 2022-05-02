// Package legocmd Let's Encrypt client to go!
// CLI application for generating Let's Encrypt certificates using the ACME package.
package legocmd

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
)

// MyUser You'll need a user or account type that implements acme.User
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

type LegoCmd struct {
	Domain   string
	Email    string
	Key      string
	BaseDir  string
	Resource *certificate.Resource
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (l *LegoCmd) DNSCert() {

	// Create a user. DNSCert accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	myUser := MyUser{
		Email: "you@yours.com",
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// init config
	cfConfig := cloudflare.NewDefaultConfig()
	cfConfig.AuthEmail = l.Email
	cfConfig.AuthKey = l.Key
	cfConfig.PollingInterval = 10 * time.Second
	cfConfig.PropagationTimeout = 300 * time.Second
	cf, _ := cloudflare.NewDNSProviderConfig(cfConfig)

	err = client.Challenge.SetDNS01Provider(cf)
	if err != nil {
		log.Fatal(err)
	}

	// DNSCert users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	myUser.Registration = reg

	request := certificate.ObtainRequest{
		Domains: []string{l.Domain},
		Bundle:  true,
	}
	resource, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}
	l.Resource = resource
}

func (l *LegoCmd) RenewCert() error {
	certPath, _, err := l.CheckCertFile()
	if err != nil {
		return err
	}
	buf, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	cert, err := x509.ParseCertificate(buf)
	if err != nil {
		return err
	}
	if time.Until(cert.NotAfter) < 30*24*time.Hour {
		log.Printf("Cert will expire in 30 days, Renew now")
		l.DNSCert()
		err = os.WriteFile(path.Join(l.BaseDir, l.Domain, fmt.Sprintf("%s.key", l.Domain)), l.Resource.PrivateKey, 0644)
		err = os.WriteFile(path.Join(l.BaseDir, l.Domain, fmt.Sprintf("%s.crt", l.Domain)), l.Resource.Certificate, 0644)
		if err != nil {
			return err
		}
	}
	log.Printf("%0.0f\n days to expired, No renew now", time.Until(cert.NotAfter).Hours()/24)
	return nil
}

func (l *LegoCmd) CheckCertFile() (string, string, error) {
	keyPath := path.Join(l.BaseDir, l.Domain, fmt.Sprintf("%s.key", l.Domain))
	certPath := path.Join(l.BaseDir, l.Domain, fmt.Sprintf("%s.crt", l.Domain))
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("cert key failed: %s", l.Domain)
	}
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("cert cert failed: %s", l.Domain)
	}
	absKeyPath, _ := filepath.Abs(keyPath)
	absCertPath, _ := filepath.Abs(certPath)
	return absCertPath, absKeyPath, nil
}
