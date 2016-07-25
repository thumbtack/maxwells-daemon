package main

import (
	"io/ioutil"
	"path"
	"testing"
	"time"
)

func TestMaintenanceDaemonOff(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	fullpath := path.Join(dir, "test")
	if err != nil {
		t.Fatalf("Couldn't create temp dir for testing: %v", err)
	}
	rollout := NewConstantRollout(0)
	md, err := NewMaintenanceDaemon(fullpath, &NilMonitor{}, rollout)
	if err != nil {
		t.Fatalf("Failed creating MaintenanceDaemon: %v", err)
	}
	defer md.Stop()
	time.Sleep(8 * time.Millisecond)
	if md.IsOn() {
		t.Fatalf("Maintenance is on when it shouldn't be")
	}
}

func TestMaintenanceDaemonOn(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	fullpath := path.Join(dir, "test")
	if err != nil {
		t.Fatalf("Couldn't create temp dir for testing: %v", err)
	}
	rollout := NewConstantRollout(1)
	md, err := NewMaintenanceDaemon(fullpath, &NilMonitor{}, rollout)
	if err != nil {
		t.Fatalf("Failed creating MaintenanceDaemon: %v", err)
	}
	defer md.Stop()
	time.Sleep(8 * time.Millisecond)
	if !md.IsOn() {
		t.Fatalf("Maintenance is off when it shouldn't be")
	}
}
