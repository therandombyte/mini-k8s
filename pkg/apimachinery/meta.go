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

// For Clients:
// Clients can say "give me a list of Pods after this version". They will get a PodList and a metadata with ResourceVersion
// or a watch mechanism will start watch based on this ResourceVersion as anchor point
// or a cache can store the last seen based on this version, and skip reprocessing if nothing changed
type ListMeta struct {
	ResourceVersion int64
}

// What: A strucutured approach for clients to know the current status and why
// Why: Instead of re-inventing tons of ad-hoc booleans across the code-base, makes easy for tooling to read and display status
// Type: Ready, Available, Progressing, Failed
// Status: True, False, Unknown
// Reason: For code programmatic checks
// Message: Huma readable details
// LastTransitionTime: when this condition last changed, for debugging and timeouts
type Condition struct {
	Type string
	Status string
	Reason string
	Message string
	LastTransitionTime time.Time
}

// What: API operation status, not for k8s objects. Consistent response to tooling
// Status: Success/Failure
// Code: HTTP Status code
type Status struct {
	Status string
	Message string
	Reason string
	Code int
}

