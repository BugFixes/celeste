package celeste_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bugfixes/celeste/cmd/celeste"
	"github.com/stretchr/testify/require"
)

type errorComponent struct{}

func (e *errorComponent) Run(ctx context.Context) error {
	return errors.New("error from component")
}

type goodComponent struct{}

func (g *goodComponent) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func TestRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := celeste.New(&goodComponent{})
	require.NoError(t, err)

	go func() {
		<-time.After(time.Second)
		cancel()
	}()

	err = a.Run(ctx)
	require.NoError(t, err)
}

func TestRunWithError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a, err := celeste.New(&errorComponent{})
	require.NoError(t, err)

	err = a.Run(ctx)
	require.Error(t, err)
}
