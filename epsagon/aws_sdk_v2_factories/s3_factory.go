package epsagonawsv2factories

import (
	"fmt"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
	"strings"
)

// S3EventDataFactory creates an Epsagon Resource from aws.Request to S3
func S3EventDataFactory(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Input).Elem()
	getResourceNameFromField(res, inputValue, "Bucket")

	handleSpecificOperations := map[string]specificOperationHandler{
		"HeadObject":  handleS3GetOrHeadObject,
		"GetObject":   handleS3GetOrHeadObject,
		"PutObject":   handleS3PutObject,
		"ListObjects": handleS3ListObject,
	}
	handleSpecificOperation(r, res, metadataOnly, handleSpecificOperations, nil, currentTracer)
}

func commonS3OperationHandler(r *AWSCall, res *protocol.Resource, metadataOnly bool) {
	responseValue := reflect.ValueOf(r.Res).Elem()
	inputValue := reflect.ValueOf(r.Input).Elem()
	outputValue := reflect.ValueOf(r.Output).Elem()

	updateMetadataFromNumValue(responseValue, "StatusCode", "status_code", res.Metadata)
	res.Metadata["region"] = r.Region

	if metadataOnly {
		return
	}

	updateMetadataFromValue(inputValue, "Key", "key", res.Metadata)
	res.Metadata["request_id"] = r.RequestID

	etag, ok := getFieldStringPtr(outputValue, "ETag")
	if ok {
		etag = strings.Trim(etag, "\"")
		res.Metadata["etag"] = etag
	}
}


func handleS3GetOrHeadObject(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	_ tracer.Tracer,
) {
	commonS3OperationHandler(r, res, metadataOnly)

	outputValue := reflect.ValueOf(r.Output).Elem()
	updateMetadataFromNumValue(outputValue, "ContentLength", "size", res.Metadata)
	updateMetadataFromNumValue(outputValue, "LastModified", "last_modified", res.Metadata)
}

func handleS3PutObject(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	_ tracer.Tracer,
) {
	commonS3OperationHandler(r, res, metadataOnly)

	outputValue := reflect.ValueOf(r.Output).Elem()
	updateMetadataFromNumValue(outputValue, "ContentLength", "size", res.Metadata)
}

type s3File struct {
	key  string
	size int64
	etag string
}

func handleS3ListObject(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	_ tracer.Tracer,
) {

	commonS3OperationHandler(r, res, metadataOnly)

	if metadataOnly {
		return
	}

	outputValue := reflect.ValueOf(r.Output).Elem()
	contentsField := outputValue.FieldByName("Contents")

	if contentsField != (reflect.Value{}) {
		length := contentsField.Len()
		files := make([]s3File, length)
		for i := 0; i < length; i++ {
			var key, etag string
			var size int64
			fileObject := contentsField.Index(i)
			etag = strings.Trim(
				fileObject.FieldByName("ETag").Elem().String(),
				"\"",
			)
			key = fileObject.FieldByName("Key").Elem().String()
			size = fileObject.FieldByName("Size").Int()

			files[i] = s3File{key, size, etag}
		}
		res.Metadata["files"] = fmt.Sprintf("%+v", files)
	}
}
