package main

import (
	"context"
	"io"
	"net/http"

	"github.com/epsagon/epsagon-go/epsagon"
	epsagonhttp "github.com/epsagon/epsagon-go/wrappers/net/http"
	epsagonredis "github.com/epsagon/epsagon-go/wrappers/redis"
	"github.com/go-redis/redis/v8"
)

func main() {
	config := epsagon.NewTracerConfig(
		"redis-wrapper-test", "",
	)
	config.MetadataOnly = false

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", epsagonhttp.WrapHandleFunc(
		config,
		func(w http.ResponseWriter, req *http.Request) {
			// initialize the redis client as usual, but also make sure
			// to pass in the epsagon tracer context
			rdb := epsagonredis.NewClient(&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			}, req.Context())

			value, _ := rdb.Get(context.Background(), "mykey").Result()

			io.WriteString(w, value)
		}),
	)

	http.ListenAndServe(":8080", mux)
}
