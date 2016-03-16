package main

import (
	"math/rand"
	"strings"
	"testing"
)

func RandomString() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	strlen := 1 + rand.Intn(8)
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func wellformed(s string) bool {
	return strings.HasSuffix(s, "\nmaster\n") || strings.HasSuffix(s, "\ncanary\n")
}

func TestCanaryHandler0(t *testing.T) {
	rollout := NewConstantRollout(0)
	handler := NewCanaryHandler(&NilMonitor{}, rollout)
	response := ""
	for i := 0; i < 4096; i++ {
		response = handler.Handle(response)
		if !strings.HasSuffix(response, "\nmaster\n") {
			t.Fatalf("Didn't receive expected 'master' assignment")
		}
		response = strings.Fields(response)[0]
	}
}

func TestCanaryHandler100(t *testing.T) {
	rollout := NewConstantRollout(1)
	handler := NewCanaryHandler(&NilMonitor{}, rollout)
	response := ""
	for i := 0; i < 4096; i++ {
		response = handler.Handle(response)
		if !strings.HasSuffix(response, "\ncanary\n") {
			t.Fatalf("Didn't receive expected 'canary' assignment")
		}
		response = strings.Fields(response)[0]
	}
}

func TestCanaryHandler50(t *testing.T) {
	rollout := NewConstantRollout(0.5)
	handler := NewCanaryHandler(&NilMonitor{}, rollout)
	for i := 0; i < 4096; i++ { // pretend QuickCheck
		result := handler.Handle("")
		if !wellformed(result) {
			t.Errorf("invalid result: %v", result)
		}
	}
}

func TestCanaryHandlerInvalidRollout(t *testing.T) {
	rollout := NewConstantRollout(-1.0)
	handler := NewCanaryHandler(&NilMonitor{}, rollout)
	result := handler.Handle("")
	if !wellformed(result) {
		t.Errorf("invalid result: %v", result)
	}
	rollout = NewConstantRollout(2.0)
	handler = NewCanaryHandler(&NilMonitor{}, rollout)
	result = handler.Handle("")
	if !wellformed(result) {
		t.Errorf("invalid result: %v", result)
	}
}

func TestCanaryHandlerWhateverInput(t *testing.T) {
	rollout := NewConstantRollout(0.5)
	handler := NewCanaryHandler(&NilMonitor{}, rollout)
	for i := 0; i < 4096; i++ {
		result := handler.Handle(RandomString())
		if !wellformed(result) {
			t.Errorf("invalid result: %v", result)
		}
	}
}

func BenchmarkCanaryHandlerMaster(b *testing.B) {
	rollout := NewConstantRollout(0)
	handler := NewCanaryHandler(&NilMonitor{}, rollout)
	for i := 0; i < b.N; i++ {
		_ = handler.Handle("")
	}
}

func BenchmarkCanaryHandlerCanary(b *testing.B) {
	rollout := NewConstantRollout(1)
	handler := NewCanaryHandler(&NilMonitor{}, rollout)
	for i := 0; i < b.N; i++ {
		_ = handler.Handle("")
	}
}
