package epsagongrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"net/http"
	"net/url"
	"strings"
)


func createGRPCEvent(origin string, method string, eventID string) *protocol.Event {
	errorcode := protocol.ErrorCode_OK

	return &protocol.Event{
		Id:        eventID + uuid.New().String(),
		Origin:    origin,
		StartTime: tracer.GetTimestamp(),
		ErrorCode: errorcode,
		Resource: &protocol.Resource{
			Type:      "grpc",
			Operation: method,
			Metadata:  map[string]string{},
		},
	}
}


func decoratePostGRPCRunner(handlerWrapper *epsagon.GenericWrapper) {
	runner := handlerWrapper.GetRunnerEvent()
	if runner != nil {
		runner.Resource.Type = "grpc"
	}
}


func decorateGRPCRequest(Resource *protocol.Resource, ctx context.Context, fullMethod string, req interface{}) {
	method := strings.TrimPrefix(fullMethod, "/")

	var hdrs http.Header
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		hdrs = make(http.Header, len(md))
		for k, vs := range md {
			for _, v := range vs {
				hdrs.Add(k, v)
			}
		}
	}

	target := hdrs.Get(":authority")
	url := getURL(method, target)

	if hdrsString, err := json.Marshal(hdrs); err != nil {
		Resource.Metadata["grpc.headers"] = string(hdrsString)
	} else {
		Resource.Metadata["grpc.headers"] = fmt.Sprintf("%+v", hdrs)
	}

	Resource.Name = url.Host
	Resource.Metadata["grpc.host"] = url.Hostname()
	Resource.Metadata["grpc.url"] = url.String()
	Resource.Metadata["grpc.path"] = url.Path
	Resource.Metadata["grpc.port"] = url.Port()
	Resource.Metadata["grpc.request.body"] = fmt.Sprintf("%+v" , req)
}


func getURL(method string, target string) *url.URL{
	var host string

	if strings.HasPrefix(target, "unix:") {
		host = "localhost"
	} else {
		host = strings.TrimPrefix(target, "dns:///")
	}
	return &url.URL{
		Scheme: "grpc",
		Host:   host,
		Path:   method,
	}
}