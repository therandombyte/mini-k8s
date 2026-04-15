// Registers a node
// Periodically lists pods
// Picks the ones already assigned to node and marks them Running

package kubelet

import (
	"context"
	"log"
	"time"

	v1 "github.com/therandombyte/mini-k8s/pkg/api/v1"
	"github.com/therandombyte/mini-k8s/pkg/apimachinery"
	"github.com/therandombyte/mini-k8s/pkg/client"
	rt "github.com/therandombyte/mini-k8s/pkg/runtime"
)

type Kubelet struct {
	NodeName string
	Client *client.Client
	Runtime rt.Runtime
}

func New(nodeName string, c *client.Client, r rt.Runtime) *Kubelet {
	return &Kubelet{
		NodeName: nodeName,
		Client: c,
		Runtime: r,
	}
}

func (k *Kubelet) RegisterNode(ctx context.Context) error {
	n := v1.NewNode(k.NodeName)
	n.Status.Capacity = v1.ResourceList{ "pods": 110,}
	n.Status.Allocatable = v1.ResourceList{ "pods": 110}
	n.Status.Conditions = []apimachinery.Condition{
		{
		Type: "Ready",
		Status: "True",
		Reason: "KubeletReady",
		LastTransitionTime: time.Now(),
		},
	}	
	return k.Client.CreateNode(ctx, n)
}

func (k *Kubelet) Run(ctx context.Context) error {
	if err := k.RegisterNode(ctx); err != nil {
		log.Printf("registering node: %v", err)
	}

	ticker := time.NewTicker(2 * time.Second) // to be replaced with watch
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := k.reconcile(ctx); err != nil {
				log.Printf("reconcile sync: %v", err)
			}
		}
	}
}

func (k *Kubelet) reconcile(ctx context.Context) error {
	// list all pods
	pods, err := k.Client.ListPods(ctx)
	if err != nil {
		return err
	}

	k.syncPods(ctx, pods.Items)
	// k.stopMissingPods(ctx, pods.Items)

	nodeStatus := &v1.NodeStatus{
		Capacity: v1.ResourceList{ "pods": 110},
		Allocatable: v1.ResourceList{ "pods": 110},
		Conditions: []apimachinery.Condition{
			{
				Type: "Ready",
				Status: "True",
				Reason: "KubeletReady",
				LastTransitionTime: time.Now(),
			},
		},
	}
	return k.Client.UpdateNodeStatus(ctx, k.NodeName, nodeStatus)
}


