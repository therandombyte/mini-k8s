package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/therandombyte/mini-k8s/pkg/client"
)

func main() {
	api := flag.String("api", "http://127.0.0.1:8089", "api server URL")
	file := flag.String("f", "", "file path for apply")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("usage: mkctl get <pods|nodes|deployments> | mkctl apply -f <file>")
	}

	c := client.New(*api)
	ctx := context.Background()

	switch args[0] {
	case "get":
		if len(args) < 2 {
			log.Fatal("usage: mkctl get <pods|nodes|deplopyments>")
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		switch args[1] {
		case "pods":
			obj, err := c.ListPods(ctx)
			if err != nil {
				log.Fatal(err)
			}
			_ = enc.Encode(obj)
		default:
			log.Fatalf("unknown resource %q", args[1])
		}
	case "apply":
		if *file == "" {
			log.Fatal("usage: mkctl apply -f <file>")
		}
	default:
		log.Fatalf("unknown command %q", args[0])
	}
}
