package sandbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSandbox(t *testing.T) {
	sbx, err := NewSandbox(t.Context(),
		WithHost("10.20.152.105"))
	assert.NoError(t, err)

	sbx.Filesystem()
}
