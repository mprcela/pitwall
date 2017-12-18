package monit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	units "github.com/docker/go-units"
	"github.com/manifoldco/promptui"
)

type TailOptions struct {
	Address string
	Service string
	Json    bool
	Pretty  bool
}

func (o TailOptions) servicesUrl() string {
	return fmt.Sprintf("http://%s/services", o.Address)
}

func (o TailOptions) logsUrl() string {
	return fmt.Sprintf("http://%s/services/%s", o.Address, o.Service)
}

func Tail(o TailOptions) {
	if o.Service == "" {
		services, err := getServices(o)
		if err != nil {
			return
		}
		o.Service, err = selectService(services)
		if err != nil {
			return
		}
	}
	tail(o)
}

type service struct {
	Service  string
	ActiveAt time.Time `json:"active_at"`
}

func (s service) String() string {
	h := units.HumanDuration(time.Now().UTC().Sub(s.ActiveAt))
	h = h + " ago"
	return fmt.Sprintf("%-30s %s", s.Service, h)
}

func selectService(services []service) (string, error) {
	prompt := promptui.Select{
		Label: "Select service:",
		Items: services,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Selected: string([]byte("\033[" + "1A")),
		},
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return services[idx].Service, nil
}

func getServices(o TailOptions) ([]service, error) {
	rsp, err := http.Get(o.servicesUrl())
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	var services []service
	if err := json.Unmarshal(buf, &services); err != nil {
		return nil, err
	}
	return services, nil
}

var dataLinePrefix = []byte("data: ")
var heartbeatLinepPrefix = []byte("event: heartbeat")

func tail(o TailOptions) error {
	rsp, err := http.Get(o.logsUrl())
	if err != nil {
		return err
	}
	logLine := NewLogLine(o.Json, o.Pretty)
	readSse(rsp.Body, func(data []byte) error {
		return logLine.Print(data)
	})
	return nil
}

func readSse(body io.ReadCloser, lineHanlder func([]byte) error) error {
	defer body.Close()

	reader := bufio.NewReader(body)
	heartbeatLine := false
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		if len(line) == 0 {
			continue
		}
		if bytes.HasPrefix(line, dataLinePrefix) && !heartbeatLine {
			data := bytes.TrimPrefix(line, dataLinePrefix)
			lineHanlder(data)
		}
		heartbeatLine = bytes.HasPrefix(line, heartbeatLinepPrefix)
	}
	return nil
}

func pp(o interface{}) {
	buf, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", buf)
}
