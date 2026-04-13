package main

import (
	"context"
	"log"

	"github.com/therandombyte/mini-k8s/pkg/apiserver"
	"github.com/therandombyte/mini-k8s/pkg/config"
	"github.com/therandombyte/mini-k8s/pkg/store/memory"
)

func main() {
	cfg := config.DefaultAPIConfig()
	st := memory.New()
	srv := apiserver.New(cfg, st)

	log.Printf("starting mk-apiserver on %s", cfg.Address)

	if err := srv.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
