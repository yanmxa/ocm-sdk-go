package kafka

import (
	"context"
	"fmt"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudeventscontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	confluent "github.com/cloudevents/sdk-go/protocol/kafka_confluent/v2"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/options"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/types"
)

type kafkaAgentOptions struct {
	KafkaOptions
	clusterName string
	agentID     string
	errorChan   chan error
}

func NewAgentOptions(kafkaOptions *KafkaOptions, clusterName, agentID string) *options.CloudEventsAgentOptions {
	kafkaAgentOptions := &kafkaAgentOptions{
		KafkaOptions: *kafkaOptions,
		clusterName:  clusterName,
		agentID:      agentID,
		errorChan:    make(chan error),
	}

	return &options.CloudEventsAgentOptions{
		CloudEventsOptions: kafkaAgentOptions,
		AgentID:            agentID,
		ClusterName:        clusterName,
	}
}

// encode the source and agent to the message key
func (o *kafkaAgentOptions) WithContext(ctx context.Context, evtCtx cloudevents.EventContext) (context.Context, error) {
	eventType, err := types.ParseCloudEventsType(evtCtx.GetType())
	if err != nil {
		return nil, err
	}

	// agent publishes event to status topic to send the resource status from a specified cluster
	originalSource, err := evtCtx.GetExtension(types.ExtensionOriginalSource)
	if err != nil {
		return nil, err
	}

	if eventType.Action == types.ResyncRequestAction && originalSource == types.SourceAll {
		// TODO support multiple sources, agent may need a source list instead of the broadcast
		topic := strings.Replace(agentBroadcastTopic, "*", o.clusterName, 1)
		return confluent.WithMessageKey(cloudeventscontext.WithTopic(ctx, topic), o.clusterName), nil
	}

	topic := strings.Replace(agentEventsTopic, "*", fmt.Sprintf("%s", originalSource), 1)
	topic = strings.Replace(topic, "*", o.clusterName, 1)
	messageKey := fmt.Sprintf("%s@%s", originalSource, o.clusterName)
	return confluent.WithMessageKey(cloudeventscontext.WithTopic(ctx, topic), messageKey), nil
}

func (o *kafkaAgentOptions) Client(ctx context.Context) (cloudevents.Client, error) {
	cf := &kafka.ConfigMap{
		"bootstrap.servers":        "kafka-kafka-tls-bootstrap-amq-streams.apps.server-foundation-sno-lite-msgxk.dev04.red-chesterfield.com:443",
		"group.id":                 "myGroup",
		"auto.offset.reset":        "earliest",
		"security.protocol":        "SSL",
		"ssl.ca.location":          "/Users/liuwei/go/src/github.com/stolostron/maestro-addon/cluster.ca.pem",
		"ssl.certificate.location": "/Users/liuwei/go/src/github.com/stolostron/maestro-addon/admin-client.pem",
		"ssl.key.location":         "/Users/liuwei/go/src/github.com/stolostron/maestro-addon/admin-client-key.pem",
	}

	c, err := o.GetCloudEventsClient(
		confluent.WithConfigMap(cf),
		// TODO support multiple sources, agent may need a source list instead of the wildcard
		confluent.WithReceiverTopics([]string{
			fmt.Sprintf("^%s", replaceLast(sourceEventsTopic, "*", o.clusterName)),
			fmt.Sprintf("^%s", sourceBroadcastTopic),
		}),
		confluent.WithSenderTopic("agentevents"),
		confluent.WithErrorHandler(func(ctx context.Context, err kafka.Error) {
			o.errorChan <- err
		}),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (o *kafkaAgentOptions) ErrorChan() <-chan error {
	return o.errorChan
}

func replaceLast(str, old, new string) string {
	last := strings.LastIndex(str, old)
	if last == -1 {
		return str
	}
	return str[:last] + new + str[last+len(old):]
}
