package helpers_test

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/aks-engine/pkg/helpers"
)

func TestNewRetryJob(t *testing.T) {
	job := helpers.NewRetryJob(1*time.Second, 30*time.Second)
	if job.State.Values == nil {
		t.Error("job state should be initialized")
	}

	if job.State.Done == true {
		t.Error("job state should not be done")
	}

	if job.State.Err != nil {
		t.Error("job state should not have an error set")
	}
}

func TestRetryJob_Do(t *testing.T) {
	job := helpers.NewRetryJob(1*time.Second, 1*time.Second)
	job.Do(context.Background(), func(ctx context.Context, state *helpers.JobState) {
		state.Done = true
		state.Values["foo"] = "bar"
	})
	<-job.Done()

	if job.State.Done != true {
		t.Error("job should be done")
	}

	if job.State.Values["foo"] != "bar" {
		t.Error("job should have Value['foo'] = 'bar'")
	}
}

func TestRetryJob_DoCancel(t *testing.T) {
	job := helpers.NewRetryJob(1*time.Second, 50*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	job.Do(ctx, func(ctx context.Context, state *helpers.JobState) {
		<- ctx.Done()
	})

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	<- job.Done()
	if job.State.Err != context.Canceled {
		t.Error("job should error with context canceled")
	}
}

func TestRetryJob_DoManyTimes(t *testing.T) {
	job := helpers.NewRetryJob(25*time.Millisecond, 1*time.Second)
	job.Do(context.Background(), func(ctx context.Context, state *helpers.JobState) {
		if state.Values["count"] == nil {
			state.Values["count"] = 1
			return
		}

		if count, ok := state.Values["count"].(int); ok {
			state.Values["count"] = count + 1
		}
	})

	<- job.Done()
	count, _ := job.State.Values["count"].(int)
	if count <= 30 {
		t.Errorf("Expected a count greater than 30, but got a count of %d\n", count)
	}
}
