package main

import (
	"context"
	"flag"
	"log"

	"github.com/therandombyte/mini-k8s/pkg/client"
	"github.com/therandombyte/mini-k8s/pkg/kubelet"
	rt "github.com/therandombyte/mini-k8s/pkg/runtime"
	ctruntime "github.com/therandombyte/mini-k8s/pkg/runtime/containerd"
	"github.com/therandombyte/mini-k8s/pkg/runtime/process"
)

func main() {
	api             := flag.String("api", "http://127.0.0.1:8089", "api server url")
	nodeName        := flag.String("node-name", "node1", "node name")
	rootDir         := flag.String("root-dir", "/tmp/mini-k8s", "runtime state root dir")
	runtimeName     := flag.String("runtime", "containerd", "runtime backend: process|containerd")
	containerSocket := flag.String("containerd-socket", "/run/containerd/containerd.sock", "containerd socket path")
	flag.Parse()
	
	c := client.New(*api)
	var runtimeImpl rt. Runtime

	switch *runtimeName {
	case "process":
		runtimeImpl = process.New(*rootDir)
	case "containerd":
		r, err := ctruntime.New(*containerSocket)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()
		runtimeImpl = r
	default:
		log.Fatalf("unknown runtime %q", *runtimeName)
	}

	k := kubelet.New(*nodeName, c, runtimeImpl)
	log.Printf("mk-kubelet starting for node %s with runtime %s", *nodeName, *runtimeName)

	if err := k.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
