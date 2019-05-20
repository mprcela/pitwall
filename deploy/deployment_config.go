package deploy

import (
	"fmt"
	"io/ioutil"

	"github.com/manifoldco/promptui"
	"github.com/minus5/svckit/log"
	yaml "gopkg.in/yaml.v2"
)

// DeploymentConfig containes parameters for specific deployment
type DeploymentConfig struct {
	root        string
	deployment  string
	Datacenters map[string]*DcConfig
}

// DcConfig contains parameters for specific datacenter
type DcConfig struct {
	Services map[string]*ServiceConfig
}

// NewDeploymentConfig creates new config for specific deployment
func NewDeploymentConfig(root, deployment string) (*DeploymentConfig, error) {
	c := &DeploymentConfig{
		root:       root,
		deployment: deployment,
	}
	return c, c.load()
}

func (c *DeploymentConfig) serviceNames() []string {
	var names []string
	for _, s := range c.Datacenters {
		for k := range s.Services {
			names = append(names, k)
		}
	}
	return names
}

// Select service
func (c *DeploymentConfig) Select() (string, error) {
	names := c.serviceNames()
	prompt := promptui.Select{
		Label: "Select service",
		Items: names,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Selected: string([]byte("\033[" + "1A")),
		},
	}
	idx, _, err := prompt.Run()
	return names[idx], err
}

// Find returns config for specific service
func (c *DeploymentConfig) Find(service string) *ServiceConfig {
	for _, s := range c.Datacenters {
		for k, sc := range s.Services {
			if k == service {
				return sc
			}
		}
	}
	return nil
}

// FindDatacenter finds datacenter for service if it exists
func (c *DeploymentConfig) FindDatacenter(service string) string {
	for d, s := range c.Datacenters {
		for k := range s.Services {
			if k == service {
				return d
			}
		}
	}
	return ""
}

// FileName returns config.yml for dc
func (c *DeploymentConfig) FileName() string {
	return fmt.Sprintf("%s/deployments/%s/config.yml", c.root, c.deployment)
}

func (c *DeploymentConfig) load() error {
	fn := c.FileName()
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Error(err)
		return err
	}
	if err := yaml.Unmarshal([]byte(data), c); err != nil {
		log.Error(err)
		return err
	}
	log.S("from", fn).Debug("deployment config")
	return nil
}

// ServiceConfig represent structure for config.yml
type ServiceConfig struct {
	Image  string
	Count  int
	Node   string `yaml:"node,omitempty"`
	CPU    int    `yaml:"cpu,omitempty"`
	Memory int    `jamyl:"mem,omitempty"`
}

// Save changes to config.yml
func (c *DeploymentConfig) Save() error {
	fn := c.FileName()
	buf, err := yaml.Marshal(c)
	if err != nil {
		log.S("fn", fn).Error(err)
		return err
	}
	return ioutil.WriteFile(fn, buf, 0644)
}
