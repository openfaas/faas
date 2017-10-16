package cluster

import (
	"regexp"
	"time"

	"github.com/Shopify/sarama"
)

var minVersion = sarama.V0_9_0_0

type ConsumerMode uint8

const (
	ConsumerModeMultiplex ConsumerMode = iota
	ConsumerModePartitions
)

// Config extends sarama.Config with Group specific namespace
type Config struct {
	sarama.Config

	// Group is the namespace for group management properties
	Group struct {

		// The strategy to use for the allocation of partitions to consumers (defaults to StrategyRange)
		PartitionStrategy Strategy

		// By default, messages and errors from the subscribed topics and partitions are all multiplexed and
		// made available through the consumer's Messages() and Errors() channels.
		//
		// Users who require low-level access can enable ConsumerModePartitions where individual partitions
		// are exposed on the Partitions() channel. Messages and errors must then be consumed on the partitions
		// themselves.
		Mode ConsumerMode

		Offsets struct {
			Retry struct {
				// The numer retries when committing offsets (defaults to 3).
				Max int
			}
			Synchronization struct {
				// The duration allowed for other clients to commit their offsets before resumption in this client, e.g. during a rebalance
				// NewConfig sets this to the Consumer.MaxProcessingTime duration of the Sarama configuration
				DwellTime time.Duration
			}
		}

		Session struct {
			// The allowed session timeout for registered consumers (defaults to 30s).
			// Must be within the allowed server range.
			Timeout time.Duration
		}

		Heartbeat struct {
			// Interval between each heartbeat (defaults to 3s). It should be no more
			// than 1/3rd of the Group.Session.Timout setting
			Interval time.Duration
		}

		// Return specifies which group channels will be populated. If they are set to true,
		// you must read from the respective channels to prevent deadlock.
		Return struct {
			// If enabled, rebalance notification will be returned on the
			// Notifications channel (default disabled).
			Notifications bool
		}

		Topics struct {
			// An additional whitelist of topics to subscribe to.
			Whitelist *regexp.Regexp
			// An additional blacklist of topics to avoid. If set, this will precede over
			// the Whitelist setting.
			Blacklist *regexp.Regexp
		}

		Member struct {
			// Custom metadata to include when joining the group. The user data for all joined members
			// can be retrieved by sending a DescribeGroupRequest to the broker that is the
			// coordinator for the group.
			UserData []byte
		}
	}
}

// NewConfig returns a new configuration instance with sane defaults.
func NewConfig() *Config {
	c := &Config{
		Config: *sarama.NewConfig(),
	}
	c.Group.PartitionStrategy = StrategyRange
	c.Group.Offsets.Retry.Max = 3
	c.Group.Offsets.Synchronization.DwellTime = c.Consumer.MaxProcessingTime
	c.Group.Session.Timeout = 30 * time.Second
	c.Group.Heartbeat.Interval = 3 * time.Second
	c.Config.Version = minVersion
	return c
}

// Validate checks a Config instance. It will return a
// sarama.ConfigurationError if the specified values don't make sense.
func (c *Config) Validate() error {
	if c.Group.Heartbeat.Interval%time.Millisecond != 0 {
		sarama.Logger.Println("Group.Heartbeat.Interval only supports millisecond precision; nanoseconds will be truncated.")
	}
	if c.Group.Session.Timeout%time.Millisecond != 0 {
		sarama.Logger.Println("Group.Session.Timeout only supports millisecond precision; nanoseconds will be truncated.")
	}
	if c.Group.PartitionStrategy != StrategyRange && c.Group.PartitionStrategy != StrategyRoundRobin {
		sarama.Logger.Println("Group.PartitionStrategy is not supported; range will be assumed.")
	}
	if !c.Version.IsAtLeast(minVersion) {
		sarama.Logger.Println("Version is not supported; 0.9. will be assumed.")
		c.Version = minVersion
	}
	if err := c.Config.Validate(); err != nil {
		return err
	}

	// validate the Group values
	switch {
	case c.Group.Offsets.Retry.Max < 0:
		return sarama.ConfigurationError("Group.Offsets.Retry.Max must be >= 0")
	case c.Group.Offsets.Synchronization.DwellTime <= 0:
		return sarama.ConfigurationError("Group.Offsets.Synchronization.DwellTime must be > 0")
	case c.Group.Offsets.Synchronization.DwellTime > 10*time.Minute:
		return sarama.ConfigurationError("Group.Offsets.Synchronization.DwellTime must be <= 10m")
	case c.Group.Heartbeat.Interval <= 0:
		return sarama.ConfigurationError("Group.Heartbeat.Interval must be > 0")
	case c.Group.Session.Timeout <= 0:
		return sarama.ConfigurationError("Group.Session.Timeout must be > 0")
	case !c.Metadata.Full && c.Group.Topics.Whitelist != nil:
		return sarama.ConfigurationError("Metadata.Full must be enabled when Group.Topics.Whitelist is used")
	case !c.Metadata.Full && c.Group.Topics.Blacklist != nil:
		return sarama.ConfigurationError("Metadata.Full must be enabled when Group.Topics.Blacklist is used")
	}

	// ensure offset is correct
	switch c.Consumer.Offsets.Initial {
	case sarama.OffsetOldest, sarama.OffsetNewest:
	default:
		return sarama.ConfigurationError("Consumer.Offsets.Initial must be either OffsetOldest or OffsetNewest")
	}

	return nil
}
