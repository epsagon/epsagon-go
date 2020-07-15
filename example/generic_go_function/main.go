package main

import (
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"log"
	"reflect"
)

func doTask(args ...interface{}) []interface{} {
	in := reflect.ValueOf(args)
	a := in.Index(0).Elem().Int()
	b := in.Index(1).Elem().String()
	log.Printf("inside doTask: b = %s", b)
	result := []interface{}{a + 1, fmt.Errorf("boom")}
	return result
}

func main() {
	log.Println("enter main")
	config := epsagon.NewTracerConfig("generic-go-wrapper", "")
	config.Debug = true
	response0 := doTask(3, "world")
	res := reflect.ValueOf(response0).Index(0).Elem().Int()
	log.Printf("First result is %d", res)
	responseValue := epsagon.GoWrapper(config, doTask)(5, "hello")
	response := reflect.ValueOf(responseValue)
	res = response.Index(0).Elem().Int()
	errInterface := response.Index(1).Interface()
	if errInterface == nil {
		log.Printf("Result was %d", res)
	} else {
		err := response.Index(1).Interface().(error)
		log.Printf("error was: %v", err)
	}
	// lambda.Start(epsagon.WrapLambdaHandler(config, myHandler))
}
