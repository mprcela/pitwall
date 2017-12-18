package monit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/manifoldco/promptui"
)

func Tail(dc, service string) {
	tail()
}

func tail() error {
	url := "http://10.50.1.106:10500/services/web_app_api"
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	reader := bufio.NewReader(rsp.Body)
	l := NewLogLine()
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err
		}
		prefix := []byte("data:")
		if bytes.HasPrefix(line, prefix) {
			data := bytes.TrimPrefix(line, prefix)
			l.Print(data)
		}
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

type LogLine struct {
	sizes     map[string]int
	knownKeys []string
}

func NewLogLine() *LogLine {
	return &LogLine{
		sizes:     make(map[string]int),
		knownKeys: []string{"time", "dc", "node", "host", "app", "file", "level", "msg"},
	}
}

func (l *LogLine) Print(data []byte) {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		fmt.Print(">> " + string(data))
		return
	}

	for _, k := range l.knownKeys {
		if v, ok := m[k]; ok {
			if k == "level" {
				switch v.(string) {
				case "info":
					fmt.Printf("%v ", info(v))
				case "error":
					fmt.Printf("%v ", warn(v))
				}
			} else {
				fmt.Printf("%v ", v)
			}
		}
	}
	fmt.Printf(" ")

	var otherKeys []string
	for k, _ := range m {
		if !l.isKnownKey(k) {
			otherKeys = append(otherKeys, k)
		}
	}
	sort.Strings(otherKeys)
	for _, k := range otherKeys {
		v := m[k]
		l.printKey(k, v)
	}

	// for k, v := range m {
	// 	if !l.isKnownKey(k) {
	// 		l.printKey(k, v)
	// 	}
	// }
	fmt.Printf("\n")
}

func (l *LogLine) printKey(key string, value interface{}) {
	fmt.Printf("%s:%v ", faint(key), value)
}

func (l *LogLine) isKnownKey(k string) bool {
	for _, s := range l.knownKeys {
		if s == k {
			return true
		}
	}
	return false
}

var faint = promptui.Styler(promptui.FGFaint)
var info = promptui.Styler(promptui.FGBlue)
var success = promptui.Styler(promptui.FGGreen)
var warn = promptui.Styler(promptui.FGRed)
