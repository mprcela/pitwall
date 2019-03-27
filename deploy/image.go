package deploy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/manifoldco/promptui"
	"github.com/mnu5/svckit/log"
)

// Image represents docker image
type Image struct {
	registryURL string
	service     string
	current     string
	tags        []Tag
}

// NewImage sets image for deploy
func NewImage(registry, service, current string) (*Image, error) {
	i := &Image{
		registryURL: registry,
		service:     service,
		current:     current,
	}
	return i, i.findTags()
}

// Select service image
func (i Image) Select() (string, error) {
	prompt := promptui.Select{
		Items: i.tags,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Selected: string([]byte("\033[" + "1A")),
			Label:    fmt.Sprintf(`{{ "Select image:" }} {{ "(* current)" | faint }}`),
		},
	}

	idx, _, err := prompt.Run()
	t := i.tags[idx]
	return fmt.Sprintf("%s/%s:%s", i.registryURL, i.service, t.tag), err
}

func (i *Image) findTags() error {
	resp, err := http.Get(fmt.Sprintf("http://%s/v2/%s/tags/list", i.registryURL, i.service))
	if err != nil {
		log.Error(err)
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return err
	}
	data := &struct {
		Name string
		Tags []string
	}{}
	err = json.Unmarshal(body, data)
	if err != nil {
		log.Error(err)
		return err
	}
	var s tags
	for _, t := range data.Tags {
		s = append(s, NewTag(t, strings.Contains(i.current, t)))
	}
	sort.Sort(s)
	i.tags = s
	return nil
}

// Tag is docker image tag
type Tag struct {
	tag     string
	created time.Time
	current bool
}

func (t Tag) String() string {
	prefix := " "
	if t.current {
		prefix = "*"
	}
	if t.created.IsZero() {
		return fmt.Sprintf("%s %-25s %-15s %s",
			prefix,
			"",
			"",
			t.tag)
	}

	h := units.HumanDuration(time.Now().UTC().Sub(t.created))
	h = h + " ago"

	return fmt.Sprintf("%s %-25s %-15s %s",
		prefix,
		h,
		t.created.Format("02.01. 15:04"),
		t.tag)
}

// NewTag is new docker image tag
func NewTag(t string, current bool) Tag {
	parts := strings.Split(t, ".")
	c, _ := time.ParseInLocation("20060102150405", parts[0], time.Local)
	return Tag{
		tag:     t,
		created: c,
		current: current,
	}
}

type tags []Tag

func (a tags) Len() int {
	return len(a)
}

func (a tags) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a tags) Less(i, j int) bool {
	t1 := a[i]
	t2 := a[j]
	if t1.created.IsZero() && !t2.created.IsZero() {
		return false
	}
	if !t1.created.IsZero() && t2.created.IsZero() {
		return true
	}
	if t1.created.IsZero() && t2.created.IsZero() {
		return t1.tag > t2.tag
	}
	return t1.created.After(t2.created)
}
