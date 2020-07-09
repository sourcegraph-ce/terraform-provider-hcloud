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

	res := &loadbalancer.RData{
		Name:         "some-load-balancer",
		LocationName: "nbg1",
		Labels: map[string]string{
			"key": "value",
		},
	}
	lbByName := &loadbalancer.DData{
		Name:             "lb_by_name",
		LoadBalancerName: res.HCLID() + ".name",
	}
	lbByID := &loadbalancer.DData{
		Name:           "lb_by_id",
		LoadBalancerID: res.HCLID() + ".id",
	}
	lbBySel := &loadbalancer.DData{
		Name:          "lb_by_sel",
		LabelSelector: fmt.Sprintf("key=${%s.labels[\"key\"]}", res.HCLID()),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", res,
					"testdata/d/hcloud_load_balancer", lbByName,
					"testdata/d/hcloud_load_balancer", lbByID,
					"testdata/d/hcloud_load_balancer", lbBySel,
				),

				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(lbByName.HCLID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(lbByName.HCLID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(lbByName.HCLID(), "target.#", "0"),
					resource.TestCheckResourceAttr(lbByName.HCLID(), "service.#", "0"),

					resource.TestCheckResourceAttr(lbByID.HCLID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(lbByID.HCLID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(lbByID.HCLID(), "target.#", "0"),
					resource.TestCheckResourceAttr(lbByID.HCLID(), "service.#", "0"),

					resource.TestCheckResourceAttr(lbBySel.HCLID(),
						"name", fmt.Sprintf("%s--%d", res.Name, tmplMan.RandInt)),
					resource.TestCheckResourceAttr(lbBySel.HCLID(), "location", res.LocationName),
					resource.TestCheckResourceAttr(lbBySel.HCLID(), "target.#", "0"),
					resource.TestCheckResourceAttr(lbBySel.HCLID(), "service.#", "0"),
				),
			},
		},
	})
}
