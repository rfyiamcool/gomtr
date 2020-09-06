package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupIps(t *testing.T) {
	ns, err := LookupIps("baidu.com")
	assert.Equal(t, err, nil)
	assert.Greater(t, len(ns), 0)

	ns, err = LookupIps("8.8.8.8")
	assert.Equal(t, err, nil)
	assert.Equal(t, len(ns), 1)
	assert.Equal(t, ns[0], "8.8.8.8")
}
