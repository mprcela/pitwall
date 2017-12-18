package monit

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

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

func formatTime(t time.Time) string {
	cp := time.Now().Add(-24 * time.Hour)
	if t.Before(cp) {
		return t.Format("02.01.2006 15:04:05.999")
	}
	return t.Format("15:04:05.999")
}

func (l *LogLine) Print(data []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	for _, k := range l.knownKeys {
		if v, ok := m[k]; ok {
			switch k {
			case "level":
				switch v.(string) {
				case "info":
					l.print(k, info(v), false)
				case "error":
					l.print(k, warn(v), false)
				}
			case "time":
				t, err := time.Parse("2006-01-02T15:04:05.999999-07:00", v.(string))
				if err == nil {
					l.print(k, formatTime(t), false)
				}
			default:
				l.print(k, v, false)
			}
		}
	}

	var otherKeys []string
	for k, _ := range m {
		if !l.isKnownKey(k) {
			otherKeys = append(otherKeys, k)
		}
	}
	sort.Strings(otherKeys)
	for _, k := range otherKeys {
		l.print(k, m[k], true)
	}
	fmt.Printf("\n")

	return nil
}

var maxValueSize = 50

func (l *LogLine) print(key string, value interface{}, printKey bool) {
	strValue := fmt.Sprintf("%v", value)
	if m, ok := value.(map[string]interface{}); ok {
		if buf, err := json.Marshal(m); err == nil {
			strValue = string(buf)
		}
	}
	if len(strValue) == 0 {
		return
	}
	strValue = l.formatSpaces(key, strValue)
	if !printKey {
		fmt.Printf("%v ", strValue)
		return
	}
	fmt.Printf("%s%s%v ", faint(key), faint(":"), strValue)
}

func (l *LogLine) formatSpaces(key, value string) string {
	valueLen := len(value)
	current, ok := l.sizes[key]
	if !ok {
		if valueLen < maxValueSize {
			l.sizes[key] = valueLen
		}
		return value
	}
	if valueLen < current {
		return value + strings.Repeat(" ", current-valueLen)

	}
	if valueLen < maxValueSize {
		l.sizes[key] = valueLen
	}
	return value
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
