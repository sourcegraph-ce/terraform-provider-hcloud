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
