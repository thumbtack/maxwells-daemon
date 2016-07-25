package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// TODO: Try github.com/golang/glog
func main() {
	socket := flag.String("socket", "/tmp/maxwells-daemon.sock", "path to the unix socket file")
	application := flag.String("application", "app", "application name (referenced by DynamoDB)")
	region := flag.String("region", "us-east-1", "AWS region for DynamoDB")
	table := flag.String("table", "MaxwellsDaemon", "DynamoDB table used for rollout data")
	delay := flag.Duration("delay", 4*time.Second, "minimum delay between DynamoDB rollout requests")
	unhealthy := flag.Duration("unhealthy", 8*time.Second, "minimum duration to allow unhealthy DynamoDB querying before reverting to 0.0 rollout")
	logfile := flag.String("logfile", "/var/log/maxwells-daemon.log", "path to the log file")
	stateDir := flag.String("state-dir", "/var/lib/maxwells-daemon", "path to the app's state directory")
	flag.Parse()

	rand.Seed(time.Now().UTC().UnixNano())
	handle, err := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("error opening log file: %v: %v", *logfile, err)
	}
	log.SetOutput(handle)

	// monitor
	monitor := NewDogStatsDMonitor(8125)

	// rollout
	client := &http.Client{
		Timeout: time.Second,
	}
	config := aws.NewConfig().WithRegion(*region).WithHTTPClient(client).WithMaxRetries(2)
	db := dynamodb.New(session.New(), config)
	rollout, err := NewDynamoDBRollout(monitor, db, *table, *application, *delay, *unhealthy)
	if err != nil {
		log.Fatalf("error creating rollout: %v\n", err)
	}

	// handler
	handler := NewCanaryHandler(monitor, rollout)

	// maintenance daemon
	maintenance, err := NewMaintenanceDaemon(path.Join(*stateDir, *application), monitor, rollout)
	if err != nil {
		log.Fatal("error creating maintenance daemon: %v\n", err)
	}

	// server
	os.Remove(*socket)
	server, err := NewUnixServer(monitor, handler, *socket)
	if err != nil {
		log.Fatalf("error starting server: %v\n", err)
	}

	// endless waiting (signal handler)
	sigchan := make(chan os.Signal, 1024)
	signal.Notify(sigchan, os.Interrupt)
	log.Printf("started daemon\n")
	select {
	case <-sigchan:
		log.Printf("received interrupt signal\n")
		server.Close()
		handle.Close()
		maintenance.Stop()
		os.Exit(130)
	}
}
