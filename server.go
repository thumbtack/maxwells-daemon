package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

// Server is an interface implemented by a value that can provide values to a
// handler in a separate goroutine.
type Server interface {
	// Close prevents the server from sending any more data to a handler.
	// A call to this method will block until the server has finished
	// processing all existing data.
	// This method will fail if the server is already closed.
	Close() error
}

// A UnixServer represents a listener on a unix socket that forwards data to a
// handler.
// This server expects data sent over a connection to be terminated with a '\n'
// symbol.
// The result from the handler will be sent back across the TCP connection,
// with each entry (assignment and location) terminated with a '\n' symbol.
type UnixServer struct {
	listener  *net.UnixListener
	waitGroup *sync.WaitGroup
	schan     chan time.Time
}

// TODO: Comply with Close()

// NewUnixServer creates a UnixServer that listens on the given file, passing
// incoming traffic to the given handler.
func NewUnixServer(monitor Monitor, handler Handler, filename string) (*UnixServer, error) {
	addr, err := net.ResolveUnixAddr("unix", filename)
	if err != nil {
		return nil, fmt.Errorf("error resolving unix address \"%v\": %v", filename, err)
	}
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, fmt.Errorf("error creating unix listener \"%v\": %v", filename, err)
	}
	os.Chmod(filename, 0666)
	unixServer := &UnixServer{
		listener:  listener,
		waitGroup: &sync.WaitGroup{},
		schan:     make(chan time.Time, 1),
	}
	unixServer.waitGroup.Add(1)
	go unixServer.serve(monitor, handler)
	return unixServer, nil
}

// Close prevents new connections from being made to the server and shuts down
// the server once existing connections are completed.
func (unixServer *UnixServer) Close() error {
	if unixServer.listener == nil {
		return fmt.Errorf("already stopped")
	}
	unixServer.schan <- time.Now()
	unixServer.waitGroup.Wait()
	unixServer.listener.Close()
	unixServer.listener = nil
	return nil
}

func (unixServer *UnixServer) serve(monitor Monitor, handler Handler) {
	defer unixServer.waitGroup.Done()
	wg := &sync.WaitGroup{}
	for {
		select {
		case <-unixServer.schan:
			wg.Wait()
			return
		default:
		}
		unixServer.listener.SetDeadline(time.Now().Add(time.Second))
		connection, err := unixServer.listener.Accept()
		if err, ok := err.(*net.OpError); ok && err.Timeout() {
			// a timeout occurred
			continue
		}
		if err != nil {
			monitor.RecordServe(err)
			log.Printf("server: error accepting connection: %v\n", err)
			connection.Close()
			continue
		}
		wg.Add(1)
		go func() {
			timeStart := time.Now()
			connection.SetDeadline(time.Now().Add(time.Second))
			var buffer bytes.Buffer
			data := make([]byte, 1024)
			for {
				requestLength, err := connection.Read(data)
				if err != nil {
					monitor.RecordServe(err)
					log.Printf("server: error reading from connection: %v\n", err)
					connection.Close()
					return
				}
				if data[requestLength-1] == '\n' {
					buffer.Write(data[:requestLength-1])
					break
				} else {
					buffer.Write(data)
				}
			}
			result := handler.Handle(string(buffer.Bytes()))
			count, err := connection.Write([]byte(result + "\n"))
			if err != nil {
				monitor.RecordServe(err)
				log.Printf("server: error writing to connection: %v\n", err)
			}
			expected := len([]byte(result + "\n"))
			if count != expected {
				err = fmt.Errorf("wrote %v out of %v bytes", count, expected)
				monitor.RecordServe(err)
				log.Printf("server: error writing to connection: %v\n", err)
			}
			connection.Close()
			monitor.RecordServe(nil)
			monitor.RecordServingTime(time.Since(timeStart))
			wg.Done()
		}()
	}
}
