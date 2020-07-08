package testtemplate

// RCertificate defines the fields for the "testdata/r/hcloud_certificate"
// template.
type RCertificate struct {
	DataCommon

	Name        string
	PrivateKey  string
	Certificate string
}
