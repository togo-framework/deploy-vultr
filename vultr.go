// Package vultr is a Vultr deploy driver for togo: provisions a cloud instance
// (Docker via cloud-init) running the app image. Select with
// deploy.provider=vultr; needs VULTR_API_KEY.
package vultr

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/togo-framework/deploy"
	"github.com/togo-framework/togo"
	"github.com/vultr/govultr/v3"
	"golang.org/x/oauth2"
)

func init() { deploy.RegisterDriver("vultr", New) }

func New(_ *togo.Kernel) (deploy.Deployer, error) {
	key := os.Getenv("VULTR_API_KEY")
	if key == "" {
		return nil, errors.New("deploy-vultr: VULTR_API_KEY not set")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: key})
	c := govultr.NewClient(oauth2.NewClient(context.Background(), ts))
	return &driver{c: c}, nil
}

type driver struct{ c *govultr.Client }

func cloudInit(image string) string {
	return fmt.Sprintf("#cloud-config\nruncmd:\n  - curl -fsSL https://get.docker.com | sh\n  - docker run -d --name app --restart always -p 80:8080 %s\n", image)
}

func (d *driver) byLabel(ctx context.Context, label string) (*govultr.Instance, error) {
	list, _, _, err := d.c.Instance.List(ctx, &govultr.ListOptions{PerPage: 200})
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].Label == label {
			return &list[i], nil
		}
	}
	return nil, nil
}

func (d *driver) Provision(ctx context.Context, spec deploy.Spec) (*deploy.Result, error) {
	region := spec.Region
	if region == "" {
		region = "fra"
	}
	plan := "vc2-1c-1gb"
	if v, ok := spec.Options["plan"].(string); ok && v != "" {
		plan = v
	}
	// OsID 1743 = Ubuntu 22.04 LTS x64
	inst, _, err := d.c.Instance.Create(ctx, &govultr.InstanceCreateReq{
		Region:   region,
		Plan:     plan,
		OsID:     1743,
		Label:    spec.App,
		Hostname: spec.App,
		UserData: base64.StdEncoding.EncodeToString([]byte(cloudInit(spec.Image))),
	})
	if err != nil {
		return nil, fmt.Errorf("vultr provision: %w", err)
	}
	return &deploy.Result{URL: "http://" + inst.MainIP, Message: "instance creating; app boots via cloud-init", Raw: map[string]any{"id": inst.ID}}, nil
}

func (d *driver) Deploy(ctx context.Context, spec deploy.Spec) (*deploy.Result, error) {
	inst, err := d.byLabel(ctx, spec.App)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return d.Provision(ctx, spec)
	}
	return &deploy.Result{URL: "http://" + inst.MainIP, Message: "instance up; redeploy the container via CI/SSH", Raw: map[string]any{"id": inst.ID}}, nil
}

func (d *driver) Destroy(ctx context.Context, spec deploy.Spec) error {
	inst, err := d.byLabel(ctx, spec.App)
	if err != nil || inst == nil {
		return err
	}
	return d.c.Instance.Delete(ctx, inst.ID)
}

func (d *driver) Status(ctx context.Context, spec deploy.Spec) (*deploy.Status, error) {
	inst, err := d.byLabel(ctx, spec.App)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		return &deploy.Status{Healthy: false, Detail: "no instance"}, nil
	}
	return &deploy.Status{Healthy: inst.Status == "active" && inst.PowerStatus == "running", Detail: inst.Status, Raw: map[string]any{"ip": inst.MainIP}}, nil
}
