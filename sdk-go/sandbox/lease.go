package sandbox

import (
	"context"

	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
)

type sandboxAllocator struct {
	sandbox     *Sandbox
	user        string
	template    TemplateType
	healthPorts []int
}

func (a *sandboxAllocator) Acquire(ctx context.Context) (*commands.Lease, error) {
	resp, err := a.sandbox.ApiRequest(ctx).
		SetContext(ctx).
		SetBody(map[string]any{
			"user_id":      a.user,
			"template":     a.template,
			"health_ports": a.healthPorts,
		}).
		SetResult(SandboxAllocateResponse{}).
		SetError(ErrorResponse{}).
		Post("/secvirt/v2/sandboxes/allocate")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(*ErrorResponse)
	}
	res := resp.Result().(*SandboxAllocateResponse)
	return &commands.Lease{SandboxID: res.SandboxID, LeaseID: res.LeaseID}, nil
}

func (a *sandboxAllocator) Release(ctx context.Context, leaseID string) error {
	resp, err := a.sandbox.ApiRequest(ctx).
		SetContext(ctx).
		SetBody(map[string]any{"lease_id": leaseID}).
		SetError(ErrorResponse{}).
		Post("/secvirt/v2/sandboxes/release")
	if err != nil {
		return err
	}
	if resp.IsError() {
		return resp.Error().(*ErrorResponse)
	}
	return nil
}
