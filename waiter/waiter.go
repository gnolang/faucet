package waiter

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type WaitFunc func(ctx context.Context) error

// Waiter is a concept used for waiting on running services
type Waiter struct {
	ctx    context.Context
	cancel context.CancelFunc

	waitFns []WaitFunc
}

func New() *Waiter {
	w := &Waiter{
		waitFns: []WaitFunc{},
	}

	w.ctx, w.cancel = signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	return w
}

func (w *Waiter) Add(fns ...WaitFunc) {
	w.waitFns = append(w.waitFns, fns...)
}

func (w *Waiter) Wait() error {
	g, ctx := errgroup.WithContext(w.ctx)

	g.Go(func() error {
		<-ctx.Done()
		w.cancel()

		return nil
	})

	for _, fn := range w.waitFns {
		fn := fn

		g.Go(
			func() error {
				return fn(ctx)
			},
		)
	}

	return g.Wait()
}

func (w *Waiter) Context() context.Context {
	return w.ctx
}

func (w *Waiter) CancelFunc() context.CancelFunc {
	return w.cancel
}
