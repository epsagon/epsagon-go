package epsagonmongo

import (
	"context"
	"testing"

	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoWrapper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mongo Driver Test Suite")
}


var _ = Describe("mongo_wrapper", func() {
	Describe("CollectionWrapper", func() {
		var (
			events          	[]*protocol.Event
			exceptions      	[]*protocol.Exception
			wrapper         	*MongoCollectionWrapper
			//called 				 .bool
			testContext			context.Context
			testDatabaseName	string
			testCollectionName	string
			//testDocument		interface{}
		)
		BeforeEach(func() {
			//called =  .false
			config := &epsagon.Config{Config: tracer.Config{
				Disable:  true,
				TestMode: true,
			}}
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
			tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
				Events:     &events,
				Exceptions: &exceptions,
				Labels:     make(map[string]interface{}),
				Config:     &config.Config,
			}


			testCollectionName = "collectionName"
			testDatabaseName = "databaseName"
			testContext = context.Background()
			client, _ := mongo.Connect(testContext, options.Client().ApplyURI("mongodb://localhost:27017"))
			wrapper = &MongoCollectionWrapper{
				database: client.Database(testDatabaseName),
				collection: client.Database(testDatabaseName).Collection(testCollectionName),
			}
		})
		Context("Writing DB", func() {
			It("calls InsertOne", func() {
				_, err := wrapper.InsertOne(context.Background(), struct {
					Name string
				}{"TestName"})
				Expect(err).To(BeNil())
			})
			It("calls InsertMany", func() {
				_, err := wrapper.InsertMany(context.Background(), []interface{}{
					bson.D{
						{"name", "hello"},
						{"age", "33"},
					},
					bson.D{
						{"name", "world"},
						{"age", "44"},
					},
				})
				Expect(err).To(BeNil())
			})
		})
		Context("Reading DB", func() {
			It("calls FindOne", func() {
				wrapper.FindOne(
					context.Background(),
					bson.D{{"name", "helloworld"}},
				)
			})
			It("calls Find", func() {
				wrapper.Find(
					context.Background(),
					bson.M{},
					)
			})
		})
	})
})
