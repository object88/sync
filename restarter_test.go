package sync

import (
	"context"
	"sync"
	"testing"
	"time"
)

func Test_WithOneInvocation(t *testing.T) {
	count := 0

	f := func(ctx context.Context) {
		time.Sleep(10 * time.Millisecond)
		count++
	}

	r := NewRestarter()
	r.Invoke(f)

	if count != 1 {
		t.Errorf("Incorrect count at end of test: %d\n", count)
	}
}

func Test_WithTwoSynchInvocations(t *testing.T) {
	count := 0

	f := func(ctx context.Context) {
		time.Sleep(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			return
		default:
			count++
		}
	}

	r := NewRestarter()
	r.Invoke(f)
	r.Invoke(f)

	if count != 2 {
		t.Errorf("Incorrect count at end of test: %d\n", count)
	}
}

func Test_WithTwoAsyncInvocations(t *testing.T) {
	count := 0
	var wg sync.WaitGroup

	f := func(ctx context.Context) {
		time.Sleep(10 * time.Millisecond)
		select {
		case <-ctx.Done():
		default:
			count++
		}

		wg.Done()
	}

	r := NewRestarter()
	wg.Add(2)

	go func() {
		time.Sleep(1 * time.Millisecond)
		r.Invoke(f)
	}()
	r.Invoke(f)

	wg.Wait()
	if count != 1 {
		t.Errorf("Incorrect count at end of test: %d\n", count)
	}
}

func Test_WithThreeAsyncInvocations(t *testing.T) {
	result := 0
	var wg sync.WaitGroup

	f := func(ctx context.Context, target int) {
		time.Sleep(10 * time.Millisecond)

		select {
		case <-ctx.Done():
		default:
			result = target
		}

		wg.Done()
	}

	r := NewRestarter()
	wg.Add(3)

	go func() {
		time.Sleep(2 * time.Millisecond)
		r.Invoke(func(ctx context.Context) {
			f(ctx, 3)
		})
	}()
	go func() {
		time.Sleep(1 * time.Millisecond)
		r.Invoke(func(ctx context.Context) {
			f(ctx, 2)
		})
	}()
	r.Invoke(func(ctx context.Context) {
		f(ctx, 1)
	})

	wg.Wait()
	if result != 3 {
		t.Errorf("Incorrect result at end of test: %d\n", result)
	}
}
