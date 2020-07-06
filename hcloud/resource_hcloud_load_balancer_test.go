package hcloud

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func init() {
	resource.AddTestSweepers("hcloud_load_balancer", &resource.Sweeper{
		Name: "hcloud_load_balancer",
		F: func(r string) error {
			var mErr error

			if err := testSweepLoadBalancers(r); err != nil {
				mErr = multierror.Append(mErr, err)
			}
			if err := testSweepCertificates(r); err != nil {
				mErr = multierror.Append(mErr, err)
			}
			return mErr
		},
	})
}

// Deprecated: there is a generic destroy check in testsupport
func testAccHcloudCheckLoadBalancerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*hcloud.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hcloud_load_balancer" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Load Balancer id is no int: %v", err)
		}
		var loadBalancer *hcloud.LoadBalancer
		loadBalancer, _, err = client.LoadBalancer.GetByID(context.Background(), id)
		if err != nil {
			return fmt.Errorf(
				"Error checking if Load Balancer (%s) is deleted: %v",
				rs.Primary.ID, err)
		}
		if loadBalancer != nil {
			return fmt.Errorf("Load Balancer (%s) has not been deleted", rs.Primary.ID)
		}
	}

	return nil
}

// Deprecated: there is a generic existence check in testsupport
func testAccHcloudCheckLoadBalancerExists(n string, loadBalancer *hcloud.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*hcloud.Client)
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Try to find the key
		foundLoadBalancer, _, err := client.LoadBalancer.GetByID(context.Background(), id)
		if err != nil {
			return err
		}

		if foundLoadBalancer == nil {
			return fmt.Errorf("Record not found")
		}

		*loadBalancer = *foundLoadBalancer
		return nil
	}
}

// Deprecated: moved to internal/loadbalancers
func testSweepLoadBalancers(region string) error {
	client, err := createClient()
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
	testSweepServers(region)
	return nil
}
