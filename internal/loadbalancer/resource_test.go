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

	res := loadbalancer.Basic
	resRenamed := &loadbalancer.RData{
		Name:         res.Name + "-renamed",
		LocationName: "nbg1",
		Algorithm:    "least_connections",
		Labels: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, &lb)),
		Steps: []resource.TestStep{
			{
				// Create a new Load Balancer using the required values
				// only.
				Config: tmplMan.Render(t, "testdata/r/hcloud_load_balancer", res),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(res.HCLID(), loadbalancer.ByID(t, &lb)),
					resource.TestCheckResourceAttr(res.HCLID(), "name",
						fmt.Sprintf("basic-load-balancer--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(res.HCLID(), "load_balancer_type", "lb11"),
					resource.TestCheckResourceAttr(res.HCLID(), "location", "nbg1"),
				),
			},
			{
				// Try to import the newly created load balancer
				ResourceName:      res.HCLID(),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update the Load Balancer created in the previous step by
				// setting all optional fields and renaming the load
				// balancer.
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", resRenamed,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resRenamed.HCLID(), "name",
						fmt.Sprintf("basic-load-balancer-renamed--%d", tmplMan.RandInt)),
					resource.TestCheckResourceAttr(resRenamed.HCLID(), "load_balancer_type", "lb11"),
					resource.TestCheckResourceAttr(resRenamed.HCLID(), "location", "nbg1"),
					resource.TestCheckResourceAttr(resRenamed.HCLID(), "algorithm.0.type", "least_connections"),
					resource.TestCheckResourceAttr(resRenamed.HCLID(), "labels.key1", "value1"),
					resource.TestCheckResourceAttr(resRenamed.HCLID(), "labels.key2", "value2"),
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
