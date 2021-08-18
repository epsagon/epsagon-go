
package main

import (
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/wrappers/net/http"
	"net/http"
)

func reporter() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			epsagon.Error("Panic in PendingReservationsHandler")
		}
	}()
	panic("Unknown Panic")
}

func SetEpsagonConfig() *epsagon.Config {
	appName := "simple-error-go"
	token := ""
	config := epsagon.NewTracerConfig(appName, token)
	config.Debug = true
	config.MetadataOnly = false
	config.SendTimeout = "10s"

	return config
}


func handler(res http.ResponseWriter, req *http.Request) {
	println("/test pinged")
	epsagon.Error("Unknown timezone", req.Context())
	//panic("this is panic")
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
