package helpers_test

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/aks-engine/pkg/helpers"
)

func TestNewRetryJob(t *testing.T) {
	job := helpers.NewRetryJob(1*time.Second)

	if job.Err != nil {
		t.Error("job state should not have an error set")
	}
}

func TestRetryJob_Do(t *testing.T) {
	job := helpers.NewRetryJob(1*time.Second)

	_, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var strVal string
	job.Do(context.Background(), func(ctx context.Context) (done bool, err error) {
		strVal = "bar"
		return true, nil
	})
	<-job.Done()

	if strVal != "bar" {
		t.Error("job should have 'bar'")
	}
}

func TestRetryJob_DoCancel(t *testing.T) {
	job := helpers.NewRetryJob(1*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	job.Do(ctx, func(ctx context.Context) (done bool, err error) {
		<-ctx.Done()
		return false, nil
	})

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	<-job.Done()
	if job.Err != context.Canceled {
		t.Error("job should error with context canceled")
	}
}

func TestRetryJob_DoManyTimes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	job := helpers.NewRetryJob(25*time.Millisecond)
	count := 0
	job.Do(ctx, func(ctx context.Context) (done bool, err error) {
		count = count + 1
		return false, nil
	})

	<-job.Done()
	if count <= 30 {
		t.Errorf("Expected a count greater than 30, but got a count of %d\n", count)
	}
}

func TestRetryJob_DoOnce(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	job := helpers.NewRetryJob(25*time.Millisecond)
	count := 0
	job.Do(ctx, func(ctx context.Context) (done bool, err error) {
		count = count + 1
		return true, nil
	})

	<-job.Done()
	if count != 1 {
		t.Error("count should only be incremented once")
	}
}
