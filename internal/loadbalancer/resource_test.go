package loadbalancer_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/server"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestLoadBalancerResource_Basic(t *testing.T) {
	var lb hcloud.LoadBalancer

	tmplMan := testtemplate.Manager{RandInt: acctest.RandInt()}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, &lb)),
		Steps: []resource.TestStep{
			{
				// Create a new Load Balancer using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_load_balancer", loadbalancer.Basic),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(loadbalancer.ResourceType+".basic-load-balancer",
						loadbalancer.ByID(t, &lb)),
					resource.TestCheckResourceAttr(loadbalancer.ResourceType+".basic-load-balancer",
						"name", fmt.Sprintf("basic-load-balancer--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(loadbalancer.ResourceType+".basic-load-balancer",
						"load_balancer_type", "lb11"),
					resource.TestCheckResourceAttr(loadbalancer.ResourceType+".basic-load-balancer",
						"location", "nbg1"),
				),
			},
			{
				// Try to import the newly created load balancer
				ResourceName:      loadbalancer.ResourceType + "." + loadbalancer.Basic.Name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the Load Balancer created in the previous step by
				// setting all optional fields and renaming the load
				// balancer.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:         loadbalancer.Basic.Name + "-renamed",
						LocationName: "nbg1",
						Algorithm:    "least_connections",
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(loadbalancer.ResourceType+".basic-load-balancer-renamed",
						"name", fmt.Sprintf("basic-load-balancer-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(loadbalancer.ResourceType+".basic-load-balancer-renamed",
						"load_balancer_type", "lb11"),
					resource.TestCheckResourceAttr(loadbalancer.ResourceType+".basic-load-balancer-renamed",
						"location", "nbg1"),
					resource.TestCheckResourceAttr(loadbalancer.ResourceType+".basic-load-balancer-renamed",
						"algorithm.0.type", "least_connections"),
					resource.TestCheckResourceAttr(loadbalancer.ResourceType+".basic-load-balancer-renamed",
						"labels.key1", "value1"),
					resource.TestCheckResourceAttr(loadbalancer.ResourceType+".basic-load-balancer-renamed",
						"labels.key2", "value2"),
				),
			},
		},
	})
}

func TestLoadBalancerResource_InlineTarget(t *testing.T) {
	var srv0, srv1 hcloud.Server

	tmplMan := testtemplate.Manager{RandInt: acctest.RandInt()}
	resource.Test(t, resource.TestCase{
		PreCheck:  testsupport.AccTestPreCheck(t),
		Providers: testsupport.AccTestProviders(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testsupport.CheckResourcesDestroyed(server.ResourceType, server.ByID(t, nil)),
			testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		),
		Steps: []resource.TestStep{
			{
				// Add two inline targets to the load balancer
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_server", &server.RData{
						Name:  "some-server",
						Type:  "cx11",
						Image: "ubuntu-20.04",
					},
					"testdata/r/hcloud_server", &server.RData{
						Name:  "another-server",
						Type:  "cx11",
						Image: "ubuntu-20.04",
					},
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:         "some-lb",
						LocationName: "nbg1",
						Algorithm:    "least_connections",
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
						ServerTargets: []loadbalancer.RDataInlineServerTarget{
							{ServerID: "hcloud_server.some-server.id"},
							{ServerID: "hcloud_server.another-server.id"},
						},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testsupport.CheckResourceExists(server.ResourceType+".some-server", server.ByID(t, &srv0)),
					testsupport.CheckResourceExists(server.ResourceType+".another-server", server.ByID(t, &srv1)),
					testsupport.CheckResourceAttrFunc(loadbalancer.ResourceType+".some-lb",
						"target.0.server_id", func() string {
							return strconv.Itoa(srv0.ID)
						}),
					testsupport.CheckResourceAttrFunc(loadbalancer.ResourceType+".some-lb",
						"target.1.server_id", func() string {
							return strconv.Itoa(srv1.ID)
						}),
				),
			},
			{
				// Remove the targets from the load balancer
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", &loadbalancer.RData{
						Name:         "some-lb",
						LocationName: "nbg1",
						Algorithm:    "least_connections",
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				),
				Check: resource.TestCheckNoResourceAttr(loadbalancer.ResourceType+".some-lb", "target"),
			},
		},
	})
}
