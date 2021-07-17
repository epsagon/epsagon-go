package epsagonredis

import (
	"context"
	"encoding/json"
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

func wrapClient(client *redis.Client, opt *redis.Options, epsagonCtx context.Context) *redis.Client {
	currentTracer := epsagon.ExtractTracer([]context.Context{epsagonCtx})
	if currentTracer != nil {
		host, port := getClientHostPort(opt)
		client.AddHook(&epsagonHook{
			host:    host,
			port:    port,
			dbIndex: string(opt.DB),
			tracer:  currentTracer,
		})
	}
	return client
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

func (epsHook *epsagonHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	cmdArgs := safeJsonify(cmd.Args())
	return epsHook.before(ctx, cmd.Name(), cmdArgs)
}

func (epsHook *epsagonHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	var errMsg string
	if err := cmd.Err(); err != nil {
		errMsg = err.Error()
	}
	return epsHook.after(ctx, cmd.String(), errMsg)
}

func (epsHook *epsagonHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	cmdArgs := safeJsonify(getPiplineCmdArgs(cmds))
	return epsHook.before(ctx, "Pipeline", cmdArgs)
}

func (epsHook *epsagonHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	var errMsg string
	if errors := getPipelineErrors(cmds); len(errors) > 0 {
		errMsg = safeJsonify(errors)

	}
	response := safeJsonify(getPipelineResponse(cmds))
	return epsHook.after(ctx, response, errMsg)
}

func (epsHook *epsagonHook) before(ctx context.Context, operation, cmdArgs string) (context.Context, error) {
	metadata := getResourceMetadata(epsHook, cmdArgs)
	epsHook.event = createEvent(epsHook, operation, metadata)
	return ctx, nil
}

func (epsHook *epsagonHook) after(ctx context.Context, response, errMsg string) error {
	event := epsHook.event
	if event == nil {
		return nil
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
	return nil
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
	metadata["Command Arguments"] = cmdArgs
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
