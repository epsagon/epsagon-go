
package mongoepsagon

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/tracer"
	//"github.com/golang/protobuf/ptypes/any"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoCollectionWrapper struct {
	Database *mongo.Database
	Collection *mongo.Collection
	Config *epsagon.Config
}

type MongoDatabaseWrapper struct {
	Database *mongo.Database
	Config *epsagon.Config
}

func WrapMongoCollection(database *mongo.Database, collection *mongo.Collection, config *epsagon.Config) *MongoCollectionWrapper {
	return &MongoCollectionWrapper{
		Database: database,
		Collection: collection,
		Config: config,
	}
}

func wrapMongoDatabase(database *mongo.Database, config *epsagon.Config) *MongoDatabaseWrapper {
	return &MongoDatabaseWrapper{
		Database: database,
		Config: config,
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
		Collection: db.Database.Collection(name, opts...),
		Config: db.Config,
	}
}


func (coll *MongoCollectionWrapper) FindOne(
		ctx context.Context,
		filter interface{},
		opts ...*options.FindOneOptions) *mongo.SingleResult {

	findOneEvent := func() *mongo.SingleResult{
		res := coll.Collection.FindOne(
			ctx,
			filter,
			opts...,
		)

		fmt.Println(res)
		return res
	}


	return wrapEvent(findOneEvent, coll.Config)
}

func wrapEvent(
	handler func() *mongo.SingleResult,
	config *epsagon.Config,
	//opts ...any.Any,
) *mongo.SingleResult{
	eventTracer := tracer.CreateTracer(&config.Config)
	eventTracer.Start()
	defer eventTracer.SendStopSignal()

	return handler()


}
