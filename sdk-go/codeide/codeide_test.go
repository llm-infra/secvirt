package codeide

import (
	"context"
	"fmt"
	"testing"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/stretchr/testify/assert"
)

func TestPackages(t *testing.T) {
	sbx, err := NewSandbox(
		context.TODO(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	res, err := sbx.Packages(context.Background(), "python")
	assert.NoError(t, err)

	fmt.Println(res)
}

func TestRunCode(t *testing.T) {
	sbx, err := NewSandbox(
		context.TODO(),
		sandbox.WithHost("192.168.134.142"),
	)
	assert.NoError(t, err)

	res, err := sbx.RunCode(context.TODO(), "python", "print(123)", nil)
	assert.NoError(t, err)

	fmt.Println(res)
}
