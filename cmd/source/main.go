package main

import (
	"context"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	"open-cluster-management.io/sdk-go/cmd/resource"
	"open-cluster-management.io/sdk-go/cmd/signal"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/options/kafka"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/types"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/work/payload"
)

type resourceCodec struct{}

var _ generic.Codec[*resource.Resource] = &resourceCodec{}

func (c *resourceCodec) EventDataType() types.CloudEventsDataType {
	return payload.ManifestEventDataType
}

func (c *resourceCodec) Encode(source string, eventType types.CloudEventsType, resource *resource.Resource) (*cloudevents.Event, error) {
	if resource.Source != "" {
		source = resource.Source
	}

	if eventType.CloudEventsDataType != payload.ManifestEventDataType {
		return nil, fmt.Errorf("unsupported cloudevents data type %s", eventType.CloudEventsDataType)
	}

	eventBuilder := types.NewEventBuilder(source, eventType).
		WithResourceID(resource.ResourceID).
		WithResourceVersion(resource.ResourceVersion).
		WithClusterName(resource.Namespace)

	if !resource.GetDeletionTimestamp().IsZero() {
		evt := eventBuilder.WithDeletionTimestamp(resource.GetDeletionTimestamp().Time).NewEvent()
		return &evt, nil
	}

	evt := eventBuilder.NewEvent()

	if err := evt.SetData(cloudevents.ApplicationJSON, &payload.Manifest{Manifest: resource.Spec}); err != nil {
		return nil, fmt.Errorf("failed to encode manifests to cloud event: %v", err)
	}

	return &evt, nil
}

func (c *resourceCodec) Decode(evt *cloudevents.Event) (*resource.Resource, error) {
	return &resource.Resource{}, nil
}

func main() {
	shutdownCtx, cancel := context.WithCancel(context.TODO())
	shutdownHandler := signal.SetupSignalHandler()
	go func() {
		defer cancel()
		<-shutdownHandler
	}()

	ctx, terminate := context.WithCancel(shutdownCtx)
	defer terminate()

	kafkaOptions, err := kafka.BuildKafkaOptionsFromFlags("/Users/liuwei/kafka.source.config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	client, err := generic.NewCloudEventSourceClient[*resource.Resource](
		ctx,
		kafka.NewSourceOptions(kafkaOptions, "maestro"),
		&resource.ResourceLister{},
		resource.StatusHashGetter,
		&resourceCodec{},
	)
	if err != nil {
		log.Fatal(err)
	}

	client.Subscribe(ctx, func(action types.ResourceAction, res *resource.Resource) error {
		fmt.Printf("received resource %v", res)
		return nil
	})

	if err := client.Publish(ctx, types.CloudEventsType{
		CloudEventsDataType: payload.ManifestEventDataType,
		SubResource:         types.SubResourceSpec,
		Action:              types.EventAction("test_create_update_request"),
	}, &resource.Resource{Namespace: "cluster1"}); err != nil {
		log.Fatal(err)
	}

	if err := client.Resync(ctx, ""); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}
