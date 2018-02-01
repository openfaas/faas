// Copyright 2012-2016 Apcera Inc. All rights reserved.
// +build ignore

package main

import (
	"flag"
	"log"
	"time"

	"github.com/nats-io/go-nats"
)

// NOTE: Use tls scheme for TLS, e.g. nats-req -s tls://demo.nats.io:4443 foo hello
func usage() {
	log.Fatalf("Usage: nats-req [-s server (%s)] <subject> <msg> \n", nats.DefaultURL)
}

func main() {
	var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		usage()
	}

	nc, err := nats.Connect(*urls)
	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}
	defer nc.Close()
	subj, payload := args[0], []byte(args[1])

	msg, err := nc.Request(subj, []byte(payload), 100*time.Millisecond)
	if err != nil {
		if nc.LastError() != nil {
			log.Fatalf("Error in Request: %v\n", nc.LastError())
		}
		log.Fatalf("Error in Request: %v\n", err)
	}

	log.Printf("Published [%s] : '%s'\n", subj, payload)
	log.Printf("Received [%v] : '%s'\n", msg.Subject, string(msg.Data))
}
