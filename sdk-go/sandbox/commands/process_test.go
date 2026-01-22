package commands

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmd(t *testing.T) {
	cmd := NewCmd(
		"http://192.168.134.142:48008",
		"",
		"root",
	)

	res, err := cmd.Run(t.Context(), "gemini-a2a-server", nil, "", true)
	assert.NoError(t, err)
	fmt.Println(res)
}

func TestPty(t *testing.T) {
	pty := NewPty(
		"http://192.168.134.142:48008",
		"",
		"root",
	)
	h, err := pty.Create(t.Context(), PtySize{Cols: 80, Rows: 24}, nil, "")
	assert.NoError(t, err)

	err = pty.SendStdin(t.Context(), h.Pid(), []byte("ls -a"))
	assert.NoError(t, err)

	res, err := h.Wait(t.Context(), WithPty(func(b []byte) { fmt.Println(string(b)) }))
	assert.NoError(t, err)
	fmt.Println(res)

	err = pty.SendStdin(t.Context(), h.Pid(), []byte("ifconfig"))
	assert.NoError(t, err)

	res, err = h.Wait(t.Context())
	assert.NoError(t, err)
	fmt.Println(res)
}
