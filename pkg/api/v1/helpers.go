// constructors for controllers and kubelet to create resources

package v1

import (
	"fmt"
	"time"

	"github.com/therandombyte/mini-k8s/pkg/apimachinery"
)

func NewNode(name string) *Node {
	return &Node{
		TypeMeta: apimachinery.TypeMeta{
			APIVersion: "mini-k8s/v1",
			Kind: "Node",
		},
		Metadata: apimachinery.ObjectMeta{
			Name: name,
			CreationTimestamp: time.Now(),
		},
	}
}

func NewPod(name, namespace string) *Pod {
	return &Pod{
		TypeMeta: apimachinery.TypeMeta{
			APIVersion: "mini-k8s/v1",
			Kind: "Pod",
		},
		Metadata: apimachinery.ObjectMeta{
			Name: name,
			Namespace: namespace,
			CreationTimestamp: time.Now(),
		},
		Status: PodStatus{
			Phase: "Pending",
		},
	}
}

func NewDeployment(name, namespace string) *Deployment {
	return &Deployment{
		TypeMeta: apimachinery.TypeMeta{
			APIVersion: "mini-k8s/v1",
			Kind: "Deployment",
		},
		Metadata: apimachinery.ObjectMeta{
			Name: name,
			Namespace: namespace,
			CreationTimestamp: time.Now(),
		},
	}
}

// Why: invoked by deployment controller to create pod from deployment spec
func PodFromTemplate(dep *Deployment, index int) *Pod {
	// new pods should be unique
	name := fmt.Sprintf("%s-%d", dep.Metadata.Name, index)
	return &Pod{
		TypeMeta: apimachinery.TypeMeta{
			APIVersion: "mini-k8s/v1",
			Kind: "Pod",
		},
		Metadata: apimachinery.ObjectMeta{
			Name: name,
			Namespace: dep.Metadata.Namespace,  // pods to live in same ns as dep
			Labels: dep.Spec.Template.Metadata.Labels,  // copy over labels
			CreationTimestamp: time.Now(),
		},
		Spec: dep.Spec.Template.Spec,
		Status: PodStatus{Phase: "Pending"},
	}
}
