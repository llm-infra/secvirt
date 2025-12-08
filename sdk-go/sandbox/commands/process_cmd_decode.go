package commands

import (
	"context"
	"encoding/json"
	"io"
)

type Decoder interface {
	Decode(data []byte, v any) error
}

type Stream[T any] struct {
	handle *CommandHandle
	events chan []byte

	decoder Decoder
	err     error
	done    bool
}

func NewStream[T any](ctx context.Context, handle *CommandHandle, decoder Decoder) *Stream[T] {
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
		)
		s.err = err
		s.done = true
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

	if err := s.decoder.Decode(raw, &nxt); err != nil {
		return nxt, err
	}

	return nxt, nil
}

func (s *Stream[T]) Close() error {
	if s.handle != nil {
		s.handle.Kill()
	}

	return nil
}
