package testtemplate

// RNetwork defines the fields for the "testdata/r/hcloud_network" template.
type RNetwork struct {
	DataCommon

	Name    string
	IPRange string
	Labels  map[string]string
}

// RNetworkSubnet defines the fields for the "testdata/r/hcloud_network_subnet"
// template.
type RNetworkSubnet struct {
	DataCommon

	Name        string
	Type        string
	NetworkID   string
	NetworkZone string
	IPRange     string
}
