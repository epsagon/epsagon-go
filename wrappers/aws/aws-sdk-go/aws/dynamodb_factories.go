package epsagonawswrapper

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/epsagon/epsagon-go/protocol"
	"github.com/epsagon/epsagon-go/tracer"
)

func dynamodbEventDataFactory(
	r *request.Request,
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
		"PutItem":          handleDynamoDBPutItem,
		"GetItem":          handleDynamoDBGetItem,
		"DeleteItem":       handleDynamoDBDeleteItem,
		"UpdateItem":       handleDynamoDBUpdateItem,
		"Scan":             handleDynamoDBScan,
		"Query":            handleDynamoDBQuery,
		"QueryWithContext": handleDynamoDBQuery,
		"BatchWriteItem":   handleDynamoDBBatchWriteItem,
		//"TransactWriteItems":            handleDynamoDBTransactWriteItems,
		//"TransactWriteItemsWithContext": handleDynamoDBTransactWriteItems,
	}
	handler := handleSpecificOperations[res.Operation]
	if handler != nil {
		handler(r, res, metadataOnly, currentTracer)
	}
}

func deserializeAttributeMap(input map[string]*dynamodb.AttributeValue) map[string]string {
	formattedData := make(map[string]string)
	for k, v := range input {
		formattedData[k] = v.String()
	}
	return formattedData
}

func deserializeRawAttributeMap(inputField reflect.Value) map[string]string {
	input := inputField.Interface().(map[string]*dynamodb.AttributeValue)
	return deserializeAttributeMap(input)
}

func jsonAttributeMap(inputField reflect.Value, currentTracer tracer.Tracer) string {
	if inputField == (reflect.Value{}) {
		return ""
	}
	formattedMap := deserializeRawAttributeMap(inputField)
	stream, err := json.Marshal(formattedMap)
	if err != nil {
		currentTracer.AddExceptionTypeAndMessage("aws-sdk-go", fmt.Sprintf("%v", err))
		return ""
	}
	return string(stream)
}

func handleDynamoDBPutItem(
	r *request.Request,
	res *protocol.Resource,
	metadataOnly bool,
	_ tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	itemField := inputValue.FieldByName("Item")
	if itemField == (reflect.Value{}) {
		return
	}
	formattedItem := deserializeRawAttributeMap(itemField)
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
	r *request.Request,
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
	r *request.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	jsonKeyField := jsonAttributeMap(inputValue.FieldByName("Key"), currentTracer)
	res.Metadata["Key"] = jsonKeyField
}

func handleDynamoDBUpdateItem(
	r *request.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	eavField := inputValue.FieldByName("ExpressionAttributeValues")
	eav := deserializeRawAttributeMap(eavField)
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
	if isValueZero(itemsField) {
		return ""
	}

	formattedItems := deserializeItemsRaw(itemsField, currentTracer)
	formattedItemsStream, err := json.Marshal(formattedItems)
	if err != nil {
		currentTracer.AddExceptionTypeAndMessage("aws-sdk-go",
			fmt.Sprintf("deserializeItems: %v", err))
	}
	return string(formattedItemsStream)
}

func deserializeItemsRaw(itemsField reflect.Value, currentTracer tracer.Tracer) interface{} {
	if isValueZero(itemsField) {
		return ""
	}
	formattedItems := make([]map[string]string, 0, itemsField.Len())
	for ind := 0; ind < itemsField.Len(); ind++ {
		formattedItem := deserializeRawAttributeMap(itemsField.Index(ind))
		if formattedItem != nil && len(formattedItem) > 0 {
			formattedItems = append(formattedItems, formattedItem)
		}
	}
	return formattedItems
}

func deserializeCondition(condition *dynamodb.Condition) map[string]interface{} {
	formattedData := make(map[string]interface{})
	formattedData["ComparisonOperator"] = *(condition.ComparisonOperator)
	attributes := make([]string, len(condition.AttributeValueList))
	for _, attribute := range condition.AttributeValueList {
		attributes = append(attributes, attribute.String())
	}
	formattedData["AttributeValueList"] = attributes
	return formattedData
}

func deserializeKeyConditions(keyConditionsField reflect.Value) interface{} {
	if isValueZero(keyConditionsField) {
		return map[string]interface{}{}
	}

	conditionMap := make(map[string]interface{})
	input := keyConditionsField.Interface().(map[string]*dynamodb.Condition)
	for k, v := range input {
		conditionMap[k] = deserializeCondition(v)
	}
	return conditionMap
}

func handleDynamoDBScan(
	r *request.Request,
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

func updateWithJsonMap(
	destination map[string]string,
	destinationKey string,
	data map[string]interface{},
) error {
	stream, err := json.Marshal(data)
	if err != nil {
		return err
	}
	destination[destinationKey] = string(stream)
	return nil
}

func updateWithStringValue(
	inputValue reflect.Value,
	destination map[string]interface{},
	key string,
) {
	value, ok := getFieldStringPtr(inputValue, key)
	if ok {
		destination[key] = value
	}
}

func updateMapFromInt64(
	inputValue reflect.Value,
	destination map[string]interface{},
	destinationKey string,
	sourceKey string,
) {
	field := inputValue.FieldByName(sourceKey)
	if isValueZero(field) {
		return
	}
	destination[destinationKey] = strconv.FormatInt(field.Elem().Int(), 10)
}

func updateMapWithFieldToJSON(
	inputValue reflect.Value,
	destination map[string]interface{},
	key string,
) {
	field := inputValue.FieldByName(key)
	if isValueZero(field) {
		return
	}
	destination[key] = field.Interface()
}

func handleDynamoDBQuery(
	r *request.Request,
	res *protocol.Resource,
	metadataOnly bool,
	currentTracer tracer.Tracer,
) {
	inputValue := reflect.ValueOf(r.Params).Elem()
	outputValue := reflect.ValueOf(r.Data).Elem()
	responseMap := map[string]interface{}{}
	parameters := map[string]interface{}{}

	updateMapFromInt64(outputValue, responseMap, "Items Count", "Count")
	updateMapFromInt64(outputValue, responseMap, "Scanned Items Count", "ScannedCount")
	updateMapWithFieldToJSON(outputValue, responseMap, "ConsumedCapacity")
	updateMapWithFieldToJSON(outputValue, responseMap, "ResponseMetadata")
	if !metadataOnly {
		itemsField := outputValue.FieldByName("Items")
		responseMap["Items"] = deserializeItemsRaw(itemsField, currentTracer)
		updateWithStringValue(inputValue, parameters, "KeyConditionExpression")
		updateWithStringValue(inputValue, parameters, "FilterExpression")
		updateWithStringValue(inputValue, parameters, "ReturnConsumedCapacity")
		updateMapWithFieldToJSON(inputValue, parameters, "ExpressionAttributeNames")
		keyConditionsField := inputValue.FieldByName("KeyConditions")
		parameters["KeyConditions"] = deserializeKeyConditions(keyConditionsField)
		eavField := inputValue.FieldByName("ExpressionAttributeValues")
		eav := deserializeRawAttributeMap(eavField)
		parameters["Expression Attribute Values"] = eav
	}
	if updateWithJsonMap(res.Metadata, "Response", responseMap) != nil {
		return
	}
	updateWithJsonMap(res.Metadata, "Parameters", parameters)
}

func handleDynamoDBBatchWriteItem(
	r *request.Request,
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
			items := make([]map[string]*dynamodb.AttributeValue, len(requestItems))
			for _, writeRequest := range requestItems[tableName] {
				items = append(items, writeRequest.PutRequest.Item)
			}
			itemsValue := reflect.ValueOf(items)
			res.Metadata["Items"] = deserializeItems(itemsValue, currentTracer)
		}
	}
}
