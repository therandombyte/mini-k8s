// defines the first set of api objects (versioned)
// the contract between api server, controller, kubelet and clients.
// Users: scheduler will need this to bind pods, kubelet to publish node and pod status, controller to reconcile

package v1

import (
	"time"

	"github.com/therandombyte/mini-k8s/pkg/apimachinery"
)

//What: a map of resources and its count ("pods": 110)
// Why: How many pods are running, so i know the node capacity and allocations
// as its for all objects, so we use a flexible map instead of a fixed struct
type ResourceList map[string]int64

// What: Minimal container spec inside the pod
// Why: Smallest mwaningful definition of the workload
// later to be enriched with env, ports, resources, etc
type Container struct {
	Name string
	Image string
	Command []string
	Args []string
}

// What: Desired state for a pod
// Why: What containers/workloads to run (by deployments/api clients), which node this pod is assigned to (by scheduler)
type PodSpec struct {
	NodeName string
	Containers []Container
}

// What: Observed state of a Pod
// Why: This is what Kubelet reports back to API server
// Controller and UIs consume status to know whether system has achieved the desired state expected in spec
type PodStatus struct {
	Phase string
	HostIP string
	PodIP string
	Conditions []apimachinery.Condition
	StartTime *time.Time
}

// What: Full pod object: type + metadata + spec/desired state + status/observed state
// Why: API Server stores and serves, Scheduler reads/writes spec, Kubelet reads spec and updates status
type Pod struct {
	apimachinery.TypeMeta
	Metadata apimachinery.ObjectMeta
	Spec PodSpec
	Status PodStatus
}

// What: List wrapper for pods
// Why: api response for get pods
type PodList struct {
	apimachinery.TypeMeta
	Metadata apimachinery.ListMeta
	Items []Pod
}

// What: A pod blueprint inside a controller
// Why: Use this inside DeploymentSpec to say what Pod should the deployment create
type PodSpecTemplate struct {
	Metadata apimachinery.ObjectMeta
	Spec PodSpec
}

// What: Obseeved state of a node 
// Why: Kubelet updates this to API server
type NodeStatus struct {
	Capacity ResourceList
	Allocatable ResourceList
	Conditions []apimachinery.Condition
}

// What: Desired scheduling properties of a Node
// Why: For now, just one field for scheduler to decide (this field is toggled by kubectl cordon)
type NodeSpec struct {
	Unschedulable bool
}

// Full node object like Pod
// Why: Kubelet will create a new node with NewNode(name), sets capacity and ready condition
// posts it to API server and periodically updates status
type Node struct {
	apimachinery.TypeMeta
	Metadata apimachinery.ObjectMeta
	Spec NodeSpec
	Status NodeStatus
}

// Scheduler will use this to list nodes and then assign pods
type NodeList struct {
	apimachinery.TypeMeta
	Metadata apimachinery.ListMeta
	Items []Node
}


type DeploymentStatus struct {}

type DeploymentSpec struct {}

type Deployment struct {}

type DeploymentList struct {}
