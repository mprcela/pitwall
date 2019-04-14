package service

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/minus5/svckit/log"
)

func NewFinder(consulAddr string) Finder {
	return Finder{consulAddr: consulAddr}
}

type Finder struct {
	consulAddr string
}

func (f Finder) consul() *api.Client {
	config := api.DefaultConfig()
	config.Address = f.consulAddr
	cli, err := api.NewClient(config)
	if err != nil {
		log.S("addr", f.consulAddr).Fatal(err)
	}
	return cli
}

func (f Finder) All(pattern string) Services {
	cli := f.consul()

	services := make(map[string][]Service)

	svcNames, _, err := cli.Catalog().Services(&api.QueryOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for name := range svcNames {
		if pattern != "" {
			if !strings.Contains(name, pattern) {
				continue
			}
		}
		svcs, _, err := cli.Catalog().Service(name, "", &api.QueryOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, svc := range svcs {
			addr := svc.ServiceAddress
			if addr == "" {
				addr = svc.Address
			}
			services[name] = append(services[name], Service{Name: name, Address: addr, Port: svc.ServicePort, Node: svc.Node})
		}
	}

	return services
}

func hasDebugInterface(addr string, port int) bool {
	url := fmt.Sprintf("http://%s:%d/debug/vars", addr, port)
	client := http.Client{
		Timeout: time.Duration(time.Second),
	}
	rsp, err := client.Get(url)
	if err != nil {
		return false
	}
	if rsp.StatusCode != 200 {
		return false
	}
	return true
}

type Services map[string][]Service

func (s Services) Names() []string {
	var n []string
	for k := range s {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}

type Service struct {
	Name    string
	Node    string
	Address string
	Port    int
	Sys     uint64
	Alloc   uint64
}

func (s Service) String() string {
	return fmt.Sprintf("%s:%d %s", s.Address, s.Port, s.Node)
}

func (f Finder) isNonGoService(name string) bool {
	nonGoServices := []string{"espresso", "maloprodaja", "corner", "web-backend", "kladomat", "kladomat-ng", "statsd", "minfin", "ponuda-server", "version-manager", "web", "izvjestaji", "lotator", "upis-validator", "sbk-api-ws", "bonovi-partner-api", "chat-api-ws", "bonovi-s2-api", "grafana-to-git", "chat-ws", "sbk-api-sse"}
	for _, n := range nonGoServices {
		if n == name {
			return true
		}
	}
	return false
}
