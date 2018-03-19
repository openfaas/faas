package main

import (
	"net/http"
	"syscall"
	"testing"
	"time"
)

type server struct{}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a mock!"))
}

func simulateSIGTERM(pid int) {
	time.Sleep(3 * time.Second)
	syscall.Kill(pid, syscall.SIGTERM)
}

func Test_Should_Gracefully_Shutdown(t *testing.T) {
	var pid int = syscall.Getpid()
	successChan := make(chan bool, 1)

	watchdog := &http.Server{Addr: ":8080", Handler: &server{}}
	go gracefulShutdown(watchdog, successChan) // Method under test
	go simulateSIGTERM(pid)

	success := <-successChan

	if !success {
		t.Fail()
	}
}
