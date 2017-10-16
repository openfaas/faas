package cluster

import "github.com/Shopify/sarama"

// Client is a group client
type Client struct {
	sarama.Client
	config Config
}

// NewClient creates a new client instance
func NewClient(addrs []string, config *Config) (*Client, error) {
	if config == nil {
		config = NewConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	client, err := sarama.NewClient(addrs, &config.Config)
	if err != nil {
		return nil, err
	}

	return &Client{Client: client, config: *config}, nil
}
