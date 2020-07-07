package testtemplate

// RServer defines the fields for the "testdata/r/hcloud_server" template.
type RServer struct {
	DataCommon

	Name         string
	Type         string
	Image        string
	LocationName string
	DataCenter   string
	SSHKeys      []string
	KeepDisk     bool
	Rescue       bool
	Backups      bool
	Labels       map[string]string
	UserData     string
}

// RServerNetwork defines the fields for the "testdata/r/hcloud_server_network"
// template.
type RServerNetwork struct {
	DataCommon

	Name      string
	ServerID  string
	NetworkID string
	IP        string
	AliasIPs  []string
}
