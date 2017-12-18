package monit

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeParse(t *testing.T) {
	v := "2017-12-18T10:21:46.580719+01:00"
	tm, err := time.Parse("2006-01-02T15:04:05.999999-07:00", v)
	assert.Nil(t, err)
	fmt.Print(tm)
}
