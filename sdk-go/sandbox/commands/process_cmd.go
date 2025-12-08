package commands

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"connectrpc.com/connect"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec/process"
	psConnect "github.com/llm-infra/secvirt/sdk-go/sandbox/spec/process/processconnect"
)

type Cmd struct {
	client psConnect.ProcessClient
}

func NewCmd(baseUrl, sandboxID, user string) *Cmd {
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

	return &Cmd{
		client: psConnect.NewProcessClient(
			httpClient,
			baseUrl,
			connect.WithInterceptors(
				spec.NewHeaderInterceptor(spec.DefaultEnvdPort, sandboxID, user),
			),
		),
	}
}

func (c *Cmd) List(ctx context.Context) ([]ProcessInfo, error) {
	res, err := c.client.List(ctx, connect.NewRequest(&process.ListRequest{}))
	if err != nil {
		return nil, err
	}

	processes := make([]ProcessInfo, 0)
	for _, v := range res.Msg.GetProcesses() {
		config := v.GetConfig()
		processes = append(processes, ProcessInfo{
			pid:  v.Pid,
			tag:  v.GetTag(),
			cmd:  config.GetCmd(),
			args: config.GetArgs(),
			envs: config.GetEnvs(),
			cwd:  config.GetCwd(),
		})
	}
	return processes, nil
}

func (c *Cmd) Kill(ctx context.Context, pid uint32) error {
	_, err := c.client.SendSignal(ctx, connect.NewRequest(&process.SendSignalRequest{
		Process: &process.ProcessSelector{
			Selector: &process.ProcessSelector_Pid{Pid: pid},
		},
		Signal: process.Signal_SIGNAL_SIGKILL,
	}))
	return err
}

func (c *Cmd) SendStdin(ctx context.Context, pid uint32, data []byte) error {
	_, err := c.client.SendInput(ctx, connect.NewRequest(&process.SendInputRequest{
		Process: &process.ProcessSelector{
			Selector: &process.ProcessSelector_Pid{
				Pid: pid,
			},
		},
		Input: &process.ProcessInput{
			Input: &process.ProcessInput_Stdin{
				Stdin: data,
			},
		},
	}))
	return err
}

func (c *Cmd) Start(
	ctx context.Context,
	cmd string,
	envs map[string]string,
	cwd string,
) (*CommandHandle, error) {
	stream, err := c.client.Start(ctx, connect.NewRequest(&process.StartRequest{
		Process: &process.ProcessConfig{
			Cmd:  "/bin/bash",
			Args: []string{"-l", "-c", cmd},
			Envs: envs,
			Cwd:  &cwd,
		},
	}))
	if err != nil {
		return nil, err
	}

	if !stream.Receive() {
		return nil, fmt.Errorf("failed to start process: %s", stream.Err())
	}

	return &CommandHandle{
		pid:         stream.Msg().Event.GetStart().Pid,
		kill:        c.Kill,
		startStream: stream,
	}, nil
}

func (c *Cmd) Run(
	ctx context.Context,
	cmd string,
	envs map[string]string,
	cwd string,
	opts ...HandleOption,
) (*CommandResult, error) {
	h, err := c.Start(ctx, cmd, envs, cwd)
	if err != nil {
		return nil, err
	}

	return h.Wait(ctx, opts...)
}

func (c *Cmd) Connect(ctx context.Context, pid uint32) (*CommandHandle, error) {
	stream, err := c.client.Connect(ctx, connect.NewRequest(&process.ConnectRequest{
		Process: &process.ProcessSelector{
			Selector: &process.ProcessSelector_Pid{
				Pid: pid,
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
		pid:           stream.Msg().Event.GetStart().Pid,
		kill:          c.Kill,
		connectStream: stream,
	}, nil
}
