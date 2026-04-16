// the CRI shim
// uses containerd client
// stores container handles by pod key
// maps task state to runtime state

package containerd

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	v1 "github.com/therandombyte/mini-k8s/pkg/api/v1"
	rt "github.com/therandombyte/mini-k8s/pkg/runtime"
)

// track one running pod (container and task together)
// other details also there, so no need to call containerd again
type entry struct {
	container containerd.Container // for metadata and snapshot
	task containerd.Task // running process in the container
	startedAt time.Time
	exitedAt *time.Time // pass as nil, so can be assigned later at exit
	exitCode uint32
	done bool 
}

// concrete runtime implementation
type Runtime struct {
	mu sync.RWMutex
	client *containerd.Client
	namespace string // containerd namespace
	entries map[string]*entry
}

// connects to containerd's unix socket (/run/containerd/containerd.sock)
// in real k8s, it uses grpc
func New(socketPath string) (*Runtime, error) {
	c, err := containerd.New(socketPath)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		client: c,
		namespace: "k8s.io",
		entries: map[string]*entry{},
	}, nil 
}

func key(namespace, name string) string {
	return namespace + "/" + name 
}

// kubelet uses this via podKeyLister interface to find containers that exist in the runtime
// but no longer exist in the API, so it can stop and delete them
func (r *Runtime) PodKeys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, 0, len(r.entries))
	for k := range r.entries {
		out = append(out, k)
	}
	return out
}

// close the containerd client when kubelet exits
func (r *Runtime) Close() error {
	return r.client.Close()
}

// again, make sure pod is running in containerd runtime
func (r *Runtime) EnsurePod(ctx context.Context, pod *v1.Pod) error {
	if len(pod.Spec.Containers) == 0 {
		return fmt.Errorf("pod %s has no containers", pod.Metadata.Name)
	}

	k := key(pod.Metadata.Namespace, pod.Metadata.Name)

	r.mu.RLock()
	// if already running, just return
	if e, ok := r.entries[k]; ok && e.task != nil && !e.done {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	cntrSpec := pod.Spec.Containers[0]
	if cntrSpec.Image == "" {
		return fmt.Errorf("contianer %s has no image", cntrSpec.Name)
	}
	// wrap the context with containerd ns, which is "k8s.io"
	cctx := namespaces.WithNamespace(ctx, r.namespace)

	// time to pull the image
	image, err := r.client.Pull(cctx, cntrSpec.Image, containerd.WithPullUnpack)
	if err != nil {
		return err
	}
	containerID := sanitizeID(fmt.Sprintf("%s-%s", pod.Metadata.Namespace, pod.Metadata.Name))
	snapshotID := containerID + "-snapshot"

	finalArgs := buildContainerArgs(cntrSpec.Command, cntrSpec.Args)
	// applies oci image config (env, working dir, default entrypoint/CMD)
	specOpts := []oci.SpecOpts{
		oci.WithImageConfig(image),
	}
	// if there are explicit args, then use oci.WithProcessArgs to override process command line
	// this is mirroring containerd usage where you pass SpecOpts into WithNewSpec to build a runtime spec
	if len(finalArgs) > 0 {
		specOpts = append(specOpts, oci.WithProcessArgs(finalArgs...))
	}

	// registers a container object with image and provided built spec
	container, err := r.client.NewContainer(
		cctx,
		containerID,
		containerd.WithImage(image),
		containerd.WithNewSnapshot(snapshotID, image),
		containerd.WithNewSpec(specOpts...),
	)
	if err != nil {
		return err
	}

	// turn the container into runnable task. IO is wired through cio.WithStdio to the parent process' stdio for simplicity
	task, err := container.NewTask(cctx, cio.NewCreator(cio.WithStdio));
	if err != nil {
		_ = container.Delete(cctx, containerd.WithSnapshotCleanup)
		return err
	}

	// wait returns a channel to read later for exit status
	exitStatusC, err := task.Wait(cctx)
	if err != nil {
		_, _ = task.Delete(cctx)
		_ = container.Delete(cctx, containerd.WithSnapshotCleanup)
		return err
	}

	if err := task.Start(cctx); err != nil {
		_, _ = task.Delete(cctx)
		_ = container.Delete(cctx, containerd.WithSnapshotCleanup)
		return err
	}

	// record the entry
	e := &entry{
		container: container,
		task: task,
		startedAt: time.Now(),
	}

	r.mu.Lock()
	r.entries[k] = e
	r.mu.Unlock()

	// watch for exit
	go func() {
		status := <-exitStatusC
		code, _, _ := status.Result()
		now := time.Now()

		r.mu.Lock()
		defer r.mu.Unlock()

		e.done = true
		e.exitCode = code
		e.exitedAt = &now
	}()
		return nil
}

func (r *Runtime) StopPod(ctx context.Context, namespace, name string) error {
	k := key(namespace, name)

	r.mu.RLock()
	e, ok := r.entries[k]
	r.mu.RUnlock()

	if !ok {
		return nil
	}

	cctx := namespaces.WithNamespace(ctx, r.namespace)

	if e.task != nil && !e.done {
		_ = e.task.Kill(cctx, syscall.SIGTERM)
		_, _ = e.task.Delete(cctx)
	}

	if e.container != nil {
		_ = e.container.Delete(cctx, containerd.WithSnapshotCleanup)
	}

	r.mu.Lock()
	delete(r.entries, k)
	r.mu.Unlock()

	return nil
}

// bridge containerd lifecyclr to pod status model
func (r *Runtime) PodStatus(ctx context.Context, namespace, name string) (rt.PodStatus, bool, error) {
	// if pod is unknown, return (empty, false, nil) -> not found
	k := key(namespace, name)

	r.mu.RLock()
	e, ok := r.entries[k]
	r.mu.RUnlock()
	if !ok {
		return rt.PodStatus{}, false, nil
	}

	// if done is set, dont call containerd, use exitcode
	// if exitcode == 0 Success, otherwise failed
	if e.done {
		state := rt.PodStateSucceded
		if e.exitCode != 0 {
			state = rt.PodStateFailed
		}
		return rt.PodStatus{
			State: state,
			PID: int(e.task.Pid()),
			StartedAt: &e.startedAt,
			ExitedAt: e.exitedAt,
			ExitCode: int(e.exitCode),
		}, true, nil
	}

	// otherwise query containerd
	// task.Status returns (Running, Created, Stopped etc)
	cctx := namespaces.WithNamespace(ctx, r.namespace)
	st, err := e.task.Status(cctx)
	if err != nil {
		return rt.PodStatus{}, true, err
	}

	switch st.Status {
	case containerd.Running:
		return rt.PodStatus{
			State:     rt.PodStateRunnning,
			PID:       int(e.task.Pid()),
			StartedAt: &e.startedAt,
		}, true, nil
	case containerd.Created:
		return rt.PodStatus{
			State:   rt.PodStatePending,
			PID:     int(e.task.Pid()),
			Message: "created",
		}, true, nil
	case containerd.Stopped:
		return rt.PodStatus{
			State:   rt.PodStatePending,
			PID:     int(e.task.Pid()),
			Message: string(st.Status),
		}, true, nil
	default:
		return rt.PodStatus{
			State: rt.PodStatePending,
			PID: int(e.task.Pid()),
			Message: string(st.Status),
		}, true, nil
	}
}
