package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec/process"
)

type HandleOption func(*HandleOptions)

type HandleOptions struct {
	onStdout func([]byte)
	onStderr func([]byte)
	onPty    func([]byte)
}

func WithStdout(fn func([]byte)) HandleOption {
	return func(ho *HandleOptions) { ho.onStdout = fn }
}

func WithStderr(fn func([]byte)) HandleOption {
	return func(ho *HandleOptions) { ho.onStderr = fn }
}

func WithPty(fn func([]byte)) HandleOption {
	return func(ho *HandleOptions) { ho.onPty = fn }
}

type CommandHandle struct {
	pid           uint32
	kill          func(context.Context, uint32) error
	startStream   *connect.ServerStreamForClient[process.StartResponse]
	connectStream *connect.ServerStreamForClient[process.ConnectResponse]
}

func (h *CommandHandle) Pid() uint32 {
	return h.pid
}

func (c *CommandHandle) Disconnect() error {
	return c.startStream.Close()
}

func (h *CommandHandle) Wait(ctx context.Context, opts ...HandleOption) (*CommandResult, error) {
	opt := &HandleOptions{}
	for _, o := range opts {
		o(opt)
	}

	var result = &CommandResult{}
	var stdout strings.Builder
	var stderr strings.Builder

	for h.startStream.Receive() {
		var event *process.ProcessEvent
		if h.startStream != nil {
			msg := h.startStream.Msg()
			event = msg.GetEvent()
		} else if h.connectStream != nil {
			msg := h.connectStream.Msg()
			event = msg.GetEvent()
		} else {
			return nil, errors.New("none stream for client")
		}

		switch {
		case event.GetData() != nil:
			data := event.GetData()

			if data := data.GetStdout(); len(data) > 0 {
				stdout.WriteString(string(data))
				if opt.onStdout != nil {
					opt.onStdout(data)
				}
			}

			if data := data.GetStderr(); len(data) > 0 {
				stderr.WriteString(string(data))
				if opt.onStderr != nil {
					opt.onStderr(data)
				}
			}

			if pty := data.GetPty(); len(pty) > 0 && opt.onPty != nil {
				opt.onPty(pty)
			}

		case event.GetEnd() != nil:
			end := event.GetEnd()
			result = &CommandResult{
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				ExitCode: end.GetExitCode(),
				Error:    end.GetError(),
			}
		}
	}

	// If Receive stopped, capture any error
	if err := h.startStream.Err(); err != nil {
		return nil, fmt.Errorf("stream error: %w", err)
	}

	// If no end event received
	if result == nil {
		return nil, errors.New("command ended without end event")
	}

	if result.ExitCode != 0 {
		return nil, &CommandExitError{Result: *result}
	}

	return result, nil
}

func (c *CommandHandle) Kill() error {
	return c.kill(context.Background(), c.pid)
}

type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int32
	Error    string
}

type CommandExitError struct {
	Result CommandResult
}

func (e *CommandExitError) Error() string {
	return fmt.Sprintf("command exited with code %d: %s", e.Result.ExitCode, e.Result.Error)
}
