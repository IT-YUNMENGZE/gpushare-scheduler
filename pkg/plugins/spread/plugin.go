package spread

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	Name = "spread"
)

var _ framework.ScorePlugin = &Spread{}
 
type Spread struct {
	handle framework.Handle
}

func New(_ runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &Spread{
		handle: handle,
	}, nil
}

func (ci *Spread) Name() string {
	return Name
}
