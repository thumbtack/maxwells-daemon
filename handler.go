package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
)

// Handler is an interface implemented by a value that can convert an input
// value to some output value.
type Handler interface {
	// Handle is the method by which input is converted to output.
	Handle(string) string
}

// A CanaryHandler represents a handler that will determine a location
// (master/canary) given an assignment.
type CanaryHandler struct {
	monitor Monitor
	rollout Rollout
}

// NewCanaryHandler creates a handler that will calculate whether an assignment
// is canaried or not.
func NewCanaryHandler(monitor Monitor, rollout Rollout) *CanaryHandler {
	return &CanaryHandler{
		monitor: monitor,
		rollout: rollout,
	}
}

// Handle parses the given assignment and returns an assignment and
// location (master/canary) each suffixed with a newline ('\n').
// A valid return value will always be produced (regardless of input).
func (canaryHandler *CanaryHandler) Handle(input string) string {
	format := "%v\n%v\n"
	if len(input) == 0 {
		input = fmt.Sprintf("%v", rand.Float64())
	}
	value := "master"
	assignment, err := strconv.ParseFloat(input, 64)
	if err != nil {
		canaryHandler.monitor.RecordHandling(value, err)
		log.Printf("handler: could not parse assignment as a number: %v\n", err)
		return fmt.Sprintf(format, rand.Float64(), value)
	}
	if assignment < 0 || assignment >= 1 {
		canaryHandler.monitor.RecordHandling(value, fmt.Errorf("assignment out of range"))
		log.Printf("handler: assignment is out of [0.0,1.0) range: %v\n", assignment)
		return fmt.Sprintf(format, rand.Float64(), value)
	}
	rollout := canaryHandler.rollout.Get()
	if rollout < 0 || rollout > 1 {
		canaryHandler.monitor.RecordHandling(value, fmt.Errorf("rollout out of range"))
		log.Printf("handler: rollout is out of [0.0,1.0] range\n")
		return fmt.Sprintf(format, assignment, value)
	}
	if assignment < rollout {
		value = "canary"
	}
	canaryHandler.monitor.RecordHandling(value, nil)
	return fmt.Sprintf(format, assignment, value)
}

// An EchoHandler represents a handler that will return the given input with no
// processing applied.
type EchoHandler struct{}

// Handle returns the given input unmodified.
func (echoHandler *EchoHandler) Handle(input string) string {
	return input
}
