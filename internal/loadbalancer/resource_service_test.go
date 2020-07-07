package loadbalancer_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/loadbalancer"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func TestAccHcloudLoadBalancerService_TCP(t *testing.T) {
	var lb hcloud.LoadBalancer

	lbResName := fmt.Sprintf("%s.%s", loadbalancer.ResourceType, loadbalancer.Basic.Name)
	svcName := "lb-tcp-service-test"
	svcResName := fmt.Sprintf("%s.%s", loadbalancer.ServiceResourceType, svcName)

	tmplMan := testtemplate.Manager{}
	resource.Test(t, resource.TestCase{
		PreCheck:     testsupport.AccTestPreCheck(t),
		Providers:    testsupport.AccTestProviders(),
		CheckDestroy: testsupport.CheckResourcesDestroyed(loadbalancer.ResourceType, loadbalancer.ByID(t, nil)),
		Steps: []resource.TestStep{
			{
				Config: tmplMan.Render(t,
					"testdata/r/hcloud_load_balancer", loadbalancer.Basic,
					"testdata/r/hcloud_load_balancer_service", &testtemplate.RLoadBalancerService{
						Name:            svcName,
						Protocol:        "tcp",
						LoadBalancerID:  fmt.Sprintf("%s.%s.id", loadbalancer.ResourceType, loadbalancer.Basic.Name),
						ListenPort:      70,
						DestinationPort: 70,
						Proxyprotocol:   true,
					},
				),
				Check: resource.ComposeTestCheckFunc(
					testsupport.CheckResourceExists(lbResName, loadbalancer.ByID(t, &lb)),
					testsupport.LiftTCF(hasService(&lb, 70)),
					testsupport.CheckResourceAttrFunc(svcResName, "load_balancer_id", func() string {
						return strconv.Itoa(lb.ID)
					}),
					resource.TestCheckResourceAttr(svcResName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(svcResName, "listen_port", "70"),
					resource.TestCheckResourceAttr(svcResName, "destination_port", "70"),
					resource.TestCheckResourceAttr(svcResName, "proxyprotocol", "true"),
				),
			},
		},
	})
}

func hasService(lb *hcloud.LoadBalancer, listenPort int) func() error {
	return func() error {
		for _, svc := range lb.Services {
			if svc.ListenPort == listenPort {
				return nil
			}
		}
		return fmt.Errorf("listen port %d: service not found", listenPort)
	}
}
