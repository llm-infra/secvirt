package filesystem

import (
	"bytes"
	"context"
	"io"
	"os"

	"connectrpc.com/connect"
	"github.com/llm-infra/secvirt/sdk-go/sandbox/spec/filesystem"
	fsConnect "github.com/llm-infra/secvirt/sdk-go/sandbox/spec/filesystem/filesystemconnect"
)

type FileSource interface {
	Reader() (io.ReadCloser, error)
}

type PathSource string

func (s PathSource) Reader() (io.ReadCloser, error) {
	return os.Open(string(s))
}

type BytesSource []byte

func (s BytesSource) Reader() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s)), nil
}

type ReaderSource struct {
	r io.Reader
}

func (s ReaderSource) Reader() (io.ReadCloser, error) {
	return io.NopCloser(s.r), nil
}

func write[T FileSource](
	ctx context.Context,
	client fsConnect.FilesystemClient,
	path string,
	source T,
) error {
	reader, err := source.Reader()
	if err != nil {
		return err
	}
	defer reader.Close()

	stream := client.Write(ctx)
	defer func() {
		if stream != nil {
			stream.CloseAndReceive()
		}
	}()

	buf := make([]byte, 64*1024)

	for {
		n, err := reader.Read(buf)
		if n > 0 {
			msg := &filesystem.WriteRequest{
				Path:  path,
				Chunk: buf[:n],
			}

			if err := stream.Send(msg); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

type streamReader struct {
	stream *connect.ServerStreamForClient[filesystem.ReadResponse]
	buffer []byte
}

func (r *streamReader) Read(p []byte) (int, error) {
	if len(r.buffer) > 0 {
		n := copy(p, r.buffer)
		r.buffer = r.buffer[n:]
		return n, nil
	}

	if !r.stream.Receive() {
		return 0, io.EOF
	}

	if err := r.stream.Err(); err != nil {
		return 0, err
	}

	chunk := r.stream.Msg().Chunk
	if len(chunk) == 0 {
		return 0, io.EOF
	}

	n := copy(p, chunk)
	r.buffer = chunk[n:]
	return n, nil
}

func (r *streamReader) Close() error {
	if r.stream != nil {
		return r.stream.Close()
	}
	return nil
}
