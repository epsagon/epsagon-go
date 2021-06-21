

package main

import (
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/example/pawan/inner"
	inner "github.com/epsagon/epsagon-go/example/pawan/inner/original_client"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
)

func main() {

	config := epsagon.NewTracerConfig("app", "")
	config.Debug = true
	epsagon.GoWrapper(
		config,
		callInner,
	)()

}


func callInner() {
	//WORKS
	c := inner.NewClient()
	c.Get("https://github.com", map[string][]string{})
	//c := inner.DefaultClient()
	//
	//client := inner.Client{
	//}
	//
	//client.HTTPClient = c


	// DOESNT WORK
	//_, _ = client.Get("https://github.com", map[string][]string{})
}