package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/sirupsen/logrus"
)

type Decoder[T any] interface {
	Decode(data []byte) (T, error)
}

type jsonLineBuffer struct {
	pending []byte
}

func newJSONLineBuffer() *jsonLineBuffer {
	return &jsonLineBuffer{}
}

func (b *jsonLineBuffer) Push(chunk []byte) [][]byte {
	if len(chunk) == 0 {
		return nil
	}

	b.pending = append(b.pending, chunk...)
	var lines [][]byte
	for {
		idx := bytes.IndexByte(b.pending, '\n')
		if idx < 0 {
			break
		}

		line := bytes.TrimSpace(b.pending[:idx])
		b.pending = b.pending[idx+1:]
		if len(line) == 0 {
			continue
		}
		if json.Valid(line) {
			lines = append(lines, append([]byte(nil), line...))
		}
	}
	return lines
}

func (b *jsonLineBuffer) Flush() [][]byte {
	line := bytes.TrimSpace(b.pending)
	b.pending = nil
	if len(line) == 0 || !json.Valid(line) {
		return nil
	}
	return [][]byte{append([]byte(nil), line...)}
}

type Stream[T any] struct {
	handle *CommandHandle
	events chan []byte

	decoder Decoder[T]
	err     error
	done    bool
}

func NewStream[T any](ctx context.Context, handle *CommandHandle, decoder Decoder[T]) *Stream[T] {
	s := &Stream[T]{
		handle: handle,
		events: make(chan []byte),

		decoder: decoder,
	}

	go func() {
		buffer := newJSONLineBuffer()
		_, err := s.handle.Wait(ctx,
			WithStdout(
				func(b []byte) {
					for _, line := range buffer.Push(b) {
						s.events <- line
					}
				},
			),
			WithStderr(
				func(b []byte) {
					logrus.WithContext(ctx).
						Debugf("command stderr: %s", string(b))
				},
			),
		)
		for _, line := range buffer.Flush() {
			s.events <- line
		}
		s.err = err
		s.done = true
		s.handle.kill(context.Background(), s.handle.pid)
		close(s.events)
	}()

	return s
}

func (s *Stream[T]) Recv() (T, error) {
	var nxt T

	if s.err != nil {
		return nxt, s.err
	}

	if s.done && len(s.events) == 0 {
		return nxt, io.EOF
	}

	raw, ok := <-s.events
	if !ok {
		return nxt, io.EOF
	}

	return s.decoder.Decode(raw)
}

func (s *Stream[T]) Close() error {
	if s.handle != nil {
		s.handle.Kill()
	}

	return nil
}
