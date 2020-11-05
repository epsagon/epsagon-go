package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/epsagon/epsagon-go/epsagon"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
)

func doTask(ctx context.Context) {
	client := http.Client{Transport: epsagonhttp.NewTracingTransport(ctx)}
	// This password will be masked in the sent trace:
	decodedJSON, err := json.Marshal(map[string]string{"password": "abcde", "animal": "lion"})
	if err != nil {
		epsagon.Label("animal", "lion", ctx)
		epsagon.TypeError(err, "json decoding error", ctx)
	}
	resp, err := client.Post("http://example.com/upload", "application/json", bytes.NewReader(decodedJSON))
	if err != nil {
		epsagon.Label("animal", "lion", ctx)
		epsagon.TypeError(err, "post", ctx)
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		if err == nil {
			epsagon.TypeError(string(body), "post status code", ctx)
		} else {
			epsagon.TypeError(err, "post status code", ctx)
		}
		epsagon.Label("animal", "lion", ctx)
	}
}

func main() {
	// With Epsagon instrumentation
	config := epsagon.NewTracerConfig("test-ignored-keys", "")
	config.Debug = true
	config.MetadataOnly = false
	config.IgnoredKeys = []string{"password"}
	epsagon.ConcurrentGoWrapper(config, doTask)()
}
