package desktop

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunOcServerWithRetry_RetriesRetryableErrors(t *testing.T) {
	t.Parallel()

	var attempts int
	var cleanups int
	err := runOcServerWithRetry(t.Context(), 3, time.Millisecond,
		func() error {
			attempts++
			if attempts < 4 {
				return &ocServerAttemptError{
					err:       errors.New("stderr error"),
					retryable: true,
				}
			}
			return nil
		},
		func() error {
			cleanups++
			return nil
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 4, attempts)
	assert.Equal(t, 3, cleanups)
}

func TestRunOcServerWithRetry_StopsOnNonRetryableError(t *testing.T) {
	t.Parallel()

	var cleanups int
	err := runOcServerWithRetry(t.Context(), 3, time.Millisecond,
		func() error {
			return errors.New("start error")
		},
		func() error {
			cleanups++
			return nil
		},
	)
	assert.EqualError(t, err, "start error")
	assert.Zero(t, cleanups)
}

func TestRunOcServerWithRetry_StopsWhenContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	var attempts int
	err := runOcServerWithRetry(ctx, 3, time.Second,
		func() error {
			attempts++
			cancel()
			return &ocServerAttemptError{
				err:       errors.New("stderr error"),
				retryable: true,
			}
		},
		func() error { return nil },
	)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, attempts)
}

func TestOcServerRunContextErr_InternalTimeoutIsRetryable(t *testing.T) {
	t.Parallel()

	parentCtx := context.Background()
	runCtx, cancel := context.WithDeadline(parentCtx, time.Now().Add(-time.Second))
	defer cancel()

	err := ocServerRunContextErr(parentCtx, runCtx)

	var attemptErr *ocServerAttemptError
	assert.ErrorAs(t, err, &attemptErr)
	assert.True(t, attemptErr.retryable)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestOcServerRunContextErr_ParentCancellationStopsRetry(t *testing.T) {
	t.Parallel()

	parentCtx, cancel := context.WithCancel(context.Background())
	runCtx, runCancel := context.WithTimeout(parentCtx, time.Second)
	cancel()
	defer runCancel()

	<-runCtx.Done()

	err := ocServerRunContextErr(parentCtx, runCtx)

	var attemptErr *ocServerAttemptError
	assert.False(t, errors.As(err, &attemptErr))
	assert.ErrorIs(t, err, context.Canceled)
}
