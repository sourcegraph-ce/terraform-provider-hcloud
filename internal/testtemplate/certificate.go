package testtemplate

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

// RCertificate defines the fields for the "testdata/r/hcloud_certificate"
// template.
type RCertificate struct {
	DataCommon

	Name        string
	PrivateKey  string
	Certificate string
}

// NewRCertificate creates data for a new certificate resource.
func NewRCertificate(t *testing.T, name, domain string) *RCertificate {
	rCert, rKey, err := acctest.RandTLSCert(domain)
	if err != nil {
		t.Fatal(err)
	}
	return &RCertificate{
		Name:        name,
		PrivateKey:  rKey,
		Certificate: rCert,
	}
}
