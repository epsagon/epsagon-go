package epsagonmongo

import (
	"context"
	"testing"
	"time"

	"github.com/benweissmann/memongo"
	"github.com/epsagon/epsagon-go/epsagon"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoServerURI string
var docsInserted int64 = 0

func TestMongoWrapper(t *testing.T) {
	mongoServer, err := memongo.Start("4.2.0")
	if err != nil {
		t.Fatal(err)
	}
	defer mongoServer.Stop()
	mongoServerURI = mongoServer.URI()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Mongo Driver Test Suite")
}


var _ = Describe("mongo_wrapper", func() {
	Describe("CollectionWrapper", func() {
		var (
			events          	[]*protocol.Event
			exceptions      	[]*protocol.Exception
			wrapper         	*MongoCollectionWrapper
			testContext			context.Context
			testDatabaseName	string
			testCollectionName	string
			cancel				func()
		)
		BeforeEach(func() {
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
			testContext, cancel = context.WithTimeout(context.Background(), 2*time.Second)

			client, _ := mongo.Connect(testContext, options.Client().ApplyURI(mongoServerURI))
			wrapper = &MongoCollectionWrapper{
				collection: client.Database(testDatabaseName).Collection(testCollectionName),
				tracer: tracer.GlobalTracer,
			}
		})
		AfterEach(func() {
			cancel()
		})
		Context("Writing DB", func() {
			It("calls InsertOne", func() {
				_, err := wrapper.InsertOne(context.Background(), struct {
					Name string
				}{"TestName"})
				docsInserted++
				Expect(err).To(BeNil())
			})
			It("calls InsertMany", func() {
				_, err := wrapper.InsertMany(context.Background(), []interface{}{
					bson.D{
						{Key: "name", Value: "hello"},
						{Key: "age", Value: "33"},
					},
					bson.D{
						{Key: "name", Value: "world"},
						{Key: "age", Value: "44"},
					},
				})
				docsInserted += 2
				Expect(err).To(BeNil())
			})
		})
		Context("Reading DB", func() {
			It("calls FindOne", func() {
				wrapper.FindOne(
					context.Background(),
					bson.D{{Key: "name", Value: "helloworld"}},
				)
			})
			It("calls Find", func() {
				wrapper.Find(
					context.Background(),
					bson.M{},
				)
			})
			It("calls CountDocuments", func() {
				res, err := wrapper.CountDocuments(
					context.Background(),
					bson.D{{}},
				)
				Expect(err).To(BeNil())
				Expect(res).To(Equal(docsInserted))
			})
		})
	})
})
