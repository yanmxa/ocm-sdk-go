package kafka

import (
	"fmt"
	"os"

	confluent "github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2"
	"gopkg.in/yaml.v2"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	// sourceEventsTopic is a topic for sources to publish their resource create/update/delete events, the first
	// asterisk is a wildcard for source, the second asterisk is a wildcard for cluster.
	sourceEventsTopic = "sourceevents.*.*"
	// agentEventsTopic is a topic for agents to publish their resource status update events, the first
	// asterisk is a wildcard for source, the second asterisk is a wildcard for cluster.
	agentEventsTopic = "agentevents.*.*"
	// sourceBroadcastTopic is for a source to publish its events to all agents, the asterisk is a wildcard for source.
	sourceBroadcastTopic = "sourcebroadcast.*"
	// agentBroadcastTopic is for a agent to publish its events to all sources, the asterisk is a wildcard for cluster.
	agentBroadcastTopic = "agentbroadcast.*"
)

type KafkaOptions struct {
	// TODO don't use this directly
	// We may only need the necessary configurations, e.g. bootstrap.servers, ssl.ca.location, ssl.certificate.location and ssl.key.location
	// We should also have a fixed group.id
	// the configMap: https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
	ConfigMap *kafka.ConfigMap `json:"configs,omitempty" yaml:"configs,omitempty"`
}

func (o *KafkaOptions) GetCloudEventsClient(clientOpts ...confluent.Option) (cloudevents.Client, error) {
	protocol, err := confluent.New(clientOpts...)
	if err != nil {
		return nil, err
	}
	// // 	// TODO: we haven't monitor the producer.Events() chan, enable it might cause memory leak.
	// // enable it along with protocol.Close() to make sure the monitor is released
	// _ = opts.ConfigMap.SetKey("go.delivery.reports", false)
	// protocol.Events()
	return cloudevents.NewClient(protocol)
}

// BuildKafkaOptionsFromFlags builds configs from a config filepath.
func BuildKafkaOptionsFromFlags(configPath string) (*KafkaOptions, error) {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var opts KafkaOptions
	if err := yaml.Unmarshal(configData, &opts); err != nil {
		return nil, err
	}

	if opts.ConfigMap == nil {
		return nil, fmt.Errorf("the configs should be set")
	}

	val, err := opts.ConfigMap.Get("bootstrap.servers", "")
	if err != nil {
		return nil, err
	}
	if val == "" {
		return nil, fmt.Errorf("bootstrap.servers is required")
	}

	options := &KafkaOptions{
		ConfigMap: opts.ConfigMap,
	}
	return options, nil
}
