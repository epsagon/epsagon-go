package epsagonmongo

import (
	"context"
	"reflect"

	//"github.com/aws/aws-sdk-go-v2/config"
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
	//mongoServer, err := memongo.Start("4.2.0")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//defer mongoServer.Stop()
	//mongoServerURI = mongoServer.URI()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Mongo Driver Test Suite")
}

var _ = Describe("mongo_wrapper", func() {
	Describe("CollectionWrapper", func() {
		var (
			mongoServer			*memongo.Server
			started				chan bool
			testConf			*epsagon.Config
			events             []*protocol.Event
			exceptions         []*protocol.Exception
			wrapper            *MongoCollectionWrapper
			testContext        context.Context
			testDatabaseName   string
			testCollectionName string
			cancel             func()
		)
		BeforeEach(func() {
			started = make(chan bool)
			// start server goroutine, runs in background until block
			go func()  {
				mongoServer, _ = memongo.Start("4.2.0")
				started <- true
			}()

			testConf = &epsagon.Config{Config: tracer.Config{
				Disable:  true,
				TestMode: true,
			}}
			events = make([]*protocol.Event, 0)
			exceptions = make([]*protocol.Exception, 0)
			tracer.GlobalTracer = &tracer.MockedEpsagonTracer{
				Events:     &events,
				Exceptions: &exceptions,
				Labels:     make(map[string]interface{}),
				Config:     &testConf.Config,
			}

			testCollectionName = "collectionName"
			testDatabaseName = "databaseName"
			testContext, cancel = context.WithTimeout(context.Background(), 2*time.Second)

			// blocking await until server is started
			select {
			case <-started:
				break
			}
			client, _ := mongo.Connect(testContext, options.Client().ApplyURI(mongoServer.URI()))
			wrapper = &MongoCollectionWrapper{
				collection: client.Database(testDatabaseName).Collection(testCollectionName),
				tracer:     tracer.GlobalTracer,
			}
		})
		AfterEach(func() {
			mongoServer.Stop()
			cancel()
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
						{Key: "name", Value: "hello"},
						{Key: "age", Value: "33"},
					},
					bson.D{
						{Key: "name", Value: "world"},
						{Key: "age", Value: "44"},
					},
				})
				Expect(err).To(BeNil())
			})
		})
		Context("Reading DB", func() {
			It("calls InsertOne and FindOne", func() {
				type doc struct {
					Name string
				}
				reqDoc := doc{Name: "TestName"}
				resDoc := reqDoc

				wrapper.InsertOne(context.Background(), reqDoc)
				res := wrapper.FindOne(
					context.Background(),
					bson.D{{Key: "name", Value: "TestName"}},
				)
				reflect.ValueOf(res).
					MethodByName("Decode").
					Call([]reflect.Value{reflect.ValueOf(&resDoc)})
				Expect(reqDoc).To(Equal(resDoc))
			})
			It("calls InsertMany and Find", func() {
				type doc struct {
					Name string
				}
				docs := []interface{}{
					bson.D{
						{Key: "name", Value: "hello"},
						{Key: "age", Value: "33"},
					},
					bson.D{
						{Key: "name", Value: "world"},
						{Key: "age", Value: "44"},
					},
				}

				wrapper.InsertMany(context.Background(), docs)
				cur, _ := wrapper.Find(
					context.Background(),
					bson.M{},
				)

				readCursor(cur)
			})
			It("calls CountDocuments", func() {
				docs := []interface{}{
					bson.D{
						{Key: "name", Value: "hello"},
						{Key: "age", Value: "33"},
					},
					bson.D{
						{Key: "name", Value: "world"},
						{Key: "age", Value: "44"},
					},
				}

				wrapper.InsertMany(context.Background(), docs)
				res, err := wrapper.CountDocuments(
					context.Background(),
					bson.D{{}},
				)
				Expect(err).To(BeNil())
				Expect(res).To(Equal(int64(2)))
			})
		})
	})
})
