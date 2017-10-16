// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
	"strings"
)

// OsEnv implements interface to wrap os.Getenv
type OsEnv struct {
}

// Getenv wraps os.Getenv
func (OsEnv) Getenv(key string) string {
	return os.Getenv(key)
}

// HasEnv provides interface for os.Getenv
type HasEnv interface {
	Getenv(key string) string
}

// ReadConfig constitutes config from env variables
type ReadConfig struct {
}

func parseBoolValue(val string) bool {
	if val == "true" {
		return true
	}
	return false
}

func parseIntValue(val string, fallback int) int {
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return parsedVal
		}
	}
	return fallback
}

// Read fetches config from environmental variables.
func (ReadConfig) Read(hasEnv HasEnv) GatewayConfig {
	cfg := GatewayConfig{
		PrometheusHost: "prometheus",
		PrometheusPort: 9090,
	}

	readTimeout := parseIntValue(hasEnv.Getenv("read_timeout"), 8)
	writeTimeout := parseIntValue(hasEnv.Getenv("write_timeout"), 8)

	cfg.ReadTimeout = time.Duration(readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(writeTimeout) * time.Second

	if len(hasEnv.Getenv("functions_provider_url")) > 0 {
		var err error
		cfg.FunctionsProviderURL, err = url.Parse(hasEnv.Getenv("functions_provider_url"))
		if err != nil {
			log.Fatal("If functions_provider_url is provided, then it should be a valid URL.", err)
		}
	}

	faasNATSAddress := hasEnv.Getenv("faas_nats_address")
	if len(faasNATSAddress) > 0 {
		cfg.NATSAddress = &faasNATSAddress
	}

	faasNATSPort := hasEnv.Getenv("faas_nats_port")
	if len(faasNATSPort) > 0 {
		port, err := strconv.Atoi(faasNATSPort)
		if err == nil {
			cfg.NATSPort = &port
		} else {
			log.Println("faas_nats_port invalid number: " + faasNATSPort)
		}
	}

	faasKafkaBrokers := hasEnv.Getenv("faas_kafka_brokers")
	if len(faasKafkaBrokers) > 0 {
		brokers := strings.Split(faasKafkaBrokers,",")
		//cleanup
		for i:=0;i<len(brokers); {
			if len(brokers[i])==0 {
				brokers=append(brokers[:i],brokers[i+1:]...)
			} else {
				i++
			}
		}
		cfg.KafkaBrokers = &brokers
	}

	faasQueueTopics := hasEnv.Getenv("faas_queue_topics")
	if len(faasQueueTopics) > 0 {
		topics := strings.Split(faasQueueTopics,",")
		for i:=0;i<len(topics); {
			if len(topics[i])==0 {
				topics=append(topics[:i],topics[i+1:]...)
			} else {
				i++
			}
		}
		cfg.QueueTopics = &topics
	}

	prometheusPort := hasEnv.Getenv("faas_prometheus_port")
	if len(prometheusPort) > 0 {
		prometheusPortVal, err := strconv.Atoi(prometheusPort)
		if err != nil {
			log.Println("Invalid port for faas_prometheus_port")
		} else {
			cfg.PrometheusPort = prometheusPortVal
		}
	}

	prometheusHost := hasEnv.Getenv("faas_prometheus_host")
	if len(prometheusHost) > 0 {
		cfg.PrometheusHost = prometheusHost
	}

	return cfg
}

// GatewayConfig for the process.
type GatewayConfig struct {
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	FunctionsProviderURL *url.URL
	NATSAddress          *string
	NATSPort             *int
	KafkaBrokers         *[]string
	QueueTopics          *[]string
	PrometheusHost       string
	PrometheusPort       int
}

// UseNATS Use NATSor not
func (g *GatewayConfig) UseNATS() bool {
	return g.NATSPort != nil &&
		g.NATSAddress != nil
}

// Use Kafka or not
func (g *GatewayConfig) UseKafka() bool {
	return g.KafkaBrokers != nil &&
		g.QueueTopics != nil
}

// UseExternalProvider decide whether to bypass built-in Docker Swarm engine
func (g *GatewayConfig) UseExternalProvider() bool {
	return g.FunctionsProviderURL != nil
}
