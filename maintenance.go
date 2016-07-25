package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// A MaintenanceDaemon is a structure to continually update the existence of a
// file depending on whether or not maintenance mode is enabled.
type MaintenanceDaemon struct {
	fullpath string
	ch       chan interface{}
	wg       *sync.WaitGroup
	rwm      *sync.RWMutex
}

func (md *MaintenanceDaemon) on() {
	_, err := os.Create(md.fullpath)
	if err != nil {
		log.Printf("maintenance: could not create maintenance file")
	}
}

func (md *MaintenanceDaemon) off() {
	err := os.Remove(md.fullpath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("maintenance: could not remove maintenance file")
	}
}

func (md *MaintenanceDaemon) update(value *float64) {
	// if there is no valid data, don't risk changing the maintenace state
	if value == nil {
		return
	}
	if *value > 0 {
		md.on()
	} else {
		md.off()
	}
}

// NewMaintenanceDaemon creates a daemon that controls the given file.
func NewMaintenanceDaemon(fullpath string, monitor Monitor, rollout Rollout) (*MaintenanceDaemon, error) {
	if fullpath == "" {
		return nil, fmt.Errorf("fullpath string is empty")
	}
	md := &MaintenanceDaemon{
		fullpath: fullpath,
		ch:       make(chan interface{}),
		wg:       &sync.WaitGroup{},
		rwm:      &sync.RWMutex{},
	}
	md.update(rollout.Get("maintenance"))
	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		select {
		case <-md.ch:
			return
		case <-tick.C:
			md.update(rollout.Get("maintenance"))
		}
	}()
	return md, nil
}

// IsOn returns the status of maintenance (on or off).
func (md *MaintenanceDaemon) IsOn() bool {
	_, err := os.Stat(md.fullpath)
	return err == nil
}

// Stop stops the daemon forever.
func (md *MaintenanceDaemon) Stop() {
	md.rwm.Lock()
	if md.ch != nil {
		md.ch <- nil
	}
	md.ch = nil
	md.rwm.Unlock()
	if md.wg != nil {
		md.wg.Wait()
	}
}
