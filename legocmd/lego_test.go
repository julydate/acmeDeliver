package legocmd

import (
	"testing"
)

func TestLegoClient(t *testing.T) {
	l := LegoCmd{
		Domain:   "test.com",
		Email:    "test@test.com",
		Key:      "test",
		BaseDir:  "./",
		Resource: nil,
	}
	l.DNSCert()
}
