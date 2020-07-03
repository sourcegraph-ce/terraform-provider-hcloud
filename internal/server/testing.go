package server

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/terraform-providers/terraform-provider-hcloud/internal/testsupport"
)

func init() {
	resource.AddTestSweepers(ResourceType, &resource.Sweeper{
		Name:         ResourceType,
		Dependencies: []string{},
		F:            Sweep,
	})
}

// Sweep removes all Servers from the backend.
func Sweep(r string) error {
	client, err := testsupport.CreateClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	servers, err := client.Server.All(ctx)
	if err != nil {
		return err
	}

	for _, srv := range servers {
		if _, err := client.Server.Delete(ctx, srv); err != nil {
			return err
		}
	}

	return nil
}

// ByID returns a function that obtains a server by its ID.
func ByID(t *testing.T, srv *hcloud.Server) func(*hcloud.Client, int) bool {
	return func(c *hcloud.Client, id int) bool {
		found, _, err := c.Server.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("find server %d: %v", id, err)
		}
		if found == nil {
			return false
		}
		if srv != nil {
			*srv = *found
		}
		return true
	}
}
