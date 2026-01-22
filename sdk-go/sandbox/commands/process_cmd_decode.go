package commands

import (
	"context"
	"encoding/json"
	"io"

	"github.com/sirupsen/logrus"
)

type Decoder[T any] interface {
	Decode(data []byte) (T, error)
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
		_, err := s.handle.Wait(ctx,
			WithStdout(
				func(data []byte) {
					if json.Valid(data) {
						s.events <- data
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
