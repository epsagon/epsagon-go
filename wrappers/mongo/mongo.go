
package mongoepsagon

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	//"time"

	//"github.com/epsagon/epsagon-go/tracer"
	"github.com/google/uuid"

	//"github.com/golang/protobuf/ptypes/any"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


//type MongoCall struct {
//	opName string
//	startTime *time.Time
//	database *MongoDatabaseWrapper
//	collection *MongoCollectionWrapper
//
//}

type MongoDatabaseWrapper struct {
	Database *mongo.Database
	Config *epsagon.Config
}

type MongoCollectionWrapper struct {
	Database *mongo.Database
	Collection *mongo.Collection
	Config *epsagon.Config
}
func WrapMongoCollection(database *mongo.Database, collection *mongo.Collection, config *epsagon.Config) *MongoCollectionWrapper {
	return &MongoCollectionWrapper{
		Database: database,
		Collection: collection,
		Config: config,
	}
}

func WrapMongoDatabase(database *mongo.Database) *MongoDatabaseWrapper {
	return &MongoDatabaseWrapper{
		Database: database,
		//Config: config,
	}
}

func (db *MongoDatabaseWrapper) Client() *mongo.Client {
	return db.Database.Client()
}

func (db *MongoDatabaseWrapper) Name() string {
	return db.Database.Name()
}

func (db *MongoDatabaseWrapper) Collection(name string, opts ...*options.CollectionOptions) *MongoCollectionWrapper {
	return &MongoCollectionWrapper{
		Database: db.Database,
		Collection: db.Database.Collection(name, opts...),
		Config: db.Config,
	}
}

func completeMongoEvent(res *protocol.Resource, startTime float64, args ...context.Context) {
	//traceArgs := make([]context.Context, 1)
	currentTracer := epsagon.ExtractTracer(args)
	//currentTracer.

	endTime := tracer.GetTimestamp()
	event := &protocol.Event{
		Id: "mongodb-" + uuid.New().String(),
		Origin: "mongodb",
		ErrorCode: protocol.ErrorCode_OK,
		Resource: res,
		StartTime: startTime,
	}
	event.Duration = endTime - event.StartTime
	currentTracer.AddEvent(event)

}

func createMongoResource(opName string, coll *MongoCollectionWrapper) *protocol.Resource {
	res := &protocol.Resource{
		Name: coll.Database.Name() + "." + coll.Collection.Name(),
		Type: "mongodb",
		Operation: opName,
		Metadata: make(map[string]string),

	}

	return res
}

func (coll *MongoCollectionWrapper) InsertOne(args ...interface{}) (interface{}, error) {


	if len(args) < 2 {
		return nil, fmt.Errorf("Not enough Args")
	}

	startTime := tracer.GetTimestamp()
	//insert := func() ([]interface{}, error)
	res := createMongoResource("InsertOne", coll)
	//return coll.Collection.InsertOne(args...)

	completeMongoEvent(res, startTime)

	fmt.Println(args[0])
	return nil, nil
}


//func (coll *MongoCollectionWrapper) FindOne(
//		ctx context.Context,
//		filter interface{},
//		opts ...*options.FindOneOptions) *mongo.SingleResult {
//
//	findOne := func() *mongo.SingleResult{
//		res := coll.Collection.FindOne(
//			ctx,
//			filter,
//			opts...,
//		)
//
//		fmt.Println(res)
//		return res
//	}
//
//
//	return wrapCall(findOne, coll.Config)
//}

//func wrapCall(
//	handler func() *mongo.SingleResult,
//	config *epsagon.Config,
//	//opts ...any.Any,
//	args ...context.Context,
//) *mongo.SingleResult {
//	//eventTracer := tracer.CreateTracer(&config.Config)
//	//var arg context.Context
//	currentTracer := epsagon.ExtractTracer(args)
//
//	startTime := tracer.GetTimestamp()
//
//
//	event := protocol.Event{
//		Id:                   uuid.NewString(),
//		StartTime:            startTime,
//		Resource:             nil,
//		Origin:               "mongo",
//		Duration:             0,
//		ErrorCode:            0,
//		Exception:            nil,
//		XXX_NoUnkeyedLiteral: struct{}{},
//		XXX_unrecognized:     nil,
//		XXX_sizecache:        0,
//	}
//
//	fmt.Println("Started Tracer")
//
//	//defer func() {
//	//	currentTracer.SendStopSignal()
//	//	fmt.Println("stopped")
//	//}()
//
//	fmt.Println("adding event")
//	currentTracer.AddEvent(
//		createMongoEvent(
//			"MyColl",
//			"FindOne",
//
//		),
//	)
//	fmt.Println("added event")
//
//
//	return handler()
//
//}

//func createMongoEvent(
//	name, method string,
//	//err error
//	) *protocol.Event {
//
//	return &protocol.Event{
//		Id: "mongo.driver-" + uuid.New().String(),
//		Origin: "mongo.driver",
//		ErrorCode: protocol.ErrorCode_OK,
//		Resource: &protocol.Resource{
//			Name:	name,
//			Type:	"mongo",
//			Operation: method,
//			Metadata: map[string]string{},
//		},
//	}
//}


