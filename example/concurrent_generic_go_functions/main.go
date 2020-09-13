package main

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/net/http"
	"log"
	"net/http"
	"sync"
	"time"
)

func doTask(ctx context.Context, a int, b string, wg *sync.WaitGroup) (int, error) {
	defer wg.Done()
	log.Printf("inside doTask: b = %s", b)
	client := epsagonhttp.Wrap(http.Client{}, ctx)
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
