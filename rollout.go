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
	// Get provides the current rollout value for a specific version.
	// nil will be returned if the value cannot be provided (non-existent
	// or stale).
	// Note that it is possible for the value to be outside of the
	// acceptable [0.0,1.0] range - this should be checked by the caller.
	Get(string) *float64
}

type version struct {
	lastUpdated time.Time
	isUnhealthy bool
	percentage  float64
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
	db       *dynamodb.DynamoDB
	table    string
	mutex    *sync.RWMutex
	versions map[string]version
}

func (r *DynamoDBRollout) update(unhealthy time.Duration, updates map[string]float64) {
	if updates == nil {
		updates = make(map[string]float64)
	}
	r.mutex.Lock()
	// update existing entries
	for key, val := range r.versions {
		newPercentage, ok := updates[key]
		if !ok {
			if time.Since(val.lastUpdated) > unhealthy {
				if !val.isUnhealthy && val.percentage > 0 {
					log.Printf("rollout: \"%v\" contains stale data", key)
				}
				val.percentage = 0
				val.isUnhealthy = true
			}
		} else {
			val.lastUpdated = time.Now()
			val.isUnhealthy = false
			val.percentage = newPercentage
		}
	}
	// add missing entries
	for key, val := range updates {
		_, ok := r.versions[key]
		if !ok {
			r.versions[key] = version{
				lastUpdated: time.Now(),
				isUnhealthy: false,
				percentage:  val,
			}
		}
	}
	r.mutex.Unlock()
}

// NewDynamoDBRollout creates a new DynamoDBRollout and begins eternally
// querying the given DynamoDB table in the given region for canary values for
// the given application.
// Queries to DynamoDB are interspersed with the given delay to avoid using up
// all the read capacity.
// If calls to DynamoDB fail / are unhealthy for the specified amount of time,
// rollout will be dropped 0.0.
func NewDynamoDBRollout(monitor Monitor, db *dynamodb.DynamoDB, table string, application string, delay time.Duration, unhealthy time.Duration) (*DynamoDBRollout, error) {
	const hashField string = "application"
	const rangeField string = "version"
	const rolloutField string = "rollout"
	// []strings are not constants
	rangeKeys := []string{
		"maintenance",
		"canary",
	}

	if db == nil {
		return nil, fmt.Errorf("dynamo.DynamoDB argument is nil")
	}
	dynamodbRollout := &DynamoDBRollout{
		db:       db,
		table:    table,
		mutex:    &sync.RWMutex{},
		versions: make(map[string]version),
	}
	var keys []map[string]*dynamodb.AttributeValue
	for _, key := range rangeKeys {
		keys = append(keys, map[string]*dynamodb.AttributeValue{
			hashField:  {S: aws.String(application)},
			rangeField: {S: aws.String(key)},
		})
	}
	batchGetItemInput := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			table: &dynamodb.KeysAndAttributes{
				ProjectionExpression: aws.String(fmt.Sprintf("%v, %v", rangeField, rolloutField)),
				ConsistentRead:       aws.Bool(true),
				Keys:                 keys,
			},
		},
	}
	loadRollouts := func() map[string]float64 {
		batchGetItemOutput, err := db.BatchGetItem(batchGetItemInput)
		if err != nil {
			log.Printf("could not fetch rollout values: %v", err)
			return nil
		}
		tableItems, ok := batchGetItemOutput.Responses[table]
		if !ok {
			log.Printf("could not find rollout values in response")
			return nil
		}
		results := make(map[string]float64)
		loadRollout := func(item map[string]*dynamodb.AttributeValue) error {
			nameRaw, ok := item[rangeField]
			if !ok {
				return fmt.Errorf("could not find \"%s\" key in response item", rangeField)
			}
			name := nameRaw.S
			if name == nil {
				return fmt.Errorf("release name is not stored as a string type")
			}
			if *name == "" {
				return fmt.Errorf("release name is empty string")
			}
			percentageRaw := item[rolloutField]
			if percentageRaw == nil {
				return fmt.Errorf("could not find \"%s\" key in response", rolloutField)
			}
			percentageString := percentageRaw.N
			if percentageString == nil {
				return fmt.Errorf("rollout value is not stored as a number type")
			}
			percentage, err := strconv.ParseFloat(*percentageString, 64)
			if err != nil {
				return fmt.Errorf("could not parse rollout value as a number: %v", err)
			}
			if percentage < 0 || percentage > 1 {
				return fmt.Errorf("rollout value is out of [0.0,1.0] range")
			}
			results[*name] = percentage
			return nil
		}
		for _, item := range tableItems {
			err := loadRollout(item)
			monitor.RecordRolloutUpdate(err)
			if err != nil {
				log.Printf("rollout: error during update: %v", err)
			}
		}
		return results
	}
	go func() {
		for {
			updates := loadRollouts()
			dynamodbRollout.update(unhealthy, updates)
			time.Sleep(delay)
		}
	}()
	return dynamodbRollout, nil
}

// Get provides the most recently read rollout value from DynamoDB.
// The return value may be outside of the [0.0,1.0] range.
func (r *DynamoDBRollout) Get(name string) *float64 {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	version, ok := r.versions[name]
	if !ok {
		log.Printf("rollout: request for nonexistent version: \"%v\"", name)
		return nil
	}
	if version.isUnhealthy {
		return nil
	}
	return &version.percentage
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

// Get provides the rollout value that this value was created with.
// The return value may be outside of the [0.0,1.0] range.
func (constantRollout *ConstantRollout) Get(_ string) *float64 {
	return &constantRollout.value
}
