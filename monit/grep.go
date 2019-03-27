package monit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mnu5/svckit/log"
)

type GrepOptions struct {
	Address   string
	Dc        string
	Service   string
	Json      bool
	Pretty    bool
	Exclude   []string
	Include   []string
	Filter    string
	StartTime time.Time
	EndTime   time.Time
}

func (o GrepOptions) url() string {
	return fmt.Sprintf("http://%s/logs", o.Address)
}

type cwReq struct {
	Service   string     `json:"service"`
	Dc        string     `json:"dc"`
	Filter    string     `json:"filter"`
	StartTime *time.Time `json:"start_time"`
	Endtime   *time.Time `json:"end_time"`
}

func Grep(o GrepOptions) error {
	if o.Service == "" {
		services, err := getGrepServices(o)
		if err != nil {
			return err
		}
		o.Service, err = selectGrepService(services)
		if err != nil {
			return err
		}
	}

	r := cwReq{
		Service: o.Service,
		Dc:      o.Dc,
		Filter:  o.Filter,
	}
	if !o.StartTime.IsZero() {
		r.StartTime = &o.StartTime
	}
	if !o.EndTime.IsZero() {
		r.Endtime = &o.EndTime
	}

	buf, _ := json.Marshal(r)
	rsp, err := http.Post(o.url(), "application/json", bytes.NewBuffer(buf))
	if err != nil {
		log.Error(err)
		return err
	}
	logLine := NewLogLine(o.Json, o.Pretty, o.Exclude, o.Include)
	readSse(rsp.Body, func(data []byte) error {
		return logLine.Print(data)
	})
	return nil
}

func getGrepServices(o GrepOptions) ([]string, error) {
	rsp, err := http.Get(o.url())
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer rsp.Body.Close()
	buf, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var services []string
	if err := json.Unmarshal(buf, &services); err != nil {
		log.Error(err)
		return nil, err
	}
	return services, nil
}

func selectGrepService(services []string) (string, error) {
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
	return services[idx], nil
}
