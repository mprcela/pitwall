package deploy

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/manifoldco/promptui"
	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
)

// TODO
// prikazi koji je trenutni image
// povezati s deploy-erom

// Run deployment process
func Run(deployment, service, path, registry, image string, noGit bool, consul string, dryRun bool) {
	l := newTerminalLogger()
	defer l.Close()
	w := Worker{
		service:     service,
		root:        env.ExpandPath(path),
		registryURL: registry,
		deployment:  deployment,
		image:       image,
		noGit:       noGit,
		consul:      consul,
		dryRun:      dryRun,
	}

	if err := w.Go(); err != nil {
		log.Error(err)
	} else {
		fmt.Printf("%s %s\n", promptui.IconGood, success("done"))
	}
}

// Worker structure for deployment
type Worker struct {
	root        string
	registryURL string
	deployment  string
	service     string
	image       string
	consul      string
	consulDc    string
	noGit       bool
	dryRun      bool

	depConfig     *DeploymentConfig
	serviceConfig *ServiceConfig
	repo          Repo
	deployer      *Deployer
}

// Go starts deployment process
func (w *Worker) Go() error {
	steps := []func() error{
		w.pull,
		w.selectService,
		w.selectImage,
		//w.confirmSelection,
		w.deploy,
		w.pullChanges,
		w.updateDepConfig,
		w.push,
	}
	return runSteps(steps)
}

func runSteps(steps []func() error) error {
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) deploy() error {
	dcs := w.depConfig.FindDatacenters(w.service)
	if len(dcs) == 0 {
		log.Fatal(fmt.Errorf("datacenters for service %s not set", w.service))
	}
	for _, dc := range dcs {
		log.Info("Deploying service %s to dacenter %s", w.service, dc)
		// temporary fix until switch is made
		nomadName := "nomad"
		ndc := dc // datacenter used to query nomad from consul
		if ndc == "js" {
			ndc = "s2"
			nomadName = "nomad-js"
		}
		address := w.getServiceAddressByTag("http", nomadName, ndc)
		d := NewDeployer(w.root, w.service, w.image, w.depConfig, address, dc, w.deployment)
		w.deployer = d
		if err := d.Go(w.dryRun); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) pull() error {
	if w.noGit {
		return nil
	}
	gitURL := "git@github.com:minus5/infrastructure.git"
	repo, err := NewRepo(w.root, gitURL)
	if err != nil {
		return err
	}
	w.repo = repo
	return nil
}

func (w *Worker) pullChanges() error {
	if w.noGit {
		return nil
	}
	return w.repo.Pull()
}

func (w *Worker) push() error {
	if w.noGit {
		return nil
	}
	return w.repo.Commit(fmt.Sprintf("deployed %s to %s", w.service, w.deployment), w.depConfig.FileName())
}

func (w *Worker) selectService() error {
	c, err := NewDeploymentConfig(w.root, w.deployment)
	if err != nil {
		return err
	}
	w.depConfig = c
	if w.service == "" {
		s, err := c.Select()
		if err != nil {
			return err
		}
		w.service = s
		log.S("service", w.service).Info("service selected")
	}
	svc := c.Find(w.service)
	if svc == nil {
		return fmt.Errorf("service %s not found", w.service)
	}
	w.serviceConfig = svc
	return nil
}

func (w *Worker) selectImage() error {
	if w.image != "" {
		log.S("image", w.image).Info("image preselected with flag")
		return nil
	}

	i, err := NewImage(w.registryURL, w.service, w.serviceConfig.Image)
	if err != nil {
		log.Error(err)
		return err
	}
	image, err := i.Select()
	if err != nil {
		return err
	}
	w.image = image
	log.S("image", image).Info("image selected")
	return nil
}

func (w *Worker) confirmSelection() error {
	prompt := promptui.Prompt{
		Label:   "Continue? ",
		Default: "y",
		//IsConfirm: true,
		Validate: func(input string) error {
			if input == "" || input == "y" || input == "n" {
				return nil
			}
			return fmt.Errorf("y/n")
		},
	}
	res, err := prompt.Run()
	if err != nil || res == "n" {
		return fmt.Errorf("aborted")
	}
	return nil
}

func (w *Worker) updateDepConfig() error {
	return w.depConfig.Save()
}

type terminalLogger struct {
	f *os.File
}

func newTerminalLogger() *terminalLogger {
	fn := fmt.Sprintf("%s%s.log", os.TempDir(), env.AppName())
	f, err := os.Create(fn)
	if err != nil {
		log.Fatal(err)
	}
	tl := &terminalLogger{f: f}
	log.SetOutput(tl)
	log.S("path", fn).Debug("logging to")
	return tl
}

var faint = promptui.Styler(promptui.FGFaint)
var info = promptui.Styler(promptui.FGBlue)
var success = promptui.Styler(promptui.FGGreen)
var warn = promptui.Styler(promptui.FGRed)
var lastMsg = ""

func (l terminalLogger) Write(p []byte) (int, error) {
	var m map[string]interface{}
	json.Unmarshal(p, &m)
	switch m["level"] {
	case "error", "fatal":
		if m := m["msg"].(string); m != lastMsg {
			fmt.Printf("%s ", promptui.IconBad)
			fmt.Printf("%s", warn(m))
			lastMsg = m
		} else {
			return len(p), nil
		}
	case "info":
		fmt.Printf("%s", info(m["msg"]))
	case "debug":
		fmt.Printf("%s", faint(m["msg"]))
	}

	var keys []string
	for k, _ := range m {
		if k == "file" || k == "host" || k == "time" || k == "app" || k == "msg" || k == "level" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m[k]
		fmt.Printf(faint(fmt.Sprintf(" %s: %v", k, v)))
	}
	fmt.Printf("\n")
	l.f.Write(p)
	return len(p), nil
}

func (terminalLogger) WriteString(s string) (int, error) {
	return len(s), nil
}

// Close logging to local file
func (l *terminalLogger) Close() {
	l.f.Close()
}

func (w *Worker) getServiceAddressByTag(tag, name, dc string) string {
	if err := dcy.ConnectTo(w.consul); err != nil {
		log.Fatal(err)
	}
	addr, err := dcy.ServiceInDcByTag(tag, name, dc)
	if err == nil {
		return addr.String()
	}
	log.Fatal(fmt.Errorf("service %s with tag %s not found in consul %s", name, tag, w.consul))
	return ""
}
