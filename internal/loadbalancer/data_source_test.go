package loadbalancer_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudDataSourceLoadBalancerTest(t *testing.T) {
	tmplMan := testtemplate.Manager{}

	lbData := &loadbalancer.RData{
		Name:         "some-load-balancer",
		LocationName: "nbg1",
		Labels: map[string]string{
			"key": "value",
		},
	}
	lbResName := fmt.Sprintf("%s.%s", loadbalancer.ResourceType, lbData.Name)
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", lbData,
					"testdata/d/hcloud_load_balancer", &loadbalancer.DData{
						Name:             "lb_by_name",
						LoadBalancerName: fmt.Sprintf("%s.name", lbResName),
					},
					"testdata/d/hcloud_load_balancer", &loadbalancer.DData{
						Name:           "lb_by_id",
						LoadBalancerID: fmt.Sprintf("%s.id", lbResName),
					},
					"testdata/d/hcloud_load_balancer", &loadbalancer.DData{
						Name:          "lb_by_sel",
						LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", lbResName),
					},
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_name", loadbalancer.DataSourceType),
						"name", fmt.Sprintf("%s--%d", lbData.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_name", loadbalancer.DataSourceType),
						"location", lbData.LocationName),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_name", loadbalancer.DataSourceType),
						"target.#", "0"),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_name", loadbalancer.DataSourceType),
						"service.#", "0"),

					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_id", loadbalancer.DataSourceType),
						"name", fmt.Sprintf("%s--%d", lbData.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_id", loadbalancer.DataSourceType),
						"location", lbData.LocationName),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_id", loadbalancer.DataSourceType),
						"target.#", "0"),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_id", loadbalancer.DataSourceType),
						"service.#", "0"),

					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_sel", loadbalancer.DataSourceType),
						"name", fmt.Sprintf("%s--%d", lbData.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_sel", loadbalancer.DataSourceType),
						"location", lbData.LocationName),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_sel", loadbalancer.DataSourceType),
						"target.#", "0"),
					resource.TestCheckResourceAttr(fmt.Sprintf("data.%s.lb_by_sel", loadbalancer.DataSourceType),
						"service.#", "0"),
				),
			},
		},
	})
}
