package commands

import (
	"context"
	"net"
	"net/http"
	"runtime"
	"time"

	"connectrpc.com/connect"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec/process"
	psConnect "github.com/llm-infra/secvirt/sdk-go/sandbox/spec/process/processconnect"
)

type Lease struct {
	SandboxID string
	LeaseID   string
}

type Allocator interface {
	Acquire(ctx context.Context) (*Lease, error)
	Release(ctx context.Context, leaseID string) error
}

func NewScheduledCmd(baseURL, sandboxID, user string, allocator Allocator) *Cmd {
	cmd := NewCmd(baseURL, sandboxID, user)
	cmd.allocator = allocator
	return cmd
}

func (c *Cmd) clientForSandbox(sandboxID string) psConnect.ProcessClient {
	return psConnect.NewProcessClient(
		newHTTPClient(),
		c.baseURL,
		connect.WithInterceptors(
			spec.NewHeaderInterceptor(spec.DefaultEnvdPort, sandboxID, c.user),
		),
	)
}

func newHTTPClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	return &http.Client{
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
}

func killWithClient(client psConnect.ProcessClient) func(context.Context, uint32) error {
	return func(ctx context.Context, pid uint32) error {
		_, err := client.SendSignal(ctx, connect.NewRequest(&process.SendSignalRequest{
			Process: &process.ProcessSelector{
				Selector: &process.ProcessSelector_Pid{Pid: pid},
			},
			Signal: process.Signal_SIGNAL_SIGKILL,
		}))
		return err
	}
}
