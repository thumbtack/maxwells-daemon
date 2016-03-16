package main

import (
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UTC().UnixNano())
	log.SetOutput(ioutil.Discard)
	flag.Parse()
	os.Exit(m.Run())
}
