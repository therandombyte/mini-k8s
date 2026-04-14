package main

import (
	"context"
	"flag"
	"log"

	"github.com/therandombyte/mini-k8s/pkg/client"
	"github.com/therandombyte/mini-k8s/pkg/kubelet"
)

func main() {
	api := flag.String("api", "http://127.0.0.1:8089", "api server url")
	nodeName := flag.String("node-name", "node1", "node name")
	flag.Parse()

	k := kubelet.New(*nodeName, client.New(*api))
	log.Printf("mk-kubelet starting for node %s", *nodeName)

	if err := k.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
