package allinone

import (
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/stretchr/testify/assert"
)

func TestCodeIDE(t *testing.T) {
	sbx, err := NewSandbox(
		t.Context(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	res, err := sbx.Packages(t.Context(), "python")
	if assert.NoError(t, err) {
		fmt.Println(res)
	}

	res, err = sbx.Packages(t.Context(), "javascript")
	if assert.NoError(t, err) {
		fmt.Println(res)
	}
}
