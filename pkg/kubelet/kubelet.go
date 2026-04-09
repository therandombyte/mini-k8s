// Registers a node
// Periodically lists pods
// Picks the ones already assigned to node and marks them Running

package kubelet

import (
	"context"

	"github.com/therandombyte/mini-k8s/pkg/client"
)

type Kubelet struct {
	NodeName string
	Client *client.Client
}

func New(nodeName string, c *client.Client) *Kubelet {
	return &Kubelet{
		NodeName: nodeName,
		Client: c,
	}
}

func (k *Kubelet) RegisterNode(ctx context.Context) error {

	return nil
}

func (k *Kubelet) Run(ctx context.Context) error {

	return nil
}

func (k *Kubelet) syncPods(ctx context.Context) error {

	return nil
}


