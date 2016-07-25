package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var region = flag.String("region", "us-west-2", "AWS region for DynamoDB")
var table = flag.String("table", "development-Deployment.Canarying", "DynamoDB table used for storage")

func putRolloutValue(db *dynamodb.DynamoDB, application string, value *dynamodb.AttributeValue) error {
	putItemInput := &dynamodb.PutItemInput{
		TableName: table,
		Item: map[string]*dynamodb.AttributeValue{
			"application": {S: aws.String(application)},
			"version":     {S: aws.String("canary")},
			"rollout":     value,
		},
	}
	_, err := db.PutItem(putItemInput)
	return err
}

func deleteRollout(db *dynamodb.DynamoDB, application string) error {
	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName: table,
		Key: map[string]*dynamodb.AttributeValue{
			"application": {S: aws.String(application)},
			"version":     {S: aws.String("canary")},
		},
	}
	_, err := db.DeleteItem(deleteItemInput)
	return err
}

func TestDynamoDBRolloutUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	client := &http.Client{
		Timeout: time.Second,
	}
	config := aws.NewConfig().WithRegion(*region).WithHTTPClient(client).WithMaxRetries(2)
	db := dynamodb.New(session.New(), config)
	key := fmt.Sprintf("test-%v", rand.Float64())
	err := putRolloutValue(db, key, &dynamodb.AttributeValue{N: aws.String("0.5")})
	if err != nil {
		t.Fatalf("error writing to dynamodb: %v", err)
	}
	defer deleteRollout(db, key)
	rollout, _ := NewDynamoDBRollout(&NilMonitor{}, db, *table, key, time.Second, 8*time.Second)
	time.Sleep(1 * time.Second)
	value := rollout.Get("canary")
	if value == nil || *value != 0.5 {
		t.Fatalf("expected to read 0.5, read %v", value)
	}
}

func TestDynamoDBRolloutMissing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	client := &http.Client{
		Timeout: time.Second,
	}
	config := aws.NewConfig().WithRegion(*region).WithHTTPClient(client).WithMaxRetries(2)
	db := dynamodb.New(session.New(), config)
	key := fmt.Sprintf("test-%v", rand.Float64())
	rollout, _ := NewDynamoDBRollout(&NilMonitor{}, db, *table, key, time.Second, 8*time.Second)
	time.Sleep(1 * time.Second)
	value := rollout.Get("canary")
	if value != nil {
		t.Fatalf("expected to read nil, read %v", value)
	}
}

func TestDynamoDBRolloutType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	client := &http.Client{
		Timeout: time.Second,
	}
	config := aws.NewConfig().WithRegion(*region).WithHTTPClient(client).WithMaxRetries(2)
	db := dynamodb.New(session.New(), config)
	key := fmt.Sprintf("test-%v", rand.Float64())
	err := putRolloutValue(db, key, &dynamodb.AttributeValue{S: aws.String("0.5")})
	if err != nil {
		t.Fatalf("error writing to dynamodb: %v", err)
	}
	defer deleteRollout(db, key)
	rollout, _ := NewDynamoDBRollout(&NilMonitor{}, db, *table, key, time.Second, 8*time.Second)
	time.Sleep(1 * time.Second)
	value := rollout.Get("canary")
	if value != nil {
		t.Fatalf("expected to read 0, read %v", value)
	}
}

func TestDynamoDBRolloutCall(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	client := &http.Client{
		Timeout: time.Second,
	}
	config := aws.NewConfig().WithRegion(*region).WithHTTPClient(client).WithMaxRetries(2)
	db := dynamodb.New(session.New(), config)
	key := fmt.Sprintf("test-%v", rand.Float64())
	err := putRolloutValue(db, key, &dynamodb.AttributeValue{N: aws.String("0.5")})
	if err != nil {
		t.Fatalf("error writing to dynamodb: %v", err)
	}
	defer deleteRollout(db, key)
	rollout, _ := NewDynamoDBRollout(&NilMonitor{}, db, *table, key, time.Second, 8*time.Second)
	time.Sleep(1 * time.Second)
	value := rollout.Get("canaryyyyy!!!")
	if value != nil {
		t.Fatalf("expected to read nil, read %v", value)
	}
}
