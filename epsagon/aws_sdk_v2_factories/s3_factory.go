package epsagonawsv2factories

import (
	"fmt"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
)

// S3EventDataFactory creats an Epsagon Resource from aws.Request to S3
func S3EventDataFactory(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Req).Elem()

	fmt.Println("\nINPUT value s3 create event::")
	fmt.Println(inputValue)

	getResourceNameFromField(res, inputValue, "Bucket")

	handleSpecificOperations := map[string]specificOperationHandler{
		"HeadObject":  handleS3HeadObject,
		"GetObject":   handleS3GetObject,
		"PutObject":   handleS3PutObject,
		"ListObjects": handleS3ListObject,
	}
	handleSpecificOperation(r, res, metadataOnly, handleSpecificOperations, nil, currentTracer)
}

func commonS3OpertionHandler(r *AWSCall, res *protocol.Resource, metadataOnly bool) {
	//inputValue := reflect.ValueOf(r.Req).Elem()
	//updateMetadataFromValue(inputValue, "Key", "key", res.Metadata)
	//outputValue := reflect.ValueOf(r.Res).Elem()
	//etag, ok := getFieldStringPtr(outputValue, "ETag")
	//if ok {
	//	etag = strings.Trim(etag, "\"")
	//	res.Metadata["etag"] = etag
	//}
}

func handleS3GetObject(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	t tracer.Tracer,
) {
	// try opening req body

	//fmt.Println("OPENING GET OBJ BODY")
	//b := r.Res.Request.Body
	//fmt.Println(b)
	//body, err := ioutil.ReadAll(b)
	//fmt.Println("OPENING GET OBJ BODY")
	//
	//defer r.Req.Body.Close()
	//if err != nil {
	//	fmt.Println("could not read body")
	//	fmt.Println(err)
	//}
	//
	//fmt.Println("UNmarshalling")
	//var s3HeadObjInput s3.HeadObjectInput
	//err = json.Unmarshal(body, &s3HeadObjInput)
	//if err != nil {
	//	fmt.Println("could not head object")
	//}
	fmt.Println("GOT TO BOTTOM OF GET")
	commonS3OpertionHandler(r, res, metadataOnly)
	handleS3GetOrHeadObject(r, res, metadataOnly, t)
}

func handleS3HeadObject(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	t tracer.Tracer,
) {
	commonS3OpertionHandler(r, res, metadataOnly)
	handleS3GetOrHeadObject(r, res, metadataOnly, t)
}

func handleS3GetOrHeadObject(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	_ tracer.Tracer,
) {
	//outputValue := reflect.ValueOf(r.Res).Elem()
	//updateMetadataFromValue(outputValue.FieldByName("ContentLength"), "ContentLength", "file_size", res.Metadata)
	//
	//lastModifiedField := outputValue.FieldByName("LastModified")
	//if lastModifiedField == (reflect.Value{}) {
	//	return
	//}
	//lastModified := lastModifiedField.Elem().Interface().(time.Time)
	//res.Metadata["last_modified"] = lastModified.String()
}

func handleS3PutObject(
	r *AWSCall,
	res *protocol.Resource,
	metadataOnly bool,
	_ tracer.Tracer,
) {
	commonS3OpertionHandler(r, res, metadataOnly)
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
	if metadataOnly {
		return
	}

	outputValue := reflect.ValueOf(r.Res).Elem()
	contentsField := outputValue.FieldByName("Contents")
	if contentsField == (reflect.Value{}) {
		return
	}
	length := contentsField.Len()
	files := make([]s3File, length)
	for i := 0; i < length; i++ {
		var key, etag string
		var size int64
		fileObject := contentsField.Index(i).Elem()
		etag = fileObject.FieldByName("ETag").Elem().String()
		key = fileObject.FieldByName("Key").Elem().String()
		size = fileObject.FieldByName("Size").Elem().Int()

		files = append(files, s3File{key, size, etag})
	}
	res.Metadata["files"] = fmt.Sprintf("%+v", files)
}
