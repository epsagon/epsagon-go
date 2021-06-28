package epsagongrpc

import (
	"context"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
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


func postGRPCRunner(handlerWrapper *epsagon.GenericWrapper) {
	runner := handlerWrapper.GetRunnerEvent()
	if runner != nil {
		runner.Resource.Type = "grpc_server"
	}
}


func extractGRPCClientRequest(Resource *protocol.Resource, method string, req interface{}, target string) {
	methodParts := strings.SplitN(method, "/", 2)

	u := getURL(method, target)

	if len(methodParts) == 2 {
		Resource.Metadata["rpc.service"] = methodParts[0]
	}

	Resource.Name = u.Host
	Resource.Metadata["net.peer.name"] = u.Hostname()
	Resource.Metadata["rpc.method"] = method
	Resource.Metadata["net.port"] = u.Port()

	reqJson, ok := jsoniter.MarshalToString(req)
	if ok == nil {
		Resource.Metadata["grpc.request.body"] = reqJson
	}
}


func extractGRPCServerRequest(Resource *protocol.Resource, ctx context.Context, fullMethod string, req interface{}) {
	method := strings.TrimPrefix(fullMethod, "/")
	methodParts := strings.SplitN(method, "/", 2)

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
	u := getURL(method, target)

	if len(methodParts) == 2 {
		Resource.Metadata["rpc.service"] = methodParts[0]
	}

	Resource.Name = u.Host
	Resource.Metadata["net.peer.name"] = u.Hostname()
	Resource.Metadata["rpc.method"] = method
	Resource.Metadata["net.port"] = u.Port()

	reqJson, ok := jsoniter.MarshalToString(req)
	if ok == nil {
		Resource.Metadata["grpc.request.body"] = reqJson
	}
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

