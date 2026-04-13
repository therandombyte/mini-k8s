// The runtime backend.
// For now, execution at os level, using os/exec

package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	v1 "github.com/therandombyte/mini-k8s/pkg/api/v1"
	rt "github.com/therandombyte/mini-k8s/pkg/runtime"
)

// if you are launching a process, there has to be a bookkeeping
// or in-memory registry of Pod -> running/completed process entry
// as its host level execution and not API level, its internal
// Runtime will later translate this into public PodStatus
type procEntry struct {
	cmd *exec.Cmd  // the launced process obj in Go
	startedAt time.Time
	exitedAt time.Time
	exitCode int   // is pod successful/failed
	done bool     // process within container still running?
}

// the obj that will implement Runtime interface
type Runtime struct {
	mu sync.RWMutex
	baseDir string
	procs map[string]*procEntry
}

func New(baseDir string) *Runtime {
	return &Runtime{
		baseDir: baseDir,
		procs: map[string]*procEntry{},
	}
}

func key(namespace, name string) string {
	return namespace + "/" + name
}

// Make sure the pod is running
func (r *Runtime) EnsurePod(ctx context.Context, pod *v1.Pod) error {
	// validate pod has atleast one container
	if (len(pod.Spec.Containers) == 0) {
		return fmt.Errorf("pod %s has no containers", pod.Metadata.Name)
	}

	k := key(pod.Metadata.Namespace, pod.Metadata.Name)

	// core reconciliation
	r.mu.RLock()
	if existing, ok := r.procs[k]; ok && existing.cmd != nil && 
		existing.cmd.Process != nil && !existing.done {
			r.mu.RUnlock()
			return nil
	}
	r.mu.RUnlock()

	// use only the first container
	c := pod.Spec.Containers[0]
	// take commands and args out from spec to form a command
	args := append([]string{}, c.Command...)
	args = append(args, c.Args...)
	if (len(args) == 0) {
		return fmt.Errorf("container %s has no command/args", c.Name)
	}

	// create pod working directory
	// logs will be written here
	podDir := filepath.Join(r.baseDir, pod.Metadata.Namespace, pod.Metadata.Name)
	if err := os.MkdirAll(podDir, 0o755); err != nil {
		return err
	}

	// redirecting stdout and stderr to files, to its available on nodes
	// and can be retireved bu kubectl logs
	stdoutFile, err := os.OpenFile(filepath.Join(podDir, "stdout.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	stderrFile, err := os.OpenFile(filepath.Join(podDir, "stderr.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		_ = stdoutFile.Close()
		return err
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	cmd.Dir = podDir

	// start the real process
	// returns immediately
	if err := cmd.Start(); err != nil {
		_= stdoutFile.Close()
		_=stderrFile.Close()
		return err
	}

	entry := &procEntry{
		cmd: cmd,
		startedAt: time.Now(),
	}

	r.mu.Lock()
	r.procs[k] = entry
	r.mu.Unlock()

	// wait async for exit
	go func() {
		err := cmd.Wait()
		now := time.Now()
		
		r.mu.Lock()
		defer r.mu.Unlock()


		entry.done = true
		entry.exitedAt = now

		if err == nil {
			entry.exitCode = 0
			return
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				entry.exitCode = status.ExitStatus()
				return
			}
		}
		entry.exitCode = 1
	}()

	return nil
}


func (r *Runtime) StopPod(ctx context.Context, namespace, name string) error {
	k := key(namespace, name)
	r.mu.RLock()
	entry, ok := r.procs[k]
	r.mu.RUnlock()
	if !ok || entry.cmd == nil || entry.cmd.Process == nil || entry.done {
		return nil
	}
	return entry.cmd.Process.Kill()
}

func(r *Runtime) PodStatus(ctx context.Context, namespace, name string) (rt.PodStatus, bool, error) {
	k := key(namespace, name)

	r.mu.RLock()
	entry, ok := r.procs[k]
	r.mu.RUnlock()

	if !ok {
		return rt.PodStatus{}, false, nil
	}

	if !entry.done {
		return 
	}
	return nil, false, nil
}
