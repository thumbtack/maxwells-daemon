package main

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestUnixServerDoubleClose(t *testing.T) {
	server, err := NewUnixServer(&NilMonitor{}, &EchoHandler{}, "/tmp/maxwells-daemon.sock")
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}
	err = server.Close()
	if err != nil {
		t.Fatalf("error closing server: %v", err)
	}
	err = server.Close()
	if err == nil {
		t.Fatalf("server was able to be doubly-closed")
	}
}

func TestUnixServerInvalidFilename(t *testing.T) {
	server, err := NewUnixServer(&NilMonitor{}, &EchoHandler{}, "/this/directory/doesnt.exist")
	if err == nil {
		server.Close()
		t.Fatalf("no error on invalid filename")
	}
}

func TestUnixServing(t *testing.T) {
	server, err := NewUnixServer(&NilMonitor{}, &EchoHandler{}, "/tmp/maxwells-daemon.sock")
	if err != nil {
		t.Fatalf("error starting server: %v", err)
	}
	defer server.Close()
	connection, err := net.Dial("unix", "/tmp/maxwells-daemon.sock")
	if err != nil {
		t.Fatalf("error connecting to server: %v", err)
	}
	connection.SetDeadline(time.Now().Add(time.Millisecond))
	sample := "0.123456789\n"
	fmt.Fprintf(connection, sample)
	response, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		t.Fatalf("could not read result from server: %v", err)
	}
	if sample != response {
		t.Fatalf("sample '%v' does not match response '%v'", sample, response)
	}
}
