package sandbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSandbox(t *testing.T) {
	sbx, err := NewSandbox(t.Context(),
		WithHost("192.168.134.142"))
	assert.NoError(t, err)

	sbx.Filesystem()
}
