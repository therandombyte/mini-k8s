package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/therandombyte/mini-k8s/pkg/client"
	"github.com/therandombyte/mini-k8s/pkg/scheduler"
)

func main() {
	api := flag.String("api", "http://127.0.0.1:8089", "api server url")
	flag.Parse()

	ctx := context.Background()
	c := client.New(*api)
	s := scheduler.New()

	log.Printf("mk-scheduler connected to %s", *api)

	for {
		pods, err := c.ListPods(ctx)
		if err != nil {
			log.Printf("list pods: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		nodes, err := c.ListNodes(ctx)
		if err != nil {
			log.Printf("list nodes: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for i := range pods.Items {
			pod := pods.Items[i]
			if pod.Spec.NodeName != "" {
				continue
			}

			node := s.PickNode(&pod, nodes.Items)
			if node == nil {
				continue
			}
			// this is the assignment
			pod.Spec.NodeName = node.Metadata.Name

			if err := c.UpdatePod(ctx, &pod); err != nil {
				log.Printf("bind pod %s: %v", pod.Metadata.Name, err)
				continue
			}
			log.Printf("scheduled pod %s onto node %s", pod.Metadata.Name, node.Metadata.Name)

		}
		time.Sleep(2 * time.Second)
	}

}
