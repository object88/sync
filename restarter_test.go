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
	if r.ctx != nil {
		t.Error("r.ctx should be null")
	}
	if r.c != nil {
		t.Error("r.c should be null")
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
	if r.ctx != nil {
		t.Error("r.ctx should be null")
	}
	if r.c != nil {
		t.Error("r.c should be null")
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
	}

	r := NewRestarter()
	wg.Add(2)

	go func() {
		time.Sleep(1 * time.Millisecond)
		r.Invoke(f)
		wg.Done()
	}()
	go func() {
		r.Invoke(f)
		wg.Done()
	}()

	wg.Wait()
	if count != 1 {
		t.Errorf("Incorrect count at end of test: %d\n", count)
	}
	if r.ctx != nil {
		t.Error("r.ctx should be null")
	}
	if r.c != nil {
		t.Error("r.c should be null")
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
	}

	r := NewRestarter()
	wg.Add(3)

	go func() {
		time.Sleep(2 * time.Millisecond)
		r.Invoke(func(ctx context.Context) {
			f(ctx, 3)
		})
		wg.Done()
	}()
	go func() {
		time.Sleep(1 * time.Millisecond)
		r.Invoke(func(ctx context.Context) {
			f(ctx, 2)
		})
		wg.Done()
	}()
	go func() {
		r.Invoke(func(ctx context.Context) {
			f(ctx, 2)
		})
		wg.Done()
	}()

	wg.Wait()
	if result != 3 {
		t.Errorf("Incorrect result at end of test: %d\n", result)
	}
	if r.ctx != nil {
		t.Error("r.ctx should be null")
	}
	if r.c != nil {
		t.Error("r.c should be null")
	}
}

func Test_WithThreeActors(t *testing.T) {
	// See https://github.com/object88/sync/issues/1

	r := NewRestarter()

	// A starts executing
	ctxA, cancelFnA := r.spinUp()
	if r.ctx == nil {
		t.Error("r.ctx should have a context after A spinUp")
	}
	if *r.ctx != ctxA {
		t.Error("r.ctx has wrong context after A spinUp")
	}
	if r.c == nil {
		t.Error("r.c should have a cancelFn after A spinUp")
	}

	// B starts executing
	ctxB, _ := r.spinUp()
	if r.ctx == nil {
		t.Error("r.ctx should have a context after B spinUp")
	}
	if *r.ctx != ctxB {
		t.Error("r.ctx has wrong context after B spinUp")
	}
	if r.c == nil {
		t.Error("r.c should have a cancelFn after B spinUp")
	}

	// A stops processing; should have a context from spinning up B.
	r.spinDown(cancelFnA)
	if r.ctx == nil {
		t.Error("r.ctx should have a context after A spinDown")
	}
	if *r.ctx != ctxB {
		t.Error("r.ctx has wrong context after A spinDown")
	}
	if r.c == nil {
		t.Fatalf("restarter does not have a context")
	}
}
