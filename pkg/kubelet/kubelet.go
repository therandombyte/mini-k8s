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
			if err := k.syncPods(ctx); err != nil {
				log.Printf("kubelet sync: %v", err)
			}
		}
	}
}

func (k *Kubelet) syncPods(ctx context.Context) error {
	pods, err := k.Client.ListPods(ctx)
	if err != nil {
		return err
	}

	for i := range pods.Items {
		pod := pods.Items[i]

		if pod.Spec.NodeName != k.NodeName {
			continue
		}
		if pod.Status.Phase == "Running" {
			continue
		}

		now := time.Now()
		status := &v1.PodStatus{
			Phase: "Running",
			HostIP: k.NodeName,
			PodIP: "10.0.0.1",
			StartTime: &now,
			Conditions: []apimachinery.Condition{
				{
					Type: "Ready",
					Status: "True",
					Reason: "Started",
					LastTransitionTime: now,
				},
			},
		}

		if err := k.Client.UpdatePodStatus(ctx, pod.Metadata.Name, status); err != nil {
			log.Printf("Update pod status %s: %v", pod.Metadata.Name, err)
		}

		nodeStatus := &v1.NodeStatus{
			Capacity: v1.ResourceList{ "pods": 110},
			Allocatable: v1.ResourceList{ "pods": 110},
			Conditions: []apimachinery.Condition{
				{
					Type: "Ready",
					Status: "True",
					Reason: "Heartbeat",
					LastTransitionTime: time.Now(),
				},
			},
		}
		return k.Client.UpdateNodeStatus(ctx, k.NodeName, nodeStatus)
	}
	return nil
}


