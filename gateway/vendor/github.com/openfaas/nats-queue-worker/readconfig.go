package main

import (
	"log"
	"os"
	"strconv"
	"time"
)

// ReadConfig constitutes config from env variables
type ReadConfig struct {
}

func (ReadConfig) Read() QueueWorkerConfig {
	cfg := QueueWorkerConfig{
		AckWait:     time.Second * 30,
		MaxInflight: 1,
	}

	if val, exists := os.LookupEnv("faas_nats_address"); exists {
		cfg.NatsAddress = val
	} else {
		cfg.NatsAddress = "nats"
	}

	if val, exists := os.LookupEnv("faas_gateway_address"); exists {
		cfg.GatewayAddress = val
	} else {
		cfg.GatewayAddress = "gateway"
	}

	if val, exists := os.LookupEnv("faas_function_suffix"); exists {
		cfg.FunctionSuffix = val
	}

	if val, exists := os.LookupEnv("faas_print_body"); exists {
		if val == "1" || val == "true" {
			cfg.DebugPrintBody = true
		} else {
			cfg.DebugPrintBody = false
		}
	}

	if val, exists := os.LookupEnv("write_debug"); exists {
		if val == "1" || val == "true" {
			cfg.WriteDebug = true
		} else {
			cfg.WriteDebug = false
		}
	}

	if value, exists := os.LookupEnv("max_inflight"); exists {
		val, err := strconv.Atoi(value)
		if err != nil {
			log.Println("max_inflight error:", err)
		} else {
			cfg.MaxInflight = val
		}
	}

	if val, exists := os.LookupEnv("ack_wait"); exists {
		ackWaitVal, durationErr := time.ParseDuration(val)
		if durationErr != nil {
			log.Println("ack_wait error:", durationErr)
		} else {
			cfg.AckWait = ackWaitVal
		}
	}

	return cfg
}

type QueueWorkerConfig struct {
	NatsAddress    string
	GatewayAddress string
	FunctionSuffix string
	DebugPrintBody bool
	WriteDebug     bool
	MaxInflight    int
	AckWait        time.Duration
}
