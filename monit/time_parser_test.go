package monit

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testNow time.Time

func init() {
	var err error
	testNow, err = time.Parse(time.RFC3339, "2018-01-02T15:04:05+00:00")
	if err != nil {
		panic(err)
	}
}

func TestParseTime(t *testing.T) {
	cases := []struct {
		s string
		t string
	}{
		{"3 days ago", "2017-12-30 15:04:05"},
		{"1 days ago", "2018-01-01 15:04:05"},
		{"3 hours ago", "2018-01-02 12:04:05"},
		{"1 hour ago", "2018-01-02 14:04:05"},
		{"3 minutes ago", "2018-01-02 15:01:05"},
		{"1 minute ago", "2018-01-02 15:03:05"},
		{"16:23", "2018-01-02 16:23:00"},
		{"08.03. 16:23", "2018-03-08 16:23:00"},
		{"2017-12-18T10:21:46.580719+01:00", "2017-12-18 10:21:46"},
		{"2017-12-18T10:21:46", "2017-12-18 10:21:46"},
	}

	for i, c := range cases {
		tm, err := parseTime(testNow, c.s)
		assert.Nil(t, err)
		assert.Equal(t, c.t, tm.Format("2006-01-02 15:04:05"), fmt.Sprintf("case no %d", i))
	}
}
