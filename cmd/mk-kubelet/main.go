package main

import (
	"context"
	"flag"
	"log"

	"github.com/therandombyte/mini-k8s/pkg/client"
	"github.com/therandombyte/mini-k8s/pkg/kubelet"
	"github.com/therandombyte/mini-k8s/pkg/runtime/process"
)

func main() {
	api := flag.String("api", "http://127.0.0.1:8089", "api server url")
	nodeName := flag.String("node-name", "node1", "node name")
	rootDir := flag.String("root-dir", "/tmp/mini-k8s", "runtime state root dir")
	flag.Parse()
	
	c := client.New(*api)
	r := process.New(*rootDir)
	k := kubelet.New(*nodeName, c, r)
	log.Printf("mk-kubelet starting for node %s", *nodeName)

	if err := k.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
