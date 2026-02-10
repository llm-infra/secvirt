package desktop

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/llm-infra/secvirt/sdk-go/desktop/opencode"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/commands"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/mel2oo/go-dkit/json"
	"github.com/mel2oo/go-dkit/otel"
	oc "github.com/sst/opencode-sdk-go"
	oco "github.com/sst/opencode-sdk-go/option"
	"github.com/sst/opencode-sdk-go/packages/ssestream"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"mvdan.cc/xurls/v2"
)

// model provider mcp
func (s *Sandbox) SetOpenCodeConfig(ctx context.Context, config *opencode.Config,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return s.Filesystem().Write(ctx,
		filepath.Join(opt.cwd, ".opencode", "opencode.json"),
		data,
	)
}

func (s *Sandbox) SetOpenCodeSkills(ctx context.Context, skills map[string]io.Reader,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	skillPath := filepath.Join(opt.cwd, ".opencode", "skills")
	for name, skill := range skills {
		data, err := io.ReadAll(skill)
		if err != nil {
			return err
		}

		temp := filepath.Join(skillPath, name)
		if err = s.Filesystem().Write(ctx, temp, data); err != nil {
			return err
		}

		_, err = s.Cmd().Run(ctx,
			fmt.Sprintf("unzip -o %s && rm -rf %s", temp, temp),
			nil,
			skillPath,
			false,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Sandbox) SetOpenCodeAgents(ctx context.Context, agents map[string]io.Reader,
	opts ...Option) error {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	for name, agent := range agents {
		path := filepath.Join(opt.cwd, ".opencode", "agents", name+".md")

		data, err := io.ReadAll(agent)
		if err != nil {
			return err
		}

		if err = s.Filesystem().Write(ctx, path, data); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sandbox) RunOcServer(ctx context.Context, port int, opts ...Option) (err error) {
	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	s.ocHandle, err = s.Cmd().Start(ctx,
		fmt.Sprintf("opencode serve --hostname 0.0.0.0 --port %d", port),
		opt.envs,
		opt.cwd,
		opt.stdin,
	)
	if err != nil {
		return err
	}

	urlCh := make(chan *url.URL)
	errCh := make(chan error)

	go func() {
		rxStrict := xurls.Strict()

		s.ocHandle.Wait(ctx,
			commands.WithStdout(
				func(b []byte) {
					urlStr := rxStrict.Find(b)
					if len(urlStr) == 0 {
						return
					}
					url, err := url.Parse(string(urlStr))
					if err == nil {
						urlCh <- url
					}
				},
			),
			commands.WithStderr(
				func(b []byte) {
					errCh <- errors.New(string(b))
				},
			),
		)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case err = <-errCh:
		return err

	case url := <-urlCh:
		port, err := strconv.Atoi(url.Port())
		if err != nil {
			return err
		}

		tr := otelhttp.NewTransport(
			spec.NewHeaderRoundTripper(
				spec.GenSandboxHeader(port, s.Name, ""),
				http.DefaultTransport,
			),
			otelhttp.WithTracerProvider(otel.Standard().TracerProvider),
			otelhttp.WithPropagators(otel.Standard().Propagators),
			otelhttp.WithSpanNameFormatter(otel.HttpSpanNameFormatter),
		)

		s.ocClient = oc.NewClient(
			oco.WithBaseURL(s.ProxyBaseURL()),
			oco.WithHTTPClient(&http.Client{Transport: tr}),
		)
		return err
	}
}

func (s *Sandbox) CloseOcServer() error {
	if s.ocHandle != nil {
		return s.ocHandle.Kill()
	}
	return nil
}

func (s *Sandbox) OcClient() *oc.Client {
	return s.ocClient
}

func (s *Sandbox) OpenCodeChat(ctx context.Context, content string,
	opts ...Option) (*oc.SessionPromptResponse, error) {
	if s.ocClient == nil {
		return nil, errors.New("opencode server not started")
	}

	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	var session *oc.Session
	var err error
	if len(opt.sessionID) > 0 {
		session, err = s.ocClient.Session.Get(ctx, opt.sessionID, oc.SessionGetParams{})
		if err != nil {
			return nil, err
		}
	} else {
		session, err = s.ocClient.Session.New(ctx, oc.SessionNewParams{})
		if err != nil {
			return nil, err
		}
	}

	// 发送消息
	params := oc.SessionPromptParams{
		Parts: oc.F([]oc.SessionPromptParamsPartUnion{
			oc.SessionPromptParamsPart{
				Type: oc.F(oc.SessionPromptParamsPartsTypeText),
				Text: oc.F(content),
			},
		}),
	}
	if len(opt.agent) > 0 {
		params.Agent = oc.F(opt.agent)
	}

	return s.ocClient.Session.Prompt(ctx, session.ID, params)
}

func (s *Sandbox) OpenCodeStreamChat(ctx context.Context, content string,
	opts ...Option) (*OpenCodeStream, error) {
	if s.ocClient == nil {
		return nil, errors.New("opencode server not started")
	}

	opt := NewOptions(s)
	for _, o := range opts {
		o(opt)
	}

	var session *oc.Session
	var err error
	if len(opt.sessionID) > 0 {
		session, err = s.ocClient.Session.Get(ctx, opt.sessionID, oc.SessionGetParams{})
		if err != nil {
			return nil, err
		}
	} else {
		session, err = s.ocClient.Session.New(ctx, oc.SessionNewParams{})
		if err != nil {
			return nil, err
		}
	}

	os := &OpenCodeStream{
		Stream: s.ocClient.Event.ListStreaming(ctx, oc.EventListParams{}),
		events: make(chan oc.EventListResponse),
	}

	// 接收消息流
	go func() {
		for os.Next() {
			os.events <- os.Current()
		}

		if err := os.Err(); err != nil {
			os.err = err
		}

		os.done = true
		close(os.events)
	}()

	// 发送消息
	params := oc.SessionPromptParams{
		Parts: oc.F([]oc.SessionPromptParamsPartUnion{
			oc.SessionPromptParamsPart{
				Type: oc.F(oc.SessionPromptParamsPartsTypeText),
				Text: oc.F(content),
			},
		}),
	}
	if len(opt.agent) > 0 {
		params.Agent = oc.F(opt.agent)
	}

	go func() {
		_, err := s.ocClient.Session.Prompt(ctx, session.ID, params)
		if err != nil {
			os.err = err
		}
	}()

	return os, nil
}

type OpenCodeStream struct {
	*ssestream.Stream[oc.EventListResponse]
	events chan oc.EventListResponse

	err  error
	done bool
}

func (s *OpenCodeStream) Recv() (*oc.EventListResponse, error) {
	if s.err != nil {
		return nil, s.err
	}

	if s.done && len(s.events) == 0 {
		return nil, io.EOF
	}

	raw, ok := <-s.events
	if !ok {
		return nil, io.EOF
	}

	return &raw, nil
}
