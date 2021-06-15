
package epsagonmongo

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"

	"github.com/epsagon/epsagon-go/epsagon"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)


// MongoCollectionWrapper is Epsagon's wrapper for mongo.Collection
type MongoCollectionWrapper struct {
	collection		*mongo.Collection
	currentTracer 	tracer.Tracer
}

func WrapMongoCollection(
	collection *mongo.Collection, ctx ...context.Context,
) *MongoCollectionWrapper {
	return &MongoCollectionWrapper{
		collection: collection,
		currentTracer: epsagon.ExtractTracer(ctx),
	}
}


func (coll *MongoCollectionWrapper) Name() string {
	return coll.collection.Name()
}

func (coll *MongoCollectionWrapper) Database() *mongo.Database {
	return coll.collection.Database()
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
	completeMongoEvent(event)

	return response, err
}


func (coll *MongoCollectionWrapper) InsertOne(
	ctx context.Context, document interface{}, opts ...*mongoOptions.InsertOneOptions,
) (*mongo.InsertOneResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.InsertOne(
		ctx,
		document,
		opts...
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "document", document, coll.currentTracer.GetConfig().Debug)
	marshalToMetadata(event.Resource.Metadata, "response", response, coll.currentTracer.GetConfig().Debug)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) InsertMany(
	ctx context.Context, documents []interface{}, opts ...*mongoOptions.InsertManyOptions,
) (*mongo.InsertManyResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.InsertMany(
		ctx,
		documents,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "documents", documents, coll.currentTracer.GetConfig().Debug)
	marshalToMetadata(event.Resource.Metadata, "response", *response, coll.currentTracer.GetConfig().Debug)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) BulkWrite(
	ctx context.Context, models []mongo.WriteModel, opts ...*mongoOptions.BulkWriteOptions,
) (*mongo.BulkWriteResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.BulkWrite(
		ctx,
		models,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(event.Resource.Metadata, "documents", models, coll.currentTracer.GetConfig().Debug)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) DeleteOne(
	ctx context.Context, filter interface{}, opts ...*mongoOptions.DeleteOptions,
) (*mongo.DeleteResult, error){
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.DeleteOne(
		ctx,
		filter,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(
		event.Resource.Metadata, "params", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "response", *response, coll.currentTracer.GetConfig().Debug,
	)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) DeleteMany(
	ctx context.Context, filter interface{}, opts ...*mongoOptions.DeleteOptions,
) (*mongo.DeleteResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.DeleteMany(
		ctx,
		filter,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(
		event.Resource.Metadata, "params", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "response", response, coll.currentTracer.GetConfig().Debug,
	)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) UpdateOne(
	ctx context.Context, filter interface{}, update interface{}, opts ...*mongoOptions.UpdateOptions,
) (*mongo.UpdateResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.UpdateOne(
		ctx,
		filter,
		update,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(
		event.Resource.Metadata, "filter", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "update_conditions", update, coll.currentTracer.GetConfig().Debug,
	)
	extractStructFields(event.Resource.Metadata, "response", *response)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) UpdateMany(
	ctx context.Context, filter interface{}, update interface{}, opts ...*mongoOptions.UpdateOptions,
) (*mongo.UpdateResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.UpdateMany(
		ctx,
		filter,
		update,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(
		event.Resource.Metadata, "filter", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "update_conditions", update, coll.currentTracer.GetConfig().Debug,
	)
	extractStructFields(event.Resource.Metadata, "response", *response)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) UpdateByID(
	ctx context.Context, id interface{}, update interface{}, opts ...*mongoOptions.UpdateOptions,
) (*mongo.UpdateResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.UpdateByID(
		ctx,
		id,
		update,
		opts...
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(
		event.Resource.Metadata, "id", id, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "update_conditions", update, coll.currentTracer.GetConfig().Debug,
	)
	extractStructFields(event.Resource.Metadata, "response", response)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) ReplaceOne(
	ctx context.Context, filter interface{}, replacement interface{}, opts ...*mongoOptions.ReplaceOptions,
) (*mongo.UpdateResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.ReplaceOne(
		ctx,
		filter,
		replacement,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(
		event.Resource.Metadata, "filter", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "replacement", replacement, coll.currentTracer.GetConfig().Debug,
	)
	extractStructFields(event.Resource.Metadata, "response", *response)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) Aggregate(
	ctx context.Context, pipeline interface{}, opts ...*mongoOptions.AggregateOptions,
) (*mongo.Cursor, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.Aggregate(
		ctx,
		pipeline,
		opts...,
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

	marshalToMetadata(
		event.Resource.Metadata, "params", pipeline, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "response", docs, coll.currentTracer.GetConfig().Debug,
	)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) CountDocuments(
	ctx context.Context, filter interface{}, opts ...*mongoOptions.CountOptions,
) (int64, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.CountDocuments(
		ctx,
		filter,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	marshalToMetadata(
		event.Resource.Metadata, "filter", filter, coll.currentTracer.GetConfig().Debug,
	)
	event.Resource.Metadata["count"] = fmt.Sprintf("%d", response)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) EstimatedDocumentCount(
	ctx context.Context, opts ...*mongoOptions.EstimatedDocumentCountOptions,
) (int64, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.EstimatedDocumentCount(
		ctx,
		opts...,
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

func (coll *MongoCollectionWrapper) Distinct(
	ctx context.Context, fieldName string, filter interface{}, opts ...*mongoOptions.DistinctOptions,
) ([]interface{}, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.Distinct(
		ctx,
		fieldName,
		filter,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	event.Resource.Metadata["field_name"] = fieldName
	marshalToMetadata(
		event.Resource.Metadata, "filter", filter, coll.currentTracer.GetConfig().Debug,
	)
	completeMongoEvent(event)
	return response, err
}

func (coll *MongoCollectionWrapper) Find(
	ctx context.Context, filter interface{}, opts ...*mongoOptions.FindOptions,
) (*mongo.Cursor, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.Find(
		ctx,
		filter,
		opts...,
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

	marshalToMetadata(
		event.Resource.Metadata, "filter", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "documents", docs, coll.currentTracer.GetConfig().Debug,
	)
	completeMongoEvent(event)
	return response, err
}


func (coll *MongoCollectionWrapper) FindOne(
	ctx context.Context, filter interface{}, opts ...*mongoOptions.FindOneOptions,
) *mongo.SingleResult {
	event := startMongoEvent(currentFuncName(), coll)
	response := coll.collection.FindOne(
		ctx,
		filter,
		opts...,
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

	marshalToMetadata(
		event.Resource.Metadata, "params", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "document", document, coll.currentTracer.GetConfig().Debug,
	)
	completeMongoEvent(event)
	return response
}


func (coll *MongoCollectionWrapper) FindOneAndDelete(
	ctx context.Context, filter interface{}, opts ...*mongoOptions.FindOneAndDeleteOptions,
) *mongo.SingleResult {
	event := startMongoEvent(currentFuncName(), coll)
	response := coll.collection.FindOneAndDelete(
		ctx,
		filter,
		opts...,
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

	marshalToMetadata(
		event.Resource.Metadata, "params", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "document", document, coll.currentTracer.GetConfig().Debug,
	)
	completeMongoEvent(event)
	return response
}


func (coll *MongoCollectionWrapper) FindOneAndReplace(
	ctx context.Context, filter interface{}, replacement interface{}, opts ...*mongoOptions.FindOneAndReplaceOptions,
) *mongo.SingleResult {
	event := startMongoEvent(currentFuncName(), coll)
	response := coll.collection.FindOneAndReplace(
		ctx,
		filter,
		replacement,
		opts...,
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

	marshalToMetadata(
		event.Resource.Metadata, "filter", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "replacement", replacement, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "document", document, coll.currentTracer.GetConfig().Debug,
	)
	completeMongoEvent(event)
	return response

}


func (coll *MongoCollectionWrapper) FindOneAndUpdate(
	ctx context.Context, filter interface{}, update interface{}, opts ...*mongoOptions.FindOneAndReplaceOptions,
) *mongo.SingleResult {
	event := startMongoEvent(currentFuncName(), coll)
	response := coll.collection.FindOneAndReplace(
		ctx,
		filter,
		update,
		opts...,
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

	marshalToMetadata(
		event.Resource.Metadata, "filter", filter, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "update", update, coll.currentTracer.GetConfig().Debug,
	)
	marshalToMetadata(
		event.Resource.Metadata, "document", document, coll.currentTracer.GetConfig().Debug,
	)
	completeMongoEvent(event)
	return response
}


func (coll *MongoCollectionWrapper) Drop(ctx context.Context) error {
	event := startMongoEvent(currentFuncName(), coll)
	err := coll.collection.Drop(
		ctx,
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

func (coll *MongoCollectionWrapper) Indexes() mongo.IndexView {
	event := startMongoEvent(currentFuncName(), coll)
	indexView := coll.collection.Indexes()
	completeMongoEvent(event)
	return indexView
}

func (coll *MongoCollectionWrapper) Watch(
	ctx context.Context, pipeline interface{}, opts ...*mongoOptions.ChangeStreamOptions,
) (*mongo.ChangeStream, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.Watch(
		ctx,
		pipeline,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		epsagon.ExtractTracer([]context.Context{}).AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}
	completeMongoEvent(event)
	return response, err
}