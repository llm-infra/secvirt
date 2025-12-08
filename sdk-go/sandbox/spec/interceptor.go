package spec

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/mel2oo/go-dkit/ext"
)

const DefaultProxyHostSuffix = ".proxy.com"
const DefaultEnvdPort = 48008

func GenSandboxHeader(port int, sandboxID, user string) map[string]string {
	headers := map[string]string{
		"X-HOST": fmt.Sprintf("%d-%s%s", port, sandboxID, DefaultProxyHostSuffix),
	}

	if len(user) > 0 {
		headers["X-USER"] = user
	}

	return headers
}

type headerInterceptor struct {
	envdPort  int
	sandboxID string
	user      string
}

func NewHeaderInterceptor(envdPort int, sandboxID string, user string) *headerInterceptor {
	return &headerInterceptor{envdPort: envdPort, sandboxID: sandboxID, user: user}
}

func (i *headerInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Set the sandbox header
		headers := GenSandboxHeader(i.envdPort, i.sandboxID, i.user)
		for k, v := range headers {
			req.Header().Set(k, v)
		}
		ext.InjectHeader(ctx, req.Header())
		return next(ctx, req)
	}
}

func (i *headerInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)

		// Set the sandbox header
		headers := GenSandboxHeader(i.envdPort, i.sandboxID, i.user)
		for k, v := range headers {
			conn.RequestHeader().Set(k, v)
		}
		ext.InjectHeader(ctx, conn.RequestHeader())
		return conn
	}
}

func (i *headerInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

type headerRoundTripper struct {
	headers map[string]string
	rt      http.RoundTripper
}

func NewHeaderRoundTripper(headers map[string]string, rt http.RoundTripper) http.RoundTripper {
	return &headerRoundTripper{
		headers: headers,
		rt:      rt,
	}
}

func (h *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
	return h.rt.RoundTrip(req)
}
