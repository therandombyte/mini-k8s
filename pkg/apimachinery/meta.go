// Contains the metadata and condition types for other resources to reuse
// Resources or objects in k8s carry standard metadata and spec for desired state
// and status for observed state

package apimachinery

import "time"

// What: Type identity of an object (its apiversion apps/v1 and its kind, Pod, Node etc)
// Why: This will be embeded into each object like Pod to self-describe and also for API server and calling clients to know
type TypeMeta struct {
	APIVersion string
	Kind string
}

// What: Metadata of an object
// Why: Controllers depend heavily on labels and generation to drive reconciliation
type ObjectMeta struct {
	Name string
	Namespace string
	UID string
	ResourceVersion int64
	Generation int64
	Labels map[string]string
	CreationTimestamp time.Time
}

// NOTE: above structs could have been merged, but is kept separate for clarity and reuse
// some objects like PodList do not have ObjectMeta
// some clients are looking for TypeMeta only (like admissions)



