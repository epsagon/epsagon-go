package main

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/net/http"
	"log"
	"net/http"
)

func doTask(ctx context.Context, a int, b string) (int, error) {
	log.Printf("inside doTask: b = %s", b)
	client := epsagonhttp.Wrap(http.Client{}, ctx)
	client.Get("https://epsagon.com/")
	return a + 1, fmt.Errorf("boom")
}

func main() {
	config := epsagon.NewTracerConfig("generic-go-wrapper", "")
	config.Debug = true
	for i := 0; i < 5; i++ {
		go epsagon.ConcurrentGoWrapper(config, doTask)(i, "hello")
	}
}
