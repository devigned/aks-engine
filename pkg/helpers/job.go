package helpers

import (
	"context"
	"sync"
	"time"
)

type (
	RetryJob struct {
		sleep   time.Duration
		timeout time.Duration
		once    sync.Once
		done    chan struct{}
		State   *JobState
	}

	WorkFunc func(ctx context.Context, state *JobState)

	JobState struct {
		Values map[string]interface{}
		Done   bool
		Err    error
	}
)

func NewRetryJob(sleep, timeout time.Duration) *RetryJob {
	return &RetryJob{
		sleep:   sleep,
		timeout: timeout,
		State: &JobState{
			Values: make(map[string]interface{}),
		},
		done: make(chan struct{}),
	}
}

func (rj *RetryJob) Do(ctx context.Context, work WorkFunc) {
	ctx, cancel := context.WithTimeout(ctx, rj.timeout)
	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				if rj.State.Err == nil {
					rj.State.Err = ctx.Err()
				}
				rj.State.Done = true
				rj.closeDone()
				return
			default:
				work(ctx, rj.State)
				if rj.State.Done {
					rj.closeDone()
					return
				}

				func(ctx context.Context) {
					sleepCtx, cancel := context.WithTimeout(ctx, rj.sleep)
					defer cancel()
					<-sleepCtx.Done()
				}(ctx)
			}
		}
	}()
}

func (rj *RetryJob) Done() <-chan struct{} {
	return rj.done
}

func (rj *RetryJob) closeDone() {
	rj.once.Do(func() {
		close(rj.done)
	})
}
