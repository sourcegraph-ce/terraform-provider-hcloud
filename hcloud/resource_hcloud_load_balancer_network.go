package hcloud

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sourcegraph-ce/logrus"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceLoadBalancerNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceLoadBalancerNetworkCreate,
		Read:   resourceLoadBalancerNetworkRead,
		Delete: resourceLoadBalancerNetworkDelete,
		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"load_balancer_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"ip": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"enable_public_interface": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  true,
			},
		},
	}
}

func resourceLoadBalancerNetworkCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	ip := net.ParseIP(d.Get("ip").(string))
	networkID := d.Get("network_id")
	loadBalancerID := d.Get("load_balancer_id")

	loadBalancer := &hcloud.LoadBalancer{ID: loadBalancerID.(int)}

	network := &hcloud.Network{ID: networkID.(int)}
	opts := hcloud.LoadBalancerAttachToNetworkOpts{
		Network: network,
		IP:      ip,
	}
	action, _, err := client.LoadBalancer.AttachToNetwork(ctx, loadBalancer, opts)
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeConflict) ||
			hcloud.IsError(err, hcloud.ErrorCodeLocked) ||
			hcloud.IsError(err, hcloud.ErrorCodeServiceError) {
			hcErr := err.(hcloud.Error)
			log.Printf("[INFO] Network (%v) %s, retrying in one second", network.ID, hcErr.Code)
			time.Sleep(time.Second)
			return resourceLoadBalancerNetworkCreate(d, m)
		} else if string(err.(hcloud.Error).Code) == "load_balancer_already_attached" { // TODO: Change to correct error code and hcloud.IsError with next hcloud-go release
			log.Printf("[INFO] Load Balancer (%v) already attachted to network %v", loadBalancer.ID, network.ID)
			d.SetId(generateLoadBalancerNetworkID(loadBalancer, network))

			return resourceLoadBalancerNetworkRead(d, m)
		}
		return err
	}
	if err := waitForNetworkAction(ctx, client, action, network); err != nil {
		return err
	}

	enablePublicInterface := d.Get("enable_public_interface").(bool)
	err = resourceLoadBalancerNetworkUpdatePublicInterface(ctx, enablePublicInterface, loadBalancer, client, d)
	if err != nil {
		return err
	}
	d.SetId(generateLoadBalancerNetworkID(loadBalancer, network))

	return resourceLoadBalancerNetworkRead(d, m)
}

func resourceLoadBalancerNetworkRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	server, network, privateNet, err := lookupLoadBalancerNetworkID(ctx, d.Id(), client)
	if err == errInvalidLoadBalancerNetworkID {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
	}
	if server == nil {
		log.Printf("[WARN] LoadBalancer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if network == nil {
		log.Printf("[WARN] Network (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if privateNet == nil {
		log.Printf("[WARN] LoadBalancer Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	d.SetId(generateLoadBalancerNetworkID(server, network))
	setLoadBalancerNetworkSchema(d, server, network, privateNet)
	return nil

}

func resourceLoadBalancerNetworkDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	server, network, _, err := lookupLoadBalancerNetworkID(ctx, d.Id(), client)

	if err != nil {
		log.Printf("[WARN] Invalid id (%s), removing from state: %s", d.Id(), err)
		d.SetId("")
		return nil
	}
	action, _, err := client.LoadBalancer.DetachFromNetwork(ctx, server, hcloud.LoadBalancerDetachFromNetworkOpts{
		Network: network,
	})
	if err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// network has already been deleted
			return nil
		} else if hcloud.IsError(err, hcloud.ErrorCodeConflict) {
			log.Printf("[INFO] Network (%v) conflict, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceLoadBalancerNetworkDelete(d, m)
		} else if hcloud.IsError(err, hcloud.ErrorCodeLocked) {
			log.Printf("[INFO] Network (%v) locked, retrying in one second", network.ID)
			time.Sleep(time.Second)
			return resourceLoadBalancerNetworkDelete(d, m)
		}

		return err
	}
	if err := waitForNetworkAction(ctx, client, action, network); err != nil {
		return err
	}

	return nil
}

func setLoadBalancerNetworkSchema(d *schema.ResourceData, server *hcloud.LoadBalancer, network *hcloud.Network, serverPrivateNet *hcloud.LoadBalancerPrivateNet) {
	d.SetId(generateLoadBalancerNetworkID(server, network))
	d.Set("ip", serverPrivateNet.IP.String())
}

func generateLoadBalancerNetworkID(server *hcloud.LoadBalancer, network *hcloud.Network) string {
	return fmt.Sprintf("%d-%d", server.ID, network.ID)
}

var errInvalidLoadBalancerNetworkID = errors.New("invalid server network id")

// lookupLoadBalancerNetworkID parses the terraform load balancer network record id and return the load balancer, network and the LoadBalancerPrivateNet
//
// id format: <load balancer id>-<network id>
// Examples:
// 123-456
func lookupLoadBalancerNetworkID(ctx context.Context, terraformID string, client *hcloud.Client) (loadBalancer *hcloud.LoadBalancer, network *hcloud.Network, loadBalancerPrivateNet *hcloud.LoadBalancerPrivateNet, err error) {
	if terraformID == "" {
		err = errInvalidLoadBalancerNetworkID
		return
	}
	parts := strings.SplitN(terraformID, "-", 2)
	if len(parts) != 2 {
		err = errInvalidLoadBalancerNetworkID
		return
	}

	loadBalancerID, err := strconv.Atoi(parts[0])
	if err != nil {
		err = errInvalidLoadBalancerNetworkID
		return
	}

	loadBalancer, _, err = client.LoadBalancer.GetByID(ctx, loadBalancerID)
	if err != nil {
		err = errInvalidLoadBalancerNetworkID
		return
	}
	if loadBalancer == nil {
		err = errInvalidLoadBalancerNetworkID
		return
	}

	networkID, err := strconv.Atoi(parts[1])
	if err != nil {
		err = errInvalidLoadBalancerNetworkID
		return
	}

	network, _, err = client.Network.GetByID(ctx, networkID)
	if network == nil {
		err = errInvalidLoadBalancerNetworkID
		return
	}

	for _, pn := range loadBalancer.PrivateNet {
		if pn.Network.ID == network.ID {
			loadBalancerPrivateNet = &pn
			return
		}
	}

	err = errInvalidLoadBalancerNetworkID
	return
}

func resourceLoadBalancerNetworkUpdatePublicInterface(
	ctx context.Context, enable bool, lb *hcloud.LoadBalancer, client *hcloud.Client, d *schema.ResourceData,
) error {
	var (
		action *hcloud.Action
		err    error
	)

	if enable {
		action, _, err = client.LoadBalancer.EnablePublicInterface(ctx, lb)
	} else {
		action, _, err = client.LoadBalancer.DisablePublicInterface(ctx, lb)
	}
	if err != nil {
		return err
	}
	return waitForLoadBalancerAction(ctx, client, action, lb)
}
