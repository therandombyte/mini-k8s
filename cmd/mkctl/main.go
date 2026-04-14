package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	v1 "github.com/therandombyte/mini-k8s/pkg/api/v1"
	"github.com/therandombyte/mini-k8s/pkg/client"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("usage: mkctl get <pods|nodes|deplopyments>")
	}
	api := "http://127.0.0.1:8089"
	
	c := client.New(api)
	ctx := context.Background()

	switch os.Args[1] {
	case "get":
		getCmd := flag.NewFlagSet("get", flag.ExitOnError)
		getCmd.Parse(os.Args[2:])
		args := getCmd.Args()

		if len(args) < 1 {
			log.Fatalf("usage: mkctl get <pods|nodes|deployments>")
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		switch os.Args[2] {
		case "pods":
			obj, err := c.ListPods(ctx)
			if err != nil {
				log.Fatal(err)
			}
			_ = enc.Encode(obj)
		case "nodes":
			obj, err := c.ListNodes(ctx)
			if err != nil {
				log.Fatal(err)
			}
			_ = enc.Encode(obj)
		default:
			log.Fatalf("unknown resource %q", os.Args[2])
		}
	case "apply":
		applyCmd := flag.NewFlagSet("apply", flag.ExitOnError)
		file := applyCmd.String("f", "", "file path for apply")
		applyCmd.Parse(os.Args[2:])

		if *file == "" {
			log.Fatal("usage: mkctl apply -f <file>")
		}		
		
		// time to read a file
		b, err := os.ReadFile(*file)
		if err != nil {
			log.Fatal(err)
		}

		// get the kind from file and put it here
		var tm struct {
			Kind string
		}

		if err := json.Unmarshal(b, &tm); err != nil {
			log.Fatal(err)
		}

		switch tm.Kind {
		case "pod":
			// declare and shove the incoming json using Unmarshal
			var obj v1.Pod
			if err := json.Unmarshal(b, &obj); err != nil {
				log.Fatal(err)
			}
			if err := c.CreatePod(ctx, &obj);err != nil {
				log.Fatal(err)
			}
		case "nodes":
			var obj v1.Node
			if err := json.Unmarshal(b, &obj); err != nil {
				log.Fatal(err)
			}
			if err := c.CreateNode(ctx, &obj); err != nil {
				log.Fatal(err)
			}
		default:
			log.Fatalf("unsupported kind %q", tm.Kind)
		}
		fmt.Println("applied")
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}
