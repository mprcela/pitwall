package deploy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTag(t *testing.T) {
	tag := NewTag("20160613151056.99a146a.b8a1fbf.747da38", false)
	assert.False(t, tag.created.IsZero())
	assert.Equal(t, "2016-06-13T15:10:56Z", tag.created.Format(time.RFC3339))

	tag = NewTag("pero", true)
	assert.True(t, tag.created.IsZero())

}
