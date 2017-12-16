package deploy

import (
	"fmt"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	nomadStructs "github.com/hashicorp/nomad/nomad/structs"
	"github.com/minus5/svckit/log"
)

type Deployer struct {
	root            string
	dc              string
	service         string
	image           string
	config          *DcConfig
	job             *api.Job
	cli             *api.Client
	jobModifyIndex  uint64
	jobEvalID       string
	jobDeploymentID string
}

func NewDeployer(root, dc, service, image string, config *DcConfig) *Deployer {
	return &Deployer{
		root:    root,
		dc:      dc,
		service: service,
		image:   image,
		config:  config,
	}
}

func (d *Deployer) Go() error {
	steps := []func() error{
		d.loadServiceConfig,
		d.connect,
		d.validate,
		d.plan,
		d.register,
		d.status,
	}
	return runSteps(steps)
}

func (d *Deployer) checkServiceConfig() error {
	if _, ok := d.config.Services[d.service]; !ok {
		return fmt.Errorf("service %d not found in datacenter config", d.service)
	}
	return nil
}

func (d *Deployer) plan() error {
	jp, _, err := d.cli.Jobs().Plan(d.job, false, nil)
	if err != nil {
		return err
	}
	d.jobModifyIndex = jp.JobModifyIndex
	log.I("modifyIndex", int(jp.JobModifyIndex)).Info("job planned")
	return nil
}

func (d *Deployer) register() error {
	jr, _, err := d.cli.Jobs().EnforceRegister(d.job, d.jobModifyIndex, nil)
	if err != nil {
		return err
	}
	d.jobEvalID = jr.EvalID
	if err := d.getDeploymentID(); err != nil {
		return err
	}
	log.S("evalID", jr.EvalID).S("deploymentID", d.jobDeploymentID).Info("job registered")
	return nil
}

func (d *Deployer) getDeploymentID() error {
	for {
		ev, _, err := d.cli.Evaluations().Info(d.jobEvalID, nil)
		if err != nil {
			return err
		}
		if ev.DeploymentID != "" {
			d.jobDeploymentID = ev.DeploymentID
			return nil
		}
		time.Sleep(time.Second)
	}
}

func (d *Deployer) status() error {
	depID := d.jobDeploymentID
	t := time.Now()
	q := &api.QueryOptions{WaitIndex: 1, AllowStale: true, WaitTime: time.Duration(5 * time.Second)}
	for {
		dep, meta, err := d.cli.Deployments().Info(depID, q)
		if err != nil {
			return err
		}
		q.WaitIndex = meta.LastIndex
		du := fmt.Sprintf("%.2fs", time.Since(t).Seconds())
		if dep.Status == nomadStructs.DeploymentStatusRunning {
			log.S("running", du).Debug("checking status")
			continue
		}
		if dep.Status == nomadStructs.DeploymentStatusSuccessful {
			log.S("after", du).Info("deployment successful")
			break
		}
		return fmt.Errorf("deployment failed status: %s", dep.Status)
	}
	return nil
}

func (d *Deployer) loadServiceConfig() error {
	fn := fmt.Sprintf("%s/nomad/service/%s.nomad", d.root, d.service)
	job, err := jobspec.ParseFile(fn)
	if err != nil {
		return err
	}
	log.S("from", fn).Debug("loaded config")
	d.job = job
	return d.checkServiceConfig()
}

func (d *Deployer) connect() error {
	if len(d.config.Nomads) == 0 {
		return fmt.Errorf("no nomad servers configured")
	}
	c := &api.Config{}
	addr := d.config.Nomads[0]
	c = c.ClientConfig(d.config.Dc, addr, false)
	cli, err := api.NewClient(c)
	if err != nil {
		return err
	}
	log.S("nomad", addr).Info("connected")
	d.cli = cli
	return nil
}

func (d *Deployer) validate() error {
	d.job.Region = &d.config.Region
	d.job.AddDatacenter(d.config.Dc)

	s := d.config.Services[d.service]
	if s.DcRegion != "" {
		d.job.Constrain(api.NewConstraint("${meta.dc_region}", "=", s.DcRegion))
	}
	if s.HostGroup != "" {
		d.job.Constrain(api.NewConstraint("${meta.hostgroup}", "=", s.HostGroup))
	}
	if s.Node != "" {
		d.job.Constrain(api.NewConstraint("${meta.node}", "=", s.Node))
	}

	for _, tg := range d.job.TaskGroups {
		if *tg.Name == d.service {
			if s.Count > 0 {
				tg.Count = &s.Count
			}
			for _, ta := range tg.Tasks {
				if ta.Name == d.service {
					ta.Config["image"] = d.image
					s.Image = d.image
				}
			}
		}
	}

	_, _, err := d.cli.Jobs().Validate(d.job, nil)
	if err != nil {
		return err
	}
	log.Info("job validated")
	return nil
}
