package greedy

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	Name = "greedy"
)

var _ framework.ScorePlugin = &Greedy{}

type Greedy struct {
	handle framework.Handle
}

func New(_ runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &Greedy{
		handle: handle,
	}, nil
}

func (ci *Greedy) Name() string {
	return Name
}
