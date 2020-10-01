package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/epsagon/epsagon-go/epsagon"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
)

func doTask(ctx context.Context, a int, b string, wg *sync.WaitGroup) (int, error) {
	defer wg.Done()
	log.Printf("inside doTask: b = %s", b)
	client := http.Client{Transport: epsagonhttp.NewTracingTransport(ctx)}
	client.Get("https://epsagon.com/")
	return a + 1, fmt.Errorf("boom")
}

func main() {
	config := epsagon.NewTracerConfig("generic-go-wrapper", "")
	config.Debug = true
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		go epsagon.ConcurrentGoWrapper(config, doTask)(i, "hello", &wg)
	}
	wg.Wait()
	time.Sleep(2 * time.Second)
}
