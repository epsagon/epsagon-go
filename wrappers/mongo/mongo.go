
package epsagonmongo

import (
	"context"
	"fmt"
	"reflect"

	"github.com/epsagon/epsagon-go/epsagon"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)


type MongoDatabaseWrapper struct {
	database *mongo.Database
}

type MongoCollectionWrapper struct {
	database *mongo.Database
	collection *mongo.Collection
}

func WrapMongoCollection(collection *mongo.Collection) *MongoCollectionWrapper {
	return &MongoCollectionWrapper{
		database: collection.Database(),
		collection: collection,
	}
}

func WrapMongoDatabase(database *mongo.Database) *MongoDatabaseWrapper {
	return &MongoDatabaseWrapper{
		database: database,
	}
}

func (db *MongoDatabaseWrapper) Client() *mongo.Client {
	return db.database.Client()
}

func (db *MongoDatabaseWrapper) Name() string {
	return db.database.Name()
}

func (db *MongoDatabaseWrapper) Collection(name string, opts ...*mongoOptions.CollectionOptions) *MongoCollectionWrapper {
	return &MongoCollectionWrapper{
		database: db.database,
		collection: db.database.Collection(name, opts...),
	}
}

func (coll *MongoCollectionWrapper) Database(args ...interface{}) *MongoDatabaseWrapper {
	return &MongoDatabaseWrapper{
		database:	coll.database,
	}
}

func (coll *MongoCollectionWrapper) Clone(opts ...*mongoOptions.CollectionOptions) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.Clone(
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}
	marshalToMetadata(event.Resource.Metadata, "params", opts)
	completeMongoEvent(event)

	return response, err
}


func (coll *MongoCollectionWrapper) InsertOne(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.InsertOne(
		args[0].(context.Context),
		args[1],
		//args[2].(*mongoOptions.InsertOneOptions),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "document", args[1])
	marshalToMetadata(event.Resource.Metadata, "response", response)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) InsertMany(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.InsertMany(
		args[0].(context.Context),
		args[1].([]interface{}),
		//args[2:].(*options.InsertManyOptions)
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "documents", args[1])
	marshalToMetadata(event.Resource.Metadata, "response", *response)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) BulkWrite(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.BulkWrite(
		args[0].(context.Context),
		args[1].([]mongo.WriteModel),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "documents", args[1])
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) DeleteOne(args ...interface{}) (interface{}, error){
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.DeleteOne(
		args[0].(context.Context),
		args[1].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "params", args[1])
	marshalToMetadata(event.Resource.Metadata, "response", *response)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) DeleteMany(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.DeleteMany(
		args[0].(context.Context),
		args[1].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "params", args[1])
	marshalToMetadata(event.Resource.Metadata, "response", response)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) UpdateOne(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.UpdateOne(
		args[0].(context.Context),
		args[1].(interface{}),
		args[2].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "filter", args[1])
	marshalToMetadata(event.Resource.Metadata, "update_conditions", args[2])
	extractStructFields(event.Resource.Metadata, "response", *response)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) UpdateMany(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.UpdateMany(
		args[0].(context.Context),
		args[1].(interface{}),
		args[2].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "filter", args[1])
	marshalToMetadata(event.Resource.Metadata, "update_conditions", args[2])
	extractStructFields(event.Resource.Metadata, "response", *response)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) UpdateByID(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.UpdateByID(
		args[0].(context.Context),
		args[1].(interface{}),
		args[2].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "id", args[1])
	marshalToMetadata(event.Resource.Metadata, "update_conditions", args[2])
	extractStructFields(event.Resource.Metadata, "response", response)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) ReplaceOne(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.ReplaceOne(
		args[0].(context.Context),
		args[1].(interface{}),
		args[2].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "filter", args[1])
	marshalToMetadata(event.Resource.Metadata, "replacement", args[2])
	extractStructFields(event.Resource.Metadata, "response", *response)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) Aggregate(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.Aggregate(
		args[0].(context.Context),
		args[1].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	docs, err := readCursor(response)
	if err != nil {
		logOperationFailure("Could not complete readCursor", err.Error())
	}

	marshalToMetadata(event.Resource.Metadata, "params", args[1])
	marshalToMetadata(event.Resource.Metadata, "response", docs)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) CountDocuments(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.CountDocuments(
		args[0].(context.Context),
		args[1].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "filter", args[1])
	event.Resource.Metadata["count"] = fmt.Sprintf("%d", response)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) EstimatedDocumentCount(args ...interface{}) (interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.EstimatedDocumentCount(
		args[0].(context.Context),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	event.Resource.Metadata["estimated_count"] = fmt.Sprintf("%d", response)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) Distinct(args ...interface{}) ([]interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.Distinct(
		args[0].(context.Context),
		args[1].(string),
		args[2].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	event.Resource.Metadata["field_name"] = args[1].(string)
	marshalToMetadata(event.Resource.Metadata, "filter", args[2])
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) Find(args ...interface{}) interface{} {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.Find(
		args[0].(context.Context),
		args[1].(interface{}),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	docs, err := readCursor(response)
	if err != nil {
		logOperationFailure("Could not complete readCursor", err.Error())
	}

	marshalToMetadata(event.Resource.Metadata, "filter", args[1])
	marshalToMetadata(event.Resource.Metadata, "documents", docs)
	completeMongoEvent(event)
	return response
}


func (coll *MongoCollectionWrapper) FindOne(args ...interface{}) interface{} {
	event := startMongoEvent(currentFuncName(), coll)
	response := coll.collection.FindOne(
		args[0].(context.Context),
		args[1].(interface{}),
	)

	var document map[string]string
	decodedResponse := reflect.ValueOf(response).
		MethodByName("Decode").
		Call([]reflect.Value{ reflect.ValueOf(&document) })
	if err := decodedResponse[0]; err.Interface() != nil {
		logOperationFailure("Could not complete Decode SingleResult", err.String())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.String(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "params", args[1])
	marshalToMetadata(event.Resource.Metadata, "document", document)
	completeMongoEvent(event)
	return response
}


func (coll *MongoCollectionWrapper) FindOneAndDelete(args ...interface{}) interface{} {
	event := startMongoEvent(currentFuncName(), coll)
	response := coll.collection.FindOneAndDelete(
		args[0].(context.Context),
		args[1].(interface{}),
	)

	var document map[string]string
	decodedResponse := reflect.ValueOf(response).
		MethodByName("Decode").
		Call([]reflect.Value{ reflect.ValueOf(&document) })
	if err := decodedResponse[0]; err.Interface() != nil {
		logOperationFailure("Could not complete Decode SingleResult", err.String())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.String(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "params", args[1])
	marshalToMetadata(event.Resource.Metadata, "document", document)
	completeMongoEvent(event)
	return response
}


func (coll *MongoCollectionWrapper) FindOneAndReplace(args ...interface{}) interface{} {
	event := startMongoEvent(currentFuncName(), coll)
	response := coll.collection.FindOneAndReplace(
		args[0].(context.Context),
		args[1].(interface{}),
		args[2].(interface{}),
	)
	var document map[string]string
	decodedResponse := reflect.ValueOf(response).
		MethodByName("Decode").
		Call([]reflect.Value{ reflect.ValueOf(&document) })
	if err := decodedResponse[0]; err.Interface() != nil {
		logOperationFailure("Could not complete Decode SingleResult", err.String())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.String(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "filter", args[1])
	marshalToMetadata(event.Resource.Metadata, "replacement", args[2])
	marshalToMetadata(event.Resource.Metadata, "document", document)
	completeMongoEvent(event)
	return response

}


func (coll *MongoCollectionWrapper) FindOneAndUpdate(args ...interface{}) interface{} {
	event := startMongoEvent(currentFuncName(), coll)
	response := coll.collection.FindOneAndReplace(
		args[0].(context.Context),
		args[1].(interface{}),
		args[2].(interface{}),
	)
	var document map[string]string
	decodedResponse := reflect.ValueOf(response).
		MethodByName("Decode").
		Call([]reflect.Value{ reflect.ValueOf(&document) })
	if err := decodedResponse[0]; err.Interface() != nil {
		logOperationFailure("Could not complete Decode SingleResult", err.String())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.String(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "filter", args[1])
	marshalToMetadata(event.Resource.Metadata, "update", args[2])
	marshalToMetadata(event.Resource.Metadata, "document", document)
	completeMongoEvent(event)
	return response
}


func (coll *MongoCollectionWrapper) Drop(args ...interface{}) error {
	event := startMongoEvent(currentFuncName(), coll)
	err := coll.collection.Drop(
		args[0].(context.Context),
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}
	completeMongoEvent(event)
	return err
}
