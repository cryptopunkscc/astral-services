package main

import (
	"github.com/cryptopunkscc/astral-services/components/run"
	"github.com/cryptopunkscc/astral-services/services/demo"
	"github.com/cryptopunkscc/astral-services/services/ui"
	"log"
	"sync"
)

func main() {
	services := []run.Service{
		ui.Serve,
		demo.Serve,
	}
	var wg sync.WaitGroup
	wg.Add(len(services))
	for _, serve := range services {
		go func(serve run.Service) {
			serve()
			wg.Done()
		}(serve)
	}
	wg.Wait()
	log.Println("finish")
}
