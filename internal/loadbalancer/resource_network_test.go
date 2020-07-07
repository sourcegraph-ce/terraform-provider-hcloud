package loadbalancer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/stretchr/testify/assert"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/network"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudLoadBalancerNetwork(t *testing.T) {
	var (
		nw hcloud.Network
		lb hcloud.LoadBalancer
	)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", &testtemplate.RNetwork{
						Name:    "test-network",
						IPRange: "10.0.0.0/16",
					},
					"testdata/r/hcloud_network_subnet", &testtemplate.RNetworkSubnet{
						Name:        "test-network-subnet",
						Type:        "cloud",
						NetworkID:   "hcloud_network.test-network.id",
						NetworkZone: "eu-central",
						IPRange:     "10.0.1.0/24",
					},
					"testdata/r/hcloud_load_balancer", &testtemplate.RLoadBalancer{
						Name:        "lb-network-test",
						Type:        "lb11",
						NetworkZone: "eu-central",
					},
					"testdata/r/hcloud_load_balancer_network", &testtemplate.RLoadBalancerNetwork{
						Name:                  "test-network",
						LoadBalancerID:        "hcloud_load_balancer.lb-network-test.id",
						NetworkID:             "hcloud_network.test-network.id",
						IP:                    "10.0.1.5",
						EnablePublicInterface: false,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(network.ResourceType+".test-network", network.ByID(t, &nw)),
					testsupport.CheckResourceExists(
						loadbalancer.ResourceType+".lb-network-test", loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(func() error {
						var privNet *hcloud.LoadBalancerPrivateNet
						for _, n := range lb.PrivateNet {
							if n.Network.ID == nw.ID {
								privNet = &n
								break
							}
						}
						if privNet == nil {
							return fmt.Errorf("load balancer has no private network")
						}
						assert.Equal(t, "10.0.1.5", privNet.IP.String())
						return nil
					}),
					resource.TestCheckResourceAttr(
						loadbalancer.NetworkResourceType+".test-network", "ip", "10.0.1.5"),
					resource.TestCheckResourceAttr(
						loadbalancer.NetworkResourceType+".test-network", "enable_public_interface", "false"),
				),
			},
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_network", &testtemplate.RNetwork{
						Name:    "test-network",
						IPRange: "10.0.0.0/16",
					},
					"testdata/r/hcloud_network_subnet", &testtemplate.RNetworkSubnet{
						Name:        "test-network-subnet",
						Type:        "cloud",
						NetworkID:   "hcloud_network.test-network.id",
						NetworkZone: "eu-central",
						IPRange:     "10.0.1.0/24",
					},
					"testdata/r/hcloud_load_balancer", &testtemplate.RLoadBalancer{
						Name:        "lb-network-test",
						Type:        "lb11",
						NetworkZone: "eu-central",
					},
					"testdata/r/hcloud_load_balancer_network", &testtemplate.RLoadBalancerNetwork{
						Name:                  "test-network",
						LoadBalancerID:        "hcloud_load_balancer.lb-network-test.id",
						NetworkID:             "hcloud_network.test-network.id",
						IP:                    "10.0.1.5",
						EnablePublicInterface: true,
					},
				),
				Check: resource.TestCheckResourceAttr(
					loadbalancer.NetworkResourceType+".test-network", "enable_public_interface", "true"),
			},
		},
	})
}
