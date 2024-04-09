package kafka

import (
	"fmt"
	"os"

	confluent "github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2"
	"gopkg.in/yaml.v2"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/types"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	defaultSpecTopic   = "spec"
	defaultStatusTopic = "status"
)

type KafkaOptions struct {
	// the configMap: https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
	ConfigMap *kafka.ConfigMap `json:"configs,omitempty" yaml:"configs,omitempty"`
	Topics    *types.Topics    `json:"-" yaml:"-"`
}

func NewKafkaOptions() *KafkaOptions {
	return &KafkaOptions{
		Topics: &types.Topics{
			SourceEvents: defaultSpecTopic,
			AgentEvents:  defaultStatusTopic,
		},
	}
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
		Topics: &types.Topics{
			SourceEvents: defaultSpecTopic,
			AgentEvents:  defaultStatusTopic,
		},
	}
	return options, nil
}
