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
		Err     error
	}

	WorkFunc func(ctx context.Context) (done bool, err error)
)

func NewRetryJob(sleep, timeout time.Duration) *RetryJob {
	return &RetryJob{
		sleep:   sleep,
		timeout: timeout,
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
				if rj.Err == nil {
					rj.Err = ctx.Err()
				}
				rj.closeDone()
				return
			default:
				done, err := work(ctx)
				rj.Err = err
				if done {
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
