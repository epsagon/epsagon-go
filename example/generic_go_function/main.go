package main

import (
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
)

func doTask(a int, b string) (int, error) {
	log.Printf("inside doTask: b = %s", b)
	return a + 1, fmt.Errorf("boom")
}

func main() {
	// Normal call
	res, err := doTask(3, "world")
	if err != nil {
		log.Printf("First result is %d", res)
	} else {
		log.Printf("error was: %v", err)
	}

	// With Epsagon instrumentation
	config := epsagon.NewTracerConfig("generic-go-wrapper", "")
	config.Debug = true
	response := epsagon.GoWrapper(config, doTask)(5, "hello")
	res2 := response[0].Int()
	errInterface := response[1].Interface()
	if errInterface == nil {
		log.Printf("Result was %d", res2)
	} else {
		err := errInterface.(error)
		log.Printf("error was: %v", err)
	}
}
