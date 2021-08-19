
package main

import (
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/net/http"
	"net/http"
)


func SetEpsagonConfig() *epsagon.Config {
	appName := "simple-error-go"
	token := "38a22955-dee3-4991-8db8-afa09fc9cef6"
	config := epsagon.NewTracerConfig(appName, token)
	config.Debug = true
	config.MetadataOnly = false
	config.SendTimeout = "10s"
	config.CollectorURL = "https://dev.tc.epsagon.com"

	return config
}


func handler(res http.ResponseWriter, req *http.Request) {
	println("/test pinged")
	epsagon.Error("Unknown timezone", req.Context())
	res.Write([]byte("Pong.\n"))
}

func main() {
	config := SetEpsagonConfig()
	serveMux := http.NewServeMux()
	serveMux.HandleFunc(
		"/test",
		epsagonhttp.WrapHandleFunc(config, handler))
	server := http.Server{
		Addr: "localhost:8082",
		Handler: serveMux,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
