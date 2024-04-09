package kafka

import (
	"context"
	"fmt"

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

	topicCtx := cloudeventscontext.WithTopic(ctx, o.Topics.AgentEvents)
	if eventType.Action == types.ResyncRequestAction {
		return confluent.WithMessageKey(topicCtx, o.clusterName), nil
	}

	// agent publishes event to status topic to send the resource status from a specified cluster
	originalSource, err := evtCtx.GetExtension(types.ExtensionOriginalSource)
	if err != nil {
		return nil, err
	}

	messageKey := fmt.Sprintf("%s@%s", originalSource, o.clusterName)
	return confluent.WithMessageKey(topicCtx, messageKey), nil
}

func (o *kafkaAgentOptions) Client(ctx context.Context) (cloudevents.Client, error) {
	c, err := o.GetCloudEventsClient(
		confluent.WithConfigMap(o.ConfigMap),
		confluent.WithReceiverTopics([]string{o.Topics.SourceEvents}),
		confluent.WithSenderTopic(o.Topics.AgentEvents),
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
