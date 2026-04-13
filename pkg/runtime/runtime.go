// the boundary between kubelet and the future runtime
// minimal status model to map runtime status to k8s style pod phase.

package runtime

import "time"

type PodState string

const (
	PodStatePending PodState = "Pending"
	PodStateRunnning PodState = "Running"
	PodStateSucceded PodState = "Succeeded"
	PodStateFailed PodState = "Failed"
)

// kubelet needs enough information to decide what Pod status.phase should be from CRI
// so startime, ExitCode, Message etc wll help
// for now, one pod == one process, later we have to add containerStatuses
type PodStatus struct {
	State PodState
	PID int
	StartedAt *time.Time
	ExitedAt *time.Time
	ExitCode int
	Message string
}

// the CRI decopuling
type Runtime interface {
	EnsurePod()
	StopPod()
	PodStatus()
}
