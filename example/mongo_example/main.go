
package main

import (
	"context"
	"fmt"
	"github.com/epsagon/epsagon-go/epsagon"
	epsagonmongo "github.com/epsagon/epsagon-go/wrappers/mongo"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)


func dbAPI() {

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

	db := client.Database("DB")

	// WRAP THE MONGO COLLECTION with WrapMongoCollection()
	coll := epsagonmongo.WrapMongoCollection(
		db.Collection("COLL"),
	)

	type doc struct {
		Name string
	}
	var res interface{}



	fmt.Println("##InsertOne")
	_, err = coll.InsertOne(
		context.Background(),
		doc{Name: "bon"},
	)
	if err != nil  {
		panic(err)
	}


	fmt.Println("##InsertMany")
	_, err = coll.InsertMany(
		context.Background(),
		[]interface{}{
			bson.D{
				{Key: "name", Value: "hello"},
				{Key: "age", Value: "33"},
			},
			bson.D{
				{Key: "name", Value: "world"},
				{Key: "age", Value: "44"},
			},
		},
	)
	if err != nil  {
		panic(err)
	}



	fmt.Println("##FindOne")
	coll.FindOne(
		context.Background(),
		bson.D{{Key: "name", Value: "bon"}},
	)


	fmt.Println("##Find")
	coll.Find(context.Background(), bson.M{})


	fmt.Println("##Aggregate")
	res, err = coll.Aggregate(
		context.Background(),
		mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.D{{Key: "name", Value: "bon"}}}},
		},
	)
	if err != nil || err == mongo.ErrNoDocuments {
		panic(err)
	}


	fmt.Println("##CountDocuments")
	res, err = coll.CountDocuments(
		context.Background(),
		bson.D{{Key: "name", Value: "bon"}},
	)
	fmt.Println(res)
	if err != nil || err == mongo.ErrNoDocuments {
		panic(err)
	}

	fmt.Println("##DeleteOne")
	res, err = coll.DeleteOne(
		context.Background(),
		bson.D{{Key: "name", Value: "bon"}},
	)
	fmt.Println(res)
	if err != nil || err == mongo.ErrNoDocuments {
		panic(err)
	}

	fmt.Println("##UpdateOne")
	res, err = coll.UpdateOne(
		context.Background(),
		bson.D{{Key: "name", Value: "bon"}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: "son"}}}},
	)
	fmt.Println(res)
	if err != nil || err == mongo.ErrNoDocuments {
		panic(err)
	}
}


func main() {

	config := epsagon.NewTracerConfig(
		"skate-shop",
		"token",
	)
	config.MetadataOnly = false
	config.Debug = true

	epsagon.GoWrapper(
		config,
		dbAPI,
	)()

}