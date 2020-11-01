package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/epsagon/epsagon-go/epsagon"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", epsagonhttp.WrapHandleFunc(
		epsagon.NewTracerConfig("test-http-mux", ""),
		func(w http.ResponseWriter, req *http.Request) {

			client := http.Client{
				Transport: epsagonhttp.NewTracingTransport(req.Context())}
			resp, err := client.Get("http://example.com")

			if err == nil {
				respBody, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					fmt.Println("First 1000 bytes recieved: ", string(respBody[:1000]))
				}
			}
			io.WriteString(w, "ping\n")
		}),
	)

	http.ListenAndServe(":8080", mux)
}
