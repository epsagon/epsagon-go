package epsagonredis

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"runtime/debug"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type epsagonHook struct {
	host    string
	port    string
	dbIndex string
	event   *protocol.Event
	tracer  tracer.Tracer
}

func NewClient(opt *redis.Options, epsagonCtx context.Context) *redis.Client {
	client := redis.NewClient(opt)
	return wrapClient(client, opt, epsagonCtx)
}

func wrapClient(client *redis.Client, opt *redis.Options, epsagonCtx context.Context) (wrappedClient *redis.Client) {
	wrappedClient = client
	defer func() { recover() }()

	currentTracer := epsagon.ExtractTracer([]context.Context{epsagonCtx})
	if currentTracer != nil {
		host, port := getClientHostPort(opt)
		client.AddHook(&epsagonHook{
			host:    host,
			port:    port,
			dbIndex: fmt.Sprint(opt.DB),
			tracer:  currentTracer,
		})
	}
	return wrappedClient
}

func getClientHostPort(opt *redis.Options) (host, port string) {
	if opt.Network == "unix" {
		return "localhost", opt.Addr
	}
	host, port, err := net.SplitHostPort(opt.Addr)
	if err == nil {
		return host, port
	}
	return opt.Addr, ""
}

func (epsHook *epsagonHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (processCtx context.Context, err error) {
	processCtx, err = ctx, nil
	defer func() { recover() }()

	cmdArgs := safeJsonify(cmd.Args())
	epsHook.before(cmd.Name(), cmdArgs)
	return
}

func (epsHook *epsagonHook) AfterProcess(ctx context.Context, cmd redis.Cmder) (err error) {
	defer func() { recover() }()

	var errMsg string
	if err := cmd.Err(); err != nil {
		errMsg = err.Error()
	}
	epsHook.after(cmd.String(), errMsg)
	return
}

func (epsHook *epsagonHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (processCtx context.Context, err error) {
	processCtx, err = ctx, nil
	defer func() { recover() }()

	cmdArgs := safeJsonify(getPiplineCmdArgs(cmds))
	epsHook.before("Pipeline", cmdArgs)
	return
}

func (epsHook *epsagonHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) (err error) {
	defer func() { recover() }()

	var errMsg string
	if errors := getPipelineErrors(cmds); len(errors) > 0 {
		errMsg = safeJsonify(errors)

	}
	response := safeJsonify(getPipelineResponse(cmds))
	epsHook.after(response, errMsg)
	return
}

func (epsHook *epsagonHook) before(operation, cmdArgs string) {
	metadata := getResourceMetadata(epsHook, cmdArgs)
	epsHook.event = createEvent(epsHook, operation, metadata)
}

func (epsHook *epsagonHook) after(response, errMsg string) {
	event := epsHook.event
	if event == nil {
		return
	}
	if !epsHook.tracer.GetConfig().MetadataOnly {
		event.Resource.Metadata["redis.response"] = response
	}

	eventEndTime := tracer.GetTimestamp()
	event.Duration = eventEndTime - event.StartTime

	if errMsg != "" {
		event.ErrorCode = protocol.ErrorCode_EXCEPTION
		event.Exception = &protocol.Exception{
			Message:   errMsg,
			Traceback: string(debug.Stack()),
			Time:      eventEndTime,
		}
	}

	epsHook.tracer.AddEvent(event)
}

func createEvent(epsHook *epsagonHook, operation string, metadata map[string]string) *protocol.Event {
	return &protocol.Event{
		Id:        "redis-" + uuid.New().String(),
		StartTime: tracer.GetTimestamp(),
		Resource: &protocol.Resource{
			Name:      epsHook.host,
			Type:      "redis",
			Operation: operation,
			Metadata:  metadata,
		},
	}
}

func getResourceMetadata(epsHook *epsagonHook, cmdArgs string) map[string]string {
	metadata := getConnectionMetadata(epsHook)
	if !epsHook.tracer.GetConfig().MetadataOnly {
		metadata["Command Arguments"] = cmdArgs
	}
	return metadata
}

func getConnectionMetadata(epsHook *epsagonHook) map[string]string {
	return map[string]string{
		"Redis Host":     epsHook.host,
		"Redis Port":     epsHook.port,
		"Redis DB Index": epsHook.dbIndex,
	}
}

func getPiplineCmdArgs(cmds []redis.Cmder) []interface{} {
	var cmdArgs []interface{}
	for _, cmd := range cmds {
		cmdArgs = append(cmdArgs, cmd.Args())
	}
	return cmdArgs
}

func getPipelineResponse(cmds []redis.Cmder) []string {
	var responses []string
	for _, cmd := range cmds {
		responses = append(responses, cmd.String())
	}
	return responses
}

func getPipelineErrors(cmds []redis.Cmder) []string {
	var errors []string
	for _, cmd := range cmds {
		if err := cmd.Err(); err != nil {
			errors = append(errors, err.Error())
		}
	}
	return errors
}

func safeJsonify(v interface{}) string {
	encodedValue, err := json.Marshal(v)
	if err == nil {
		return string(encodedValue)
	}
	return ""
}
