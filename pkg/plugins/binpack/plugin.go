package binpack

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	Name = "binpack"
)

var _ framework.ScorePlugin = &Binpack{}

type Binpack struct {
	handle framework.Handle
}

func New(_ runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &Binpack{
		handle: handle,
	}, nil
}

func (ci *Binpack) Name() string {
	return Name
}
