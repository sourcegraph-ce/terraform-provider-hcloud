package hcloud

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	log "github.com/sourcegraph-ce/logrus"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hetznercloud/hcloud-go/hcloud"
)

func resourceServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceServerCreate,
		Read:   resourceServerRead,
		Update: resourceServerUpdate,
		Delete: resourceServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"server_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(val interface{}, key string) (i []string, errors []error) {
					image := val.(string)
					if len(image) == 0 {
						errors = append(errors, fmt.Errorf("%q must have more then 0 characters. Have you set the name instead of an ID?", key))
					}
					return
				},
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"datacenter": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"user_data": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: userDataDiffSuppress,
				StateFunc: func(v interface{}) string {
					switch v.(type) {
					case string:
						return userDataHashSum(v.(string))
					default:
						return ""
					}
				},
			},
			"ssh_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
			},
			"keep_disk": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"backup_window": {
				Type:       schema.TypeString,
				Deprecated: "You should remove this property from your terraform configuration.",
				Computed:   true,
			},
			"backups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ipv4_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iso": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rescue": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func userDataHashSum(userData string) string {
	sum := sha1.Sum([]byte(userData))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func userDataDiffSuppress(k, old, new string, d *schema.ResourceData) bool {
	userData := d.Get(k).(string)
	if new != "" && userData != "" {
		if _, err := base64.StdEncoding.DecodeString(old); err != nil {
			return userDataHashSum(old) == new
		}
	}
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}

func resourceServerCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	var err error
	image, _, err := client.Image.Get(ctx, d.Get("image").(string))
	if err != nil {
		return err
	}

	opts := hcloud.ServerCreateOpts{
		Name: d.Get("name").(string),
		ServerType: &hcloud.ServerType{
			Name: d.Get("server_type").(string),
		},
		Image:    image,
		UserData: d.Get("user_data").(string),
	}

	opts.SSHKeys, err = getSSHkeys(ctx, client, d)
	if err != nil {
		return err
	}

	if datacenter, ok := d.GetOk("datacenter"); ok {
		opts.Datacenter = &hcloud.Datacenter{Name: datacenter.(string)}
	}

	if location, ok := d.GetOk("location"); ok {
		opts.Location = &hcloud.Location{Name: location.(string)}
	}
	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	res, _, err := client.Server.Create(ctx, opts)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(res.Server.ID))
	if err := waitForServerAction(ctx, client, res.Action, res.Server); err != nil {
		return err
	}
	for _, nextAction := range res.NextActions {
		if err := waitForServerAction(ctx, client, nextAction, res.Server); err != nil {
			return err
		}

	}

	backups := d.Get("backups").(bool)
	if err := setBackups(ctx, client, res.Server, backups); err != nil {
		return err
	}

	if iso, ok := d.GetOk("iso"); ok {
		if err := setISO(ctx, client, res.Server, iso.(string)); err != nil {
			return err
		}
	}

	if rescue, ok := d.GetOk("rescue"); ok {
		if err := setRescue(ctx, client, res.Server, rescue.(string), opts.SSHKeys, 0); err != nil {
			return err
		}
	}

	return resourceServerRead(d, m)
}

func resourceServerRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	server, _, err := client.Server.Get(ctx, d.Id())
	if err != nil {
		if resourceServerIsNotFound(err, d) {
			return nil
		}
		return err
	}
	if server == nil {
		d.SetId("")
		return nil
	}
	setServerSchema(d, server)

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": server.PublicNet.IPv4.IP.String(),
	})

	return nil
}

func resourceServerUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	server, _, err := client.Server.Get(ctx, d.Id())
	if err != nil {
		return err
	}
	if server == nil {
		d.SetId("")
		return nil
	}

	d.Partial(true)
	if d.HasChange("name") {
		newName := d.Get("name")
		_, _, err := client.Server.Update(ctx, server, hcloud.ServerUpdateOpts{
			Name: newName.(string),
		})
		if err != nil {
			if resourceServerIsNotFound(err, d) {
				return nil
			}
			return err
		}
		d.SetPartial("name")
	}
	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := client.Server.Update(ctx, server, hcloud.ServerUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceServerIsNotFound(err, d) {
				return nil
			}
			return err
		}
		d.SetPartial("labels")
	}
	if d.HasChange("server_type") {
		serverType := d.Get("server_type").(string)
		keepDisk := d.Get("keep_disk").(bool)

		if server.Status == hcloud.ServerStatusRunning {
			action, _, err := client.Server.Poweroff(ctx, server)
			if err != nil {
				return err
			}
			if err := waitForServerAction(ctx, client, action, server); err != nil {
				return err
			}
		}

		action, _, err := client.Server.ChangeType(ctx, server, hcloud.ServerChangeTypeOpts{
			ServerType:  &hcloud.ServerType{Name: serverType},
			UpgradeDisk: !keepDisk,
		})
		if err != nil {
			return err
		}
		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
		d.SetPartial("server_type")
	}

	if d.HasChange("backups") {
		backups := d.Get("backups").(bool)
		if err := setBackups(ctx, client, server, backups); err != nil {
			return err
		}
		d.SetPartial("backups")
	}

	if d.HasChange("iso") {
		iso := d.Get("iso").(string)
		if err := setISO(ctx, client, server, iso); err != nil {
			return err
		}
		d.SetPartial("iso")
	}

	if d.HasChange("rescue") {
		rescue := d.Get("rescue").(string)
		sshKeys, err := getSSHkeys(ctx, client, d)
		if err != nil {
			return err
		}
		if err := setRescue(ctx, client, server, rescue, sshKeys, 0); err != nil {
			return err
		}
	}

	d.Partial(false)
	return resourceServerRead(d, m)
}

func resourceServerDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*hcloud.Client)
	ctx := context.Background()

	serverID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid server id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.Server.Delete(ctx, &hcloud.Server{ID: serverID}); err != nil {
		return err
	}

	return nil
}

func resourceServerIsNotFound(err error, d *schema.ResourceData) bool {
	if hcerr, ok := err.(hcloud.Error); ok && hcerr.Code == hcloud.ErrorCodeNotFound {
		log.Printf("[WARN] Server (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setBackups(ctx context.Context, client *hcloud.Client, server *hcloud.Server, backups bool) error {
	if server.BackupWindow != "" && !backups {
		action, _, err := client.Server.DisableBackup(ctx, server)
		if err != nil {
			return err
		}
		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
		return nil
	}
	if server.BackupWindow == "" && backups {
		action, _, err := client.Server.EnableBackup(ctx, server, "")
		if err != nil {
			return err
		}
		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
	}
	return nil
}

func setISO(ctx context.Context, client *hcloud.Client, server *hcloud.Server, isoIDOrName string) error {
	isoChange := false
	if server.ISO != nil {
		isoChange = true
		action, _, err := client.Server.DetachISO(ctx, server)
		if err != nil {
			return err
		}
		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
	}
	if isoIDOrName != "" {
		isoChange = true

		iso, _, err := client.ISO.Get(ctx, isoIDOrName)
		if err != nil {
			return err
		}

		if iso == nil {
			return fmt.Errorf("ISO not found: %s", isoIDOrName)
		}

		action, _, err := client.Server.AttachISO(ctx, server, iso)
		if err != nil {
			return err
		}
		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
	}

	if isoChange {
		action, _, err := client.Server.Reset(ctx, server)
		if err != nil {
			return err
		}
		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
	}
	return nil
}

func setRescue(ctx context.Context, client *hcloud.Client, server *hcloud.Server, rescue string, sshKeys []*hcloud.SSHKey, retry int) error {
	rescueChanged := false
	if server.RescueEnabled {
		rescueChanged = true
		action, _, err := client.Server.DisableRescue(ctx, server)
		if err != nil {
			return err
		}
		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
	}
	if rescue != "" {
		rescueChanged = true
		res, _, err := client.Server.EnableRescue(ctx, server, hcloud.ServerEnableRescueOpts{
			Type:    hcloud.ServerRescueType(rescue),
			SSHKeys: sshKeys,
		})
		if err != nil {
			return err
		}
		if err := waitForServerAction(ctx, client, res.Action, server); err != nil {
			if retry < 5 {
				log.Printf("[INFO] server (%d) action enable_rescue failed, retrying...", server.ID)
				time.Sleep(time.Duration(1+retry) * time.Second)
				retry = retry + 1
				return setRescue(ctx, client, server, rescue, sshKeys, retry)
			}
			return err
		}
	}
	if rescueChanged {
		action, _, err := client.Server.Reset(ctx, server)
		if err != nil {
			return err
		}
		if err := waitForServerAction(ctx, client, action, server); err != nil {
			return err
		}
	}
	return nil
}

func getSSHkeys(ctx context.Context, client *hcloud.Client, d *schema.ResourceData) (sshKeys []*hcloud.SSHKey, err error) {
	for _, sshKeyValue := range d.Get("ssh_keys").([]interface{}) {
		sshKeyIDOrName := sshKeyValue.(string)
		var sshKey *hcloud.SSHKey
		sshKey, _, err = client.SSHKey.Get(ctx, sshKeyIDOrName)
		if err != nil {
			return
		}
		if sshKey == nil {
			err = fmt.Errorf("SSH key not found: %s", sshKeyIDOrName)
			return
		}
		sshKeys = append(sshKeys, sshKey)
	}
	return
}

func waitForServerAction(ctx context.Context, client *hcloud.Client, action *hcloud.Action, server *hcloud.Server) error {
	log.Printf("[INFO] server (%d) waiting for %q action to complete...", server.ID, action.Command)
	_, errCh := client.Action.WatchProgress(ctx, action)
	if err := <-errCh; err != nil {
		return err
	}
	log.Printf("[INFO] server (%d) %q action succeeded", server.ID, action.Command)
	return nil
}

func setServerSchema(d *schema.ResourceData, s *hcloud.Server) {
	d.SetId(strconv.Itoa(s.ID))
	d.Set("name", s.Name)
	d.Set("datacenter", s.Datacenter.Name)
	d.Set("location", s.Datacenter.Location.Name)
	d.Set("status", s.Status)
	d.Set("server_type", s.ServerType.Name)
	d.Set("ipv4_address", s.PublicNet.IPv4.IP.String())
	d.Set("ipv6_address", s.PublicNet.IPv6.IP.String()+"1")
	d.Set("ipv6_network", s.PublicNet.IPv6.Network.String())
	d.Set("backup_window", s.BackupWindow)
	d.Set("backups", s.BackupWindow != "")
	d.Set("labels", s.Labels)
	if s.Image != nil {
		if s.Image.Name != "" {
			d.Set("image", s.Image.Name)
		} else {
			d.Set("image", s.Image.ID)
		}
	}
}
