package monit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Tail(addr, service string) {
	url := fmt.Sprintf("http://%s/services/%s", addr, service)
	tail(url)
}

var dataLinePrefix = []byte("data:")
var heartbeatLinepPrefix = []byte("event: heartbeat")

func tail(url string) error {
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	logLine := NewLogLine()
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
