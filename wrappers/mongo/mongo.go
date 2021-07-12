package epsagonmongo

import (
	"context"
	"fmt"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/tracer"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

// MongoCollectionWrapper is Epsagon's wrapper for mongo.Collection
type MongoCollectionWrapper struct {
	collection *mongo.Collection
	tracer     tracer.Tracer
}

func WrapMongoCollection(
	collection *mongo.Collection, ctx ...context.Context,
) *MongoCollectionWrapper {
	return &MongoCollectionWrapper{
		collection: collection,
		tracer:     epsagon.ExtractTracer(ctx),
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}
	if event != nil {
		completeMongoEvent(coll.tracer, event)
	}
	return response, err
}

func (coll *MongoCollectionWrapper) InsertOne(
	ctx context.Context, document interface{}, opts ...*mongoOptions.InsertOneOptions,
) (*mongo.InsertOneResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	fmt.Println("EVENT::::")
	fmt.Println(event)
	response, err := coll.collection.InsertOne(
		ctx,
		document,
		opts...,
	)

	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "document", document, config)
			marshalToMetadata(event.Resource.Metadata, "response", response, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "documents", documents, config)
			marshalToMetadata(event.Resource.Metadata, "response", *response, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "documents", models, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
	return response, err
}

func (coll *MongoCollectionWrapper) DeleteOne(
	ctx context.Context, filter interface{}, opts ...*mongoOptions.DeleteOptions,
) (*mongo.DeleteResult, error) {
	event := startMongoEvent(currentFuncName(), coll)
	response, err := coll.collection.DeleteOne(
		ctx,
		filter,
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "params", filter, config)
			marshalToMetadata(event.Resource.Metadata, "response", *response, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "params", filter, config)
			marshalToMetadata(event.Resource.Metadata, "response", response, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "filter", filter, config)
			marshalToMetadata(event.Resource.Metadata, "update_conditions", update, config)
			extractStructFields(event.Resource.Metadata, "response", *response)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "filter", filter, config)
			marshalToMetadata(event.Resource.Metadata, "update_conditions", update, config)
			extractStructFields(event.Resource.Metadata, "response", *response)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		opts...,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "id", id, config)
			marshalToMetadata(event.Resource.Metadata, "update_conditions", update, config)
			extractStructFields(event.Resource.Metadata, "response", response)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "filter", filter, config)
			marshalToMetadata(event.Resource.Metadata, "replacement", replacement, config)
			extractStructFields(event.Resource.Metadata, "response", *response)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	docs, err := readCursor(response)
	if err != nil {
		logOperationFailure("Could not complete readCursor", err.Error())
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "params", pipeline, config)
			marshalToMetadata(event.Resource.Metadata, "response", docs, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "filter", filter, config)
		}
		event.Resource.Metadata["count"] = fmt.Sprintf("%d", response)
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			event.Resource.Metadata["estimated_count"] = fmt.Sprintf("%d", response)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		event.Resource.Metadata["field_name"] = fieldName
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "filter", filter, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	docs, err := readCursor(response)
	if err != nil {
		logOperationFailure("Could not complete readCursor", err.Error())
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "filter", filter, config)
			marshalToMetadata(event.Resource.Metadata, "documents", docs, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
	response.Decode(&document)

	if err := response.Err(); err != nil {
		logOperationFailure("Could not complete Decode SingleResult", err.Error())
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "params", filter, config)
			marshalToMetadata(event.Resource.Metadata, "document", document, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
	response.Decode(&document)
	if err := response.Err(); err != nil {
		logOperationFailure("Could not complete Decode SingleResult", err.Error())
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "params", filter, config)
			marshalToMetadata(event.Resource.Metadata, "document", document, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
	response.Decode(&document)

	if err := response.Err(); err != nil {
		logOperationFailure("Could not complete Decode SingleResult", err.Error())
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}

	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "filter", filter, config)
			marshalToMetadata(event.Resource.Metadata, "replacement", replacement, config)
			marshalToMetadata(event.Resource.Metadata, "document", document, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
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
	response.Decode(&document)
	if err := response.Err(); err != nil {
		logOperationFailure("Could not complete Decode SingleResult", err.Error())
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}
	if event != nil {
		if config := coll.tracer.GetConfig(); !config.MetadataOnly {
			marshalToMetadata(event.Resource.Metadata, "filter", filter, config)
			marshalToMetadata(event.Resource.Metadata, "update", update, config)
			marshalToMetadata(event.Resource.Metadata, "document", document, config)
		}
		completeMongoEvent(coll.tracer, event)
	}
	return response
}

func (coll *MongoCollectionWrapper) Drop(ctx context.Context) error {
	event := startMongoEvent(currentFuncName(), coll)
	err := coll.collection.Drop(
		ctx,
	)
	if err != nil {
		logOperationFailure(fmt.Sprintf("Could not complete %s", currentFuncName()), err.Error())
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}
	if event != nil {
		completeMongoEvent(coll.tracer, event)
	}
	return err
}

func (coll *MongoCollectionWrapper) Indexes() mongo.IndexView {
	event := startMongoEvent(currentFuncName(), coll)
	indexView := coll.collection.Indexes()
	if event != nil {
		completeMongoEvent(coll.tracer, event)
	}
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
		coll.tracer.AddExceptionTypeAndMessage(
			"mongo-driver",
			err.Error(),
		)
	}
	if event != nil {
		completeMongoEvent(coll.tracer, event)
	}
	return response, err
}
