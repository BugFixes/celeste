package celeste

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Celeste struct {
	version    string
	commitHash string
	httpRouter Component
}

//go:generate mockery --name=Component
type Component interface {
	Run(ctx context.Context) error
}

var version string
var commitHash string

func New(router Component) (*Celeste, error) {
	return &Celeste{
		version:    version,
		commitHash: commitHash,
		httpRouter: router,
	}, nil
}

func Version() string {
	return version
}

func CommitHash() string {
	return commitHash
}

func (c Celeste) Run(ctx context.Context) error {
	errGrp, errCtx := errgroup.WithContext(ctx)
	errGrp.Go(func() error {
		return c.httpRouter.Run(errCtx)
	})

	return errGrp.Wait()
}
