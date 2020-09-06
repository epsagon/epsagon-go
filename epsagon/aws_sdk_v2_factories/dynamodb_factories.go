package epsagonawsv2factories

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
	"reflect"
)

// DynamodbEventDataFactory to create epsagon Resource from aws.Request to DynamoDB
func DynamodbEventDataFactory(
	r *aws.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	tableName, ok := getFieldStringPtr(inputValue, "TableName")
	if ok {
		res.Name = tableName
	}
	handleSpecificOperations := map[string]specificOperationHandler{
		"PutItem":        handleDynamoDBPutItem,
		"GetItem":        handleDynamoDBGetItem,
		"DeleteItem":     handleDynamoDBDeleteItem,
		"UpdateItem":     handleDynamoDBUpdateItem,
		"Scan":           handleDynamoDBScan,
		"BatchWriteItem": handleDynamoDBBatchWriteItem,
	}
	handleSpecificOperation(r, res, metadataOnly, handleSpecificOperations, nil, currentTracer)
}

func deserializeAttributeMap(inputField reflect.Value) map[string]string {
	formattedItem := make(map[string]string)
	input := inputField.Interface().(map[string]dynamodb.AttributeValue)
	for k, v := range input {
		formattedItem[k] = v.String()
	}
	return formattedItem
}

func jsonAttributeMap(inputField reflect.Value, currentTracer tracer.Tracer) string {
	if inputField == (reflect.Value{}) {
		return ""
	}
	formattedMap := deserializeAttributeMap(inputField)
	stream, err := json.Marshal(formattedMap)
	if err != nil {
		currentTracer.AddExceptionTypeAndMessage("aws-sdk-go", fmt.Sprintf("%v", err))
		return ""
	}
	return string(stream)
}

func handleDynamoDBPutItem(
	r *aws.Request,
	res *protocol.Resource,
	metadataOnly bool,
	_ tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	itemField := inputValue.FieldByName("Item")
	if itemField == (reflect.Value{}) {
		return
	}
	formattedItem := deserializeAttributeMap(itemField)
	formattedItemStream, err := json.Marshal(formattedItem)
	if err != nil {
		// TODO send tracer exception?
		return
	}
	if !metadataOnly {
		res.Metadata["Item"] = string(formattedItemStream)
	}
	h := md5.New()
	h.Write(formattedItemStream)
	res.Metadata["item_hash"] = hex.EncodeToString(h.Sum(nil))
}

func handleDynamoDBGetItem(
	r *aws.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	jsonKeyField := jsonAttributeMap(inputValue.FieldByName("Key"), currentTracer)
	res.Metadata["Key"] = jsonKeyField

	if !metadataOnly {
		outputValue := reflect.ValueOf(r.Data).Elem()
		jsonItemField := jsonAttributeMap(outputValue.FieldByName("Item"), currentTracer)
		res.Metadata["Item"] = jsonItemField
	}
}

func handleDynamoDBDeleteItem(
	r *aws.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	jsonKeyField := jsonAttributeMap(inputValue.FieldByName("Key"), currentTracer)
	res.Metadata["Key"] = jsonKeyField
}

func handleDynamoDBUpdateItem(
	r *aws.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	eavField := inputValue.FieldByName("ExpressionAttributeValues")
	eav := deserializeAttributeMap(eavField)
	eavStream, err := json.Marshal(eav)
	if err != nil {
		return
	}
	updateParameters := map[string]string{
		"Expression Attribute Values": string(eavStream),
	}
	jsonKeyField := jsonAttributeMap(inputValue.FieldByName("Key"), currentTracer)
	updateParameters["Key"] = jsonKeyField
	updateMetadataFromValue(inputValue,
		"UpdateExpression", "UpdateExpression", updateParameters)
	updateParamsStream, err := json.Marshal(updateParameters)
	if err != nil {
		return
	}
	res.Metadata["Update Parameters"] = string(updateParamsStream)
}

func deserializeItems(itemsField reflect.Value, currentTracer tracer.Tracer) string {
	if itemsField == (reflect.Value{}) {
		return ""
	}
	formattedItems := make([]map[string]string, itemsField.Len())
	for ind := 0; ind < itemsField.Len(); ind++ {
		formattedItems = append(formattedItems,
			deserializeAttributeMap(itemsField.Index(ind)))
	}
	formattedItemsStream, err := json.Marshal(formattedItems)
	if err != nil {
		currentTracer.AddExceptionTypeAndMessage("aws-sdk-go",
			fmt.Sprintf("sederializeItems: %v", err))
	}
	return string(formattedItemsStream)
}

func handleDynamoDBScan(
	r *aws.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	outputValue := reflect.ValueOf(r.Params).Elem()
	updateMetadataFromInt64(outputValue, "Count", "Items Count", res.Metadata)
	updateMetadataFromInt64(outputValue, "ScannedCount", "Scanned Items Count", res.Metadata)
	itemsField := outputValue.FieldByName("Items")
	if !metadataOnly {
		res.Metadata["Items"] = deserializeItems(itemsField, currentTracer)
	}
}

func handleDynamoDBBatchWriteItem(
	r *aws.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	requestItemsField := inputValue.FieldByName("RequestItems")
	if requestItemsField != (reflect.Value{}) {
		var tableName string
		requestItems, ok := requestItemsField.Interface().(map[string][]*dynamodb.WriteRequest)
		if !ok {
			currentTracer.AddExceptionTypeAndMessage("aws-sdk-go",
				"handleDynamoDBBatchWriteItem: Failed to cast RequestItems")
			return
		}
		for k := range requestItems {
			tableName = k
			break
		}
		res.Name = tableName
		// TODO not ignore other tables
		if !metadataOnly {
			items := make([]map[string]dynamodb.AttributeValue, len(requestItems))
			for _, writeRequest := range requestItems[tableName] {
				items = append(items, writeRequest.PutRequest.Item)
			}
			itemsValue := reflect.ValueOf(items)
			res.Metadata["Items"] = deserializeItems(itemsValue, currentTracer)
		}
	}
}
