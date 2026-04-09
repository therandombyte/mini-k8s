// go easy

package scheduler

import v1 "github.com/therandombyte/mini-k8s/pkg/api/v1"

type Scheduler struct {}

func New() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) PickNode(_ *v1.Pod, nodes []v1.Node) *v1.Node {
	for i := range nodes {
		if !nodes[i].Spec.Unschedulable {
			return &nodes[i]
		}
	}
	return nil
}
