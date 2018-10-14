package main

import (
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"log"
	"time"
)

func main() {
	log.Println("vim-go")
	epsagon.CreateTracer(
		"erez-test-go",
		"4e988052-fb55-4c6d-b46f-f17f3dd81b6b",
		"http://dev.tc.epsagon.com",
	)
	defer epsagon.StopTracer()

	event := protocol.Event{
		Id:        "1234-test-event-1",
		StartTime: float64(time.Now().Unix()),
		Resource: &protocol.Resource{
			Name:      "erez-test",
			Type:      "test-type",
			Operation: "test-operation",
			Metadata:  map[string]string{"hello": "world"},
		},
		Origin:    "runner",
		Duration:  float64(1),
		ErrorCode: protocol.ErrorCode_OK,
	}
	epsagon.AddEvent(event)
}
