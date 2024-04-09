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

type kafkaSourceOptions struct {
	KafkaOptions
	sourceID  string
	errorChan chan error
}

func NewSourceOptions(opts *KafkaOptions, sourceID string) *options.CloudEventsSourceOptions {
	sourceOptions := &kafkaSourceOptions{
		KafkaOptions: *opts,
		sourceID:     sourceID,
		errorChan:    make(chan error),
	}

	return &options.CloudEventsSourceOptions{
		CloudEventsOptions: sourceOptions,
		SourceID:           sourceID,
	}
}

func (o *kafkaSourceOptions) WithContext(ctx context.Context,
	evtCtx cloudevents.EventContext) (context.Context, error) {

	eventType, err := types.ParseCloudEventsType(evtCtx.GetType())
	if err != nil {
		return nil, err
	}

	topicCtx := cloudeventscontext.WithTopic(ctx, o.Topics.SourceEvents)
	if eventType.Action == types.ResyncRequestAction {
		return confluent.WithMessageKey(topicCtx, o.sourceID), nil
	}

	clusterName, err := evtCtx.GetExtension(types.ExtensionClusterName)
	if err != nil {
		return nil, err
	}

	// source publishes event to spec topic to send the resource spec to a specified cluster
	messageKey := fmt.Sprintf("%s@%s", o.sourceID, clusterName)
	return confluent.WithMessageKey(topicCtx, messageKey), nil
}

func (o *kafkaSourceOptions) Client(ctx context.Context) (cloudevents.Client, error) {
	c, err := o.GetCloudEventsClient(
		confluent.WithConfigMap(o.ConfigMap),
		confluent.WithReceiverTopics([]string{o.Topics.AgentEvents}),
		confluent.WithSenderTopic(o.Topics.SourceEvents),
		confluent.WithErrorHandler(func(ctx context.Context, err kafka.Error) {
			o.errorChan <- err
		}),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (o *kafkaSourceOptions) ErrorChan() <-chan error {
	return o.errorChan
}
