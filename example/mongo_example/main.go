
package main

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	epsagonmongo "github.com/epsagon/epsagon-go/wrappers/mongo"
	"os"
	"time"

	//"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type singleRes struct {
	Value string
	Key string
}

func exemplar() {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI("mongodb://localhost:27017"),
	)

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	err = client.Ping(ctx, readpref.Primary())

	db := epsagonmongo.WrapMongoDatabase(
		client.Database("epsagon"),
	)

	coll := db.Collection("godev")
	fmt.Println(coll)



	coll.InsertOne(
		context.Background(),
		struct {
			name string
		}{ "jon" },
	)


	//collection := client.Database("epsagon").Collection("godev")
	//epsagonColl := epsagonmongo.WrapMongoCollection(
	//	client.Database("epsagon"),
	//	collection,
	//	//epsagon.NewTracerConfig(
	//	//	"mongo-dev",
	//	//	"38a22955-dee3-4991-8db8-afa09fc9cef6",
	//	//),
	//)


	//var res singleRes
	//err = epsagonColl.FindOne(
	//	ctx,
	//	bson.D{{"key", "value"}},
	//).Decode(&res)
	//if err != nil || err == mongo.ErrNoDocuments {
	//	panic(err)
	//}


}


func main() {

	err := os.Setenv("EPSAGON_METADATA", "FALSE")
	if err != nil {
		return
	}
	err = os.Setenv("EPSAGON_COLLECTOR_URL", "https://dev.tc.epsagon.com")
	if err != nil {
		return
	}
	err = os.Setenv("EPSAGON_DEBUG", "TRUE")
	if err != nil {
		return
	}

	config := epsagon.NewTracerConfig(
		"mongo-dev",
		"38a22955-dee3-4991-8db8-afa09fc9cef6",
	)
	config.MetadataOnly = false
	epsagon.GoWrapper(
		config,
		exemplar,
	)()

}