package spec

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
)

const DefaultProxyHostSuffix = ".proxy.com"
const DefaultEnvdPort = 48008

func GenProxyHeader(port int, sandboxID string) map[string]string {
	return map[string]string{"X-HOST": fmt.Sprintf("%d-%s%s",
		port, sandboxID, DefaultProxyHostSuffix)}
}

func SetSandboxHeader(header http.Header, port int, sandboxID, user string) {
	header.Set("X-Host", fmt.Sprintf("%d-%s%s", port, sandboxID, DefaultProxyHostSuffix))
	header.Set("X-User", user)
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
		SetSandboxHeader(req.Header(), i.envdPort, i.sandboxID, i.user)

		return next(ctx, req)
	}
}

func (i *headerInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)

		// Set the sandbox header
		SetSandboxHeader(conn.RequestHeader(), i.envdPort, i.sandboxID, i.user)

		return conn
	}
}

func (i *headerInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
