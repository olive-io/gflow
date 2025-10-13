package runner

import (
	"context"
	"testing"
)

var _ Task = (*SimpleTask)(nil)

type SimpleTask struct{}

func (s SimpleTask) Commit(ctx context.Context) (any, error) {
	//TODO implement me
	panic("implement me")
}

func (s SimpleTask) Rollback(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s SimpleTask) Destroy(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s SimpleTask) String() string {
	//TODO implement me
	panic("implement me")
}

func TestExtractTask(t *testing.T) {

}
