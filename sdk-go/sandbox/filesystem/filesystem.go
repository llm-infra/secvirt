package filesystem

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"time"

	"connectrpc.com/connect"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec/filesystem"
	fsConnect "github.com/llm-infra/secvirt/sdk-go/sandbox/spec/filesystem/filesystemconnect"
)

type Filesystem struct {
	client fsConnect.FilesystemClient
}

func NewFileSystem(baseUrl, sandboxID, user string) *Filesystem {
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

	return &Filesystem{
		client: fsConnect.NewFilesystemClient(
			httpClient,
			baseUrl,
			connect.WithInterceptors(
				spec.NewHeaderInterceptor(spec.DefaultEnvdPort, sandboxID, user),
			),
		),
	}
}

func (f *Filesystem) Read(ctx context.Context, path string) ([]byte, error) {
	stream, err := f.client.Read(
		ctx,
		connect.NewRequest(&filesystem.ReadRequest{Path: path}),
	)
	if err != nil {
		return nil, err
	}

	allChunk := make([]byte, 0)
	for stream.Receive() {
		chunk := stream.Msg().Chunk
		allChunk = append(allChunk, chunk...)
	}

	if err := stream.Err(); err != nil {
		return nil, err
	}

	return allChunk, nil
}

func (f *Filesystem) ReadStream(ctx context.Context, path string) (*streamReader, error) {
	stream, err := f.client.Read(
		ctx,
		connect.NewRequest(&filesystem.ReadRequest{Path: path}),
	)
	if err != nil {
		return nil, err
	}
	return &streamReader{stream: stream}, nil
}

func (f *Filesystem) Write(ctx context.Context, path string, source any) error {
	switch s := source.(type) {
	case string:
		return write(ctx, f.client, path, PathSource(s))
	case []byte:
		return write(ctx, f.client, path, BytesSource(s))
	case io.Reader:
		return write(ctx, f.client, path, ReaderSource{r: s})
	default:
		return fmt.Errorf("unsupported source type: %T", source)
	}
}

func (f *Filesystem) List(ctx context.Context, path string, depth int) ([]*filesystem.EntryInfo, error) {
	res, err := f.client.ListDir(
		ctx,
		connect.NewRequest(&filesystem.ListDirRequest{
			Path:  path,
			Depth: uint32(depth),
		}),
	)
	if err != nil {
		return nil, err
	}

	return res.Msg.Entries, nil
}

func (f *Filesystem) Exist(ctx context.Context, path string) (bool, error) {
	_, err := f.client.Stat(
		ctx,
		connect.NewRequest(&filesystem.StatRequest{Path: path}),
	)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (f *Filesystem) Remove(ctx context.Context, path string) error {
	_, err := f.client.Remove(
		ctx,
		connect.NewRequest(&filesystem.RemoveRequest{Path: path}),
	)
	return err
}

func (f *Filesystem) Rename(ctx context.Context, old, new string) (*filesystem.EntryInfo, error) {
	res, err := f.client.Move(
		ctx,
		connect.NewRequest(&filesystem.MoveRequest{
			Source:      old,
			Destination: new,
		}),
	)
	if err != nil {
		return nil, err
	}
	return res.Msg.Entry, nil
}

func (f *Filesystem) Mkdir(ctx context.Context, path string) (bool, error) {
	_, err := f.client.MakeDir(
		ctx,
		connect.NewRequest(&filesystem.MakeDirRequest{Path: path}),
	)
	if err != nil {
		conErr, ok := err.(*connect.Error)
		if ok {
			if conErr.Code() == connect.CodeAlreadyExists {
				return true, nil
			}
		}
		return false, err
	}

	return true, nil
}
