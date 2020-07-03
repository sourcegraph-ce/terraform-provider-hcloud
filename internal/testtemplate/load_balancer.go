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
