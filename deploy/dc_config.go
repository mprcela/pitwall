package deploy

import (
	"fmt"
	"io/ioutil"

	"github.com/manifoldco/promptui"
	"github.com/minus5/svckit/log"
	yaml "gopkg.in/yaml.v2"
)

type DcConfig struct {
	root     string
	dc       string
	Region   string
	Dc       string
	Nomads   []string
	Services map[string]*ServiceConfig
}

func NewDcConfig(root, dc string) (*DcConfig, error) {
	c := &DcConfig{
		root: root,
		dc:   dc,
	}
	return c, c.load()
}

func (c *DcConfig) serviceNames() []string {
	var names []string
	for k, _ := range c.Services {
		names = append(names, k)
	}
	return names
}

func (c *DcConfig) Select() (string, error) {
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

func (c *DcConfig) Find(service string) *ServiceConfig {
	for k, sc := range c.Services {
		if k == service {
			return sc
		}
	}
	return nil
}

func (c *DcConfig) FileName() string {
	return fmt.Sprintf("%s/datacenters/%s/config.yml", c.root, c.dc)
}

func (c *DcConfig) load() error {
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
	log.S("from", fn).Debug("datacenter config")
	return nil
}

type ServiceConfig struct {
	Image     string
	Count     int
	DcRegion  string `yaml:"dc_region,omitempty"`
	HostGroup string `yaml:"hostgroup,omitempty"`
	Node      string `yaml:"node,omitempty"`
}

func (c *DcConfig) Save() error {
	fn := c.FileName()
	buf, err := yaml.Marshal(c)
	if err != nil {
		log.S("fn", fn).Error(err)
		return err
	}
	return ioutil.WriteFile(fn, buf, 0644)
}
