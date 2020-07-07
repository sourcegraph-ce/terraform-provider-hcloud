package loadbalancer

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/network"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/server"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testtemplate"
)

func init() {
	resource.AddTestSweepers(ResourceType, &resource.Sweeper{
		Name: ResourceType,
		Dependencies: []string{
			server.ResourceType,
			network.ResourceType,
		}, // TODO add certificates and possible more sweepers
		F: Sweep,
	})
}

// Basic Load Balancer for use in load balancer related test.
//
// Do not modify!
var Basic= &testtemplate.RLoadBalancer{
	Name:         "basic-load-balancer",
	LocationName: "nbg1",
}

// Sweep removes all Load Balancers from the backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	loadBalancers, err := client.LoadBalancer.All(ctx)
	if err != nil {
		return err
	}

	for _, loadBalancer := range loadBalancers {
		if _, err := client.LoadBalancer.Delete(ctx, loadBalancer); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a loadbalancer by its ID.
func ByID(t *testing.T, lb *hcloud.LoadBalancer) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.LoadBalancer.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find load balancer %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if lb != nil {
			*lb = *found
		}
		return true
	}
}
