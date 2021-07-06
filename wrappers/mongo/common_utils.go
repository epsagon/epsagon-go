package epsagonmongo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"

	"reflect"
	"runtime"
	"strings"
)

// currentFuncName returns the name of the caller function as a string
// for func Foo() { x := currentFuncName }, x == "Foo"
func currentFuncName() string {
	current := make([]uintptr, 1)
	if level := runtime.Callers(2, current); level == 0 {
		return ""
	}
	caller := runtime.FuncForPC(current[0] - 1)
	if caller == nil {
		return ""
	}
	sysFuncName := caller.Name()
	return partitionByDelimiterAtIndex(sysFuncName, ".", -1)

}

func startMongoEvent(opName string, coll *MongoCollectionWrapper) *protocol.Event {
	defer epsagon.GeneralEpsagonRecover("mongo-driver", currentFuncName(), coll.tracer)
	return &protocol.Event{
		Id:        "mongodb-" + uuid.New().String(),
		Origin:    "mongodb",
		ErrorCode: protocol.ErrorCode_OK,
		StartTime: tracer.GetTimestamp(),
		Resource:  createMongoResource(opName, coll),
	}
}

func completeMongoEvent(currentTracer tracer.Tracer, event *protocol.Event) {
	defer epsagon.GeneralEpsagonRecover("mongo-driver", currentFuncName(), currentTracer)
	event.Duration = tracer.GetTimestamp() - event.StartTime
	currentTracer.AddEvent(event)

}

func createMongoResource(opName string, coll *MongoCollectionWrapper) *protocol.Resource {
	defer epsagon.GeneralEpsagonRecover("mongo-driver", currentFuncName(), coll.tracer)
	return &protocol.Resource{
		Name:      coll.Database().Name() + "." + coll.collection.Name(),
		Type:      "mongodb",
		Operation: opName,
		Metadata:  make(map[string]string),
	}
}

// extractStructFields parses the fields from a struct and adds it to metadata under field name metaField
// attempts to convert to int if possible, else keeps as a string
func extractStructFields(
	metadata map[string]string,
	metaField string,
	s interface{},
) {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	valuesMap := make(map[string]string, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		if val.Field(i).CanInterface() {
			fieldVal := val.Field(i)
			v := fmt.Sprintf("%q", fieldVal)
			if _, err := strconv.Atoi(v); err == nil {
				v = strconv.FormatInt(fieldVal.Int(), 10)
			}
			valuesMap[val.Type().Field(i).Name] = v
		}
	}
	doc, _ := json.Marshal(valuesMap)
	metadata[metaField] = string(doc)
}

// marshalToMetadata marshals any object to JSON and adds it as a string to metadata under metaFielf
func marshalToMetadata(
	metadata map[string]string,
	metaField string,
	s interface{},
	config *tracer.Config,
) {
	docBytes, err := json.Marshal(s)
	if err != nil {
		epsagon.DebugLog(config.Debug, "Could not Marshal JSON", err)
	}
	docString := string(docBytes)
	if docString == "" {
		return
	}
	metadata[metaField] = docString
}

// readCursor accepts a cursor and returns a slice of maps
// each map represents a mongo document
func readCursor(cursor *mongo.Cursor) ([]map[string]string, error) {
	var documents []map[string]string
	err := cursor.All(context.Background(), &documents)
	return documents, err
}

func logOperationFailure(messages ...string) {
	for _, m := range messages {
		epsagon.DebugLog(true, "[MONGO]", m)
	}
}

// partition a string by delimiter and return the partitioned at the index
func partitionByDelimiterAtIndex(original, delimiter string, index int) string {
	s := strings.Split(original, delimiter)
	i := moduloFloor(len(s), index)
	return s[i]

}

// flooring an index by size
// useful for wrapping around negative indices to positive
func moduloFloor(size, index int) int {
	return (index + size) % size
}
