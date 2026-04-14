// bridge between runtime view of pod and api server podstatus

package kubelet

import (
	"time"

	v1 "github.com/therandombyte/mini-k8s/pkg/api/v1"
	"github.com/therandombyte/mini-k8s/pkg/apimachinery"
	rt "github.com/therandombyte/mini-k8s/pkg/runtime"
)

// converting rt.PodState typed enum to string
func toAPIPhase(state rt.PodState) string {
	switch state{
	case rt.PodStateRunnning:
		return "Running"
	case rt.PodStateSucceded:
		return "Succeeded"
	case rt.PodStateFailed:
		return "Failed"
	default:
		return "Pending"
	}
}

// create API facing PodStatus from runtime rt.PodStatus
func buildPodStatus(nodeName string, rs rt.PodStatus) *v1.PodStatus {
	now := time.Now()

	status := &v1.PodStatus{
		Phase: toAPIPhase(rs.State),
		HostIP: nodeName,
		PodIP: "",
		StartTime: rs.StartedAt,
	}

	ready := "False"
	reason := "Pending"
	if rs.State == rt.PodStateRunnning {
		ready = "True"
		reason = "Started"
	} else if rs.State == rt.PodStateSucceded {
		reason = "Exited"
	} else if rs.State == rt.PodStateFailed {
		reason = "Failed"
	}


	status.Conditions = []apimachinery.Condition {
		{
			Type: "ready",
			Status: ready,
			Reason: reason,
			LastTransitionTime: now,
		},
	}

	return status
}
