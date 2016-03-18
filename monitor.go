package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

// Monitor is an interface implemented by a value that can record metrics about
// the daemon.
type Monitor interface {
	RecordServe(error)
	RecordServingTime(time.Duration)
	RecordHandling(string, error)
	RecordRolloutUpdate(error)
}

// A DogStatsDMonitor represents a proxy for sending metrics to Datadog using
// the namespace prefix "maxwellsdaemon.".
type DogStatsDMonitor struct {
	port uint16
	conn net.Conn
}

// NewDogStatsDMonitor creates a connection to the dogstatsd agent and prepares
// to accept metrics.
func NewDogStatsDMonitor(port uint16) *DogStatsDMonitor {
	return &DogStatsDMonitor{
		port: port,
		conn: nil,
	}
}

func (statsdMonitor *DogStatsDMonitor) send(s string) {
	if statsdMonitor.conn == nil {
		conn, err := net.Dial("udp", fmt.Sprintf("127.0.0.1:%v", statsdMonitor.port))
		if err != nil {
			log.Printf("monitor: could not connect to DogStatsD: %v\n", err)
			return
		}
		statsdMonitor.conn = conn
	}
	written, err := statsdMonitor.conn.Write([]byte(s))
	if err != nil || written != len([]byte(s)) {
		log.Printf("monitor: error writing data to DogStatsD: %v\n", err)
		_ = statsdMonitor.conn.Close()
		statsdMonitor.conn = nil
		return
	}
}

func (statsdMonitor *DogStatsDMonitor) RecordServe(err error) {
	if err != nil {
		statsdMonitor.send("maxwellsdaemon.server.success:0|c\n")
		statsdMonitor.send("maxwellsdaemon.server.failure:1|c\n")
	} else {
		statsdMonitor.send("maxwellsdaemon.server.success:1|c\n")
		statsdMonitor.send("maxwellsdaemon.server.failure:0|c\n")
	}
}

func (statsdMonitor *DogStatsDMonitor) RecordServingTime(duration time.Duration) {
	milliseconds := duration / time.Millisecond
	statsdMonitor.send(fmt.Sprintf("maxwellsdaemon.server.delay:%v|h\n", milliseconds))
}

func (statsdMonitor *DogStatsDMonitor) RecordHandling(location string, err error) {
	if err != nil {
		statsdMonitor.send("maxwellsdaemon.handler.success:0|c\n")
		statsdMonitor.send("maxwellsdaemon.handler.failure:1|c\n")
	} else {
		statsdMonitor.send("maxwellsdaemon.handler.success:1|c|#location:" + location + "\n")
		statsdMonitor.send("maxwellsdaemon.handler.failure:0|c\n")
	}
}

func (statsdMonitor *DogStatsDMonitor) RecordRolloutUpdate(err error) {
	if err != nil {
		statsdMonitor.send("maxwellsdaemon.rollout.update.success:0|c\n")
		statsdMonitor.send("maxwellsdaemon.rollout.update.failure:1|c\n")
	} else {
		statsdMonitor.send("maxwellsdaemon.rollout.update.success:1|c\n")
		statsdMonitor.send("maxwellsdaemon.rollout.update.failure:0|c\n")
	}
}

// NilMonitor represents a monitor sink - it records nothing.
type NilMonitor struct{}

func (nilMonitor *NilMonitor) RecordServe(_ error)               {}
func (nilMonitor *NilMonitor) RecordServingTime(_ time.Duration) {}
func (nilMonitor *NilMonitor) RecordHandling(_ string, _ error)  {}
func (nilMonitor *NilMonitor) RecordRolloutUpdate(_ error)       {}
