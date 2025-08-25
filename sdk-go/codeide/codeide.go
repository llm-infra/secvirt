package codeide

import (
	"context"
	"errors"
	"fmt"

	"github.com/llm-infra/secvirt/sdk-go/sandbox"
	"github.com/mel2oo/go-dkit/json"
)

var defaultCodeIDEPort = 8000

type Sandbox struct {
	*sandbox.Sandbox
}

func NewSandbox(ctx context.Context, opts ...sandbox.Option) (*Sandbox, error) {
	client, err := sandbox.NewSandbox(ctx,
		append(opts,
			sandbox.WithTemplate(sandbox.TemplateCodeIDE),
			sandbox.WithHealthPorts([]int{defaultCodeIDEPort}))...)
	if err != nil {
		return nil, err
	}

	return &Sandbox{Sandbox: client}, nil
}

func (s *Sandbox) Packages(ctx context.Context, lang string) ([]PackagesResponse, error) {
	resp, err := s.ProxyRequest(ctx, defaultCodeIDEPort).
		SetResult([]PackagesResponse{}).
		SetError(sandbox.ErrorResponse{}).
		Get(fmt.Sprintf("/codeide/v1/packages/%s", lang))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(*sandbox.ErrorResponse)
	}

	result := resp.Result().(*[]PackagesResponse)
	return *result, nil
}

func (s *Sandbox) RunCodeV1(ctx context.Context, lang, code string,
	inputs map[string]any) (*RunCodeResponseV1, error) {
	resp, err := s.ProxyRequest(ctx, defaultCodeIDEPort).
		SetBody(map[string]any{
			"lang":   lang,
			"code":   code,
			"inputs": inputs,
		}).
		SetResult(RunCodeResponse{}).
		SetError(sandbox.ErrorResponse{}).
		Post("/codeide/v1/execute")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(*sandbox.ErrorResponse)
	}

	res := resp.Result().(*RunCodeResponse)
	if res.Result == nil && res.Errors != nil {
		return nil, errors.New(json.MarshalPureString(res.Errors))
	}

	output, ok := res.Result.(string)
	if ok {
		return &RunCodeResponseV1{
			Output:  output,
			Console: json.MarshalPureString(res.Stdouts),
		}, nil
	} else {
		return &RunCodeResponseV1{
			Output:  json.MarshalPureString(res.Result),
			Console: json.MarshalPureString(res.Stdouts),
		}, nil
	}
}

func (s *Sandbox) RunCode(ctx context.Context, lang, code string,
	inputs map[string]any) (*RunCodeResponse, error) {
	resp, err := s.ProxyRequest(ctx, defaultCodeIDEPort).
		SetBody(map[string]any{
			"lang":   lang,
			"code":   code,
			"inputs": inputs,
		}).
		SetResult(RunCodeResponse{}).
		SetError(sandbox.ErrorResponse{}).
		Post("/codeide/v1/execute")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, resp.Error().(*sandbox.ErrorResponse)
	}

	return resp.Result().(*RunCodeResponse), nil
}
