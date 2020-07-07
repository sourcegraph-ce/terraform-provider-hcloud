package testtemplate

// RLoadBalancer defines the fields for the "testdata/r/hcloud_load_balancer"
// template.
type RLoadBalancer struct {
	DataCommon

	Name          string
	Type          string
	LocationName  string
	NetworkZone   string
	Algorithm     string
	ServerTargets []RLoadBalancerInlineServerTarget
	Labels        map[string]string
}

// RLoadBalancerInlineServerTarget represents a Load Balancer server target
// that is added inline to the Load Balancer.
type RLoadBalancerInlineServerTarget struct {
	ServerID string
}

// RLoadBalancerService defines the fields for the
// "testdata/r/hcloud_load_balancer_service" template.
type RLoadBalancerService struct {
	DataCommon

	Name            string
	LoadBalancerID  string
	Protocol        string
	ListenPort      int
	DestinationPort int
	Proxyprotocol   bool
}

// RLoadBalancerTarget defines the fields for the
// "testdata/r/hcloud_load_balancer_target" template.
type RLoadBalancerTarget struct {
	DataCommon

	Name           string
	Type           string
	LoadBalancerID string
	ServerID       string
	UsePrivateIP   bool
	DependsOn      []string
}

// RLoadBalancerNetwork defines the fields for the
// "testdata/r/hcloud_load_balancer_network" template.
type RLoadBalancerNetwork struct {
	DataCommon

	Name                  string
	LoadBalancerID        string
	NetworkID             string
	IP                    string
	EnablePublicInterface bool
}
