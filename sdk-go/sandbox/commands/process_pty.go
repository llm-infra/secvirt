package commands

import (
	"context"
	"errors"
	"net"
	"net/http"
	"runtime"
	"time"

	"connectrpc.com/connect"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec/process"
	psConnect "github.com/llm-infra/secvirt/sdk-go/sandbox/spec/process/processconnect"
)

type Pty struct {
	client psConnect.ProcessClient
}

func NewPty(baseUrl, sandboxID, user string) *Pty {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}

	return &Pty{
		client: psConnect.NewProcessClient(
			httpClient,
			baseUrl,
			connect.WithInterceptors(
				spec.NewHeaderInterceptor(spec.DefaultEnvdPort, sandboxID, user),
			),
		),
	}
}

func (c *Pty) Kill(ctx context.Context, pid uint32) error {
	_, err := c.client.SendSignal(ctx, connect.NewRequest(&process.SendSignalRequest{
		Process: &process.ProcessSelector{
			Selector: &process.ProcessSelector_Pid{Pid: pid},
		},
		Signal: process.Signal_SIGNAL_SIGKILL,
	}))
	return err
}

func (c *Pty) SendStdin(ctx context.Context, pid uint32, data []byte) error {
	_, err := c.client.SendInput(ctx, connect.NewRequest(&process.SendInputRequest{
		Process: &process.ProcessSelector{
			Selector: &process.ProcessSelector_Pid{
				Pid: pid,
			},
		},
		Input: &process.ProcessInput{
			Input: &process.ProcessInput_Pty{
				Pty: data,
			},
		},
	}))
	return err
}

func (c *Pty) Create(
	ctx context.Context,
	size PtySize,
	envs map[string]string,
	cwd string,
) (*CommandHandle, error) {
	if envs == nil {
		envs = make(map[string]string)
	}
	envs["TERM"] = "xterm-256color"

	stream, err := c.client.Start(ctx, connect.NewRequest(&process.StartRequest{
		Process: &process.ProcessConfig{
			Cmd:  "/bin/bash",
			Args: []string{"-i", "-l"},
			Envs: envs,
			Cwd:  &cwd,
		},
		Pty: &process.PTY{
			Size: &process.PTY_Size{
				Rows: size.Rows,
				Cols: size.Cols,
			},
		},
	}))
	if err != nil {
		return nil, err
	}

	if !stream.Receive() {
		return nil, errors.New("failed to start process")
	}

	return &CommandHandle{
		pid:         stream.Msg().Event.GetStart().Pid,
		kill:        c.Kill,
		startStream: stream,
	}, nil
}

func (c *Pty) Resize(ctx context.Context, pid uint32, size PtySize) error {
	_, err := c.client.Update(ctx, connect.NewRequest(&process.UpdateRequest{
		Process: &process.ProcessSelector{
			Selector: &process.ProcessSelector_Pid{
				Pid: pid,
			},
		},
		Pty: &process.PTY{
			Size: &process.PTY_Size{
				Rows: size.Rows,
				Cols: size.Cols,
			},
		},
	}))
	return err
}
