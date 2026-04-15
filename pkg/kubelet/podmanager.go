// Core reconciliation loop: for kubelet to reconcile desired pods to real runtime state

package kubelet

import (
	"context"
	"log"

	v1 "github.com/therandombyte/mini-k8s/pkg/api/v1"
)

//drive runtime from desired state
func (k *Kubelet) syncPods(ctx context.Context, pods []v1.Pod) {
	for _, pod := range pods {
		if pod.Spec.NodeName != k.NodeName {
			continue
		}
		// ensure pod ids running in the runtime
		if err := k.Runtime.EnsurePod(ctx, &pod); err != nil {
			log.Printf("ensure pod %s: %v", pod.Metadata.Name, err)
			continue
		}

		//ask runtime for actual pod status
		rs, ok, err := k.Runtime.PodStatus(ctx, pod.Metadata.Namespace, pod.Metadata.Name)
		if err != nil {
			log.Printf("pod status %s: %v", pod.Metadata.Name, err)
			continue
		}
		// ok tells whether runtime knows about the pod
		if !ok {
			continue
		}

		// map runtime status to API status
		newStatus := buildPodStatus(k.NodeName, rs)

		// avoid redundant status update
		if k.samePodStatus(pod.Status, *newStatus) {
			continue
		}

		//update pod status to the api server
		if err := k.Client.UpdatePodStatus(ctx, pod.Metadata.Name, newStatus); err != nil {
			log.Printf("update pod status %s: %v", pod.Metadata.Name, err)
		}

	}
}

// change detection
func (k *Kubelet) samePodStatus(a, b v1.PodStatus) bool {
	if a.Phase != b.Phase {
		return false
	}

	if len(a.Conditions) != len(b.Conditions) {
		return false
	}

	for i := range a.Conditions {
		if a.Conditions[i].Type != b.Conditions[i].Type ||
			a.Conditions[i].Status != b.Conditions[i].Status ||
			a.Conditions[i].Reason != b.Conditions[i].Reason {
				return false
		}
		
	}
	return true
}

// delete side of reconciliation, or the garbage collector
func (k *Kubelet) stopMissingPods(ctx context.Context, pods []v1.Pod) {
	seen := map[string]bool{}
	// build a set of pod that should exist
	for _ ,pod := range pods {
		if pod.Spec.NodeName != k.NodeName {
			continue
		}
		seen[pod.Metadata.Namespace+"/"+pod.Metadata.Name] = true
	}
	// ask the runtime for the list of pod keys its tracking
	// any pod not in the seen set, stop it
	for _, name := range k.RuntimePodKeys(ctx) {
		if !seen[name] {
			ns, podName := splitRuntimeKey(name)
			if err := k.Runtime.StopPod(ctx, ns, podName); err != nil {
				log.Printf("stop pod %s: %v", name, err)
			}
		}
	}
}

// keys are in the form of ns/name, this will return it as separate strings
func splitRuntimeKey (k string) (string, string) {
	for i := 0; i < len(k); i++ {
		if k[i] == '/' {
			return k[:i], k[i+1:]
		}
	}
	return "", k
}

// any runtime that implements podKeyLister can be queried for its pod keys
type podKeyLister interface {
	PodKeys() []string
}

// check if the runtime implements podKeyLister, if yes return the keys
func (k *Kubelet) RuntimePodKeys (ctx context.Context) []string {
	if l, ok := k.Runtime.(podKeyLister); ok {
		return l.PodKeys()
	}
	return nil
}
