package deploy

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
)

// TODO
// prikazi koji je trenutni image
// povezati s deploy-erom

func Run(dc, service, path, registry string, noGit bool, address string) {
	l := newTerminalLogger()
	defer l.Close()

	w := Worker{
		service:     service,
		root:        env.ExpandPath(path),
		registryUrl: registry,
		dc:          dc,
		noGit:       noGit,
		address:     address,
	}
	if err := w.Go(); err != nil {
		log.Error(err)
	} else {
		fmt.Printf("%s %s\n", promptui.IconGood, success("done"))
	}
}

type Worker struct {
	root        string
	registryUrl string
	dc          string
	service     string
	image       string
	address     string
	noGit       bool

	dcConfig      *DcConfig
	serviceConfig *ServiceConfig
	repo          Repo
	deployer      *Deployer
}

func (w *Worker) Go() error {
	steps := []func() error{
		w.pull,
		w.selectService,
		w.selectImage,
		//w.confirmSelection,
		w.deploy,
		w.updateDcConfig,
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
	d := NewDeployer(w.root, w.dc, w.service, w.image, w.dcConfig, w.address)
	w.deployer = d
	return d.Go()
}

func (w *Worker) pull() error {
	if w.noGit {
		return nil
	}
	gitUrl := "git@github.com:minus5/infrastructure.git"
	repo, err := NewRepo(w.root, gitUrl)
	if err != nil {
		return err
	}
	w.repo = repo
	return nil
}

func (w *Worker) push() error {
	if w.noGit {
		return nil
	}
	return w.repo.Commit(fmt.Sprintf("deployed %s to %s", w.service, w.dc), w.dcConfig.FileName())
}

func (w *Worker) selectService() error {
	c, err := NewDcConfig(w.root, w.dc)
	if err != nil {
		return err
	}
	w.dcConfig = c
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
	i, err := NewImage(w.registryUrl, w.service, w.serviceConfig.Image)
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

func (w *Worker) updateDcConfig() error {
	return w.dcConfig.Save()
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

	for k, v := range m {
		if k == "file" || k == "host" || k == "time" || k == "app" || k == "msg" || k == "level" {
			continue
		}
		fmt.Printf(faint(fmt.Sprintf(" %s: %v", k, v)))
	}
	fmt.Printf("\n")
	l.f.Write(p)
	return len(p), nil
}

func (terminalLogger) WriteString(s string) (int, error) {
	return len(s), nil
}

func (t *terminalLogger) Close() {
	t.f.Close()
}
