package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// Rollout is an interface implemented by a value that can provide a rollout
// percentage as a decimal in the range [0.0,1.0]
type Rollout interface {
	// ReadRollout provides the current rollout value.
	// Note that it is possible for the value to be outside of the
	// acceptable [0.0,1.0] range - this should be checked by the caller.
	ReadRollout() float64
}

// A DynamoDBRollout represents a rollout that is continuously fetched from
// DynamoDB.
// The DynamoDB table must have a hash string key of "application", a range
// string key of "version", and a rollout number value stored under the key
// "rollout".
// "version" must always be set to "canary".
// If enough calls to DynamoDB fail, the rollout value will drop to 0 to
// minimize possible damange (i.e. the inability to rollback a canary).
type DynamoDBRollout struct {
	db      *dynamodb.DynamoDB
	table   string
	mutex   *sync.RWMutex
	rollout float64
}

// NewDynamoDBRollout creates a new DynamoDBRollout and begins eternally
// querying DynamoDB for canary values.
func NewDynamoDBRollout(monitor Monitor, db *dynamodb.DynamoDB, table string, application string, delay time.Duration, unhealthy time.Duration) *DynamoDBRollout {
	if db == nil {
		return nil
	}
	dynamodbRollout := &DynamoDBRollout{
		db:      db,
		table:   table,
		mutex:   &sync.RWMutex{},
		rollout: 0,
	}

	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]*dynamodb.AttributeValue{
			"application": {S: aws.String(application)},
			"version":     {S: aws.String("canary")},
		},
		ProjectionExpression: aws.String("rollout"),
		ConsistentRead:       aws.Bool(true),
	}
	loadRollout := func() (float64, error) {
		getItemOutput, err := db.GetItem(getItemInput)
		if err != nil {
			return 0, fmt.Errorf("could not fetch rollout value: %v\n", err)
		}
		percentageRaw := getItemOutput.Item["rollout"]
		if percentageRaw == nil {
			return 0, fmt.Errorf("could not find \"rollout\" key in response\n")
		}
		percentageString := percentageRaw.N
		if percentageString == nil {
			return 0, fmt.Errorf("rollout value is not stored as a number type\n")
		}
		percentage, err := strconv.ParseFloat(*percentageString, 64)
		if err != nil {
			return 0, fmt.Errorf("could not parse rollout value as a number: %v\n", err)
		}
		if percentage < 0 || percentage > 1 {
			return 0, fmt.Errorf("rollout value is out of [0.0,1.0] range")
		}
		return percentage, nil
	}
	go func() {
		lastHealthy := time.Now()
		for {
			percentage, err := loadRollout()
			monitor.RecordRolloutUpdate(err)
			if err != nil {
				log.Printf("rollout: %v\n", err)
				if time.Since(lastHealthy) > unhealthy {
					log.Printf("rollout: unhealthy for %v (rollout dropped to 0.0)\n", lastHealthy)
					percentage = 0
				}
			} else {
				lastHealthy = time.Now()
			}
			dynamodbRollout.mutex.Lock()
			dynamodbRollout.rollout = percentage
			dynamodbRollout.mutex.Unlock()
			time.Sleep(delay)
		}
	}()
	return dynamodbRollout
}

// ReadRollout provides the most recently read rollout value from DynamoDB.
// The return value may be outside of the [0.0,1.0] range.
func (dynamodbRollout *DynamoDBRollout) ReadRollout() float64 {
	dynamodbRollout.mutex.RLock()
	defer dynamodbRollout.mutex.RUnlock()
	return dynamodbRollout.rollout
}

// A ConstantRollout represents a rollout that will always have the same value.
type ConstantRollout struct {
	value float64
}

// NewConstantRollout creates a rollout that will always provide the sepcified
// value.
func NewConstantRollout(value float64) *ConstantRollout {
	return &ConstantRollout{
		value: value,
	}
}

// ReadRollout provides the rollout value that this value was created with.
// The return value may be outside of the [0.0,1.0] range.
func (constantRollout *ConstantRollout) ReadRollout() float64 {
	return constantRollout.value
}
