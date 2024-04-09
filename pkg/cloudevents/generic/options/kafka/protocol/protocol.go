package kafka_confluent

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	cecontext "github.com/cloudevents/sdk-go/v2/context"
)

//TODO: the implementation will be removed once the this pr is merged: https://github.com/cloudevents/sdk-go/pull/988

var (
	_ protocol.Sender   = (*Protocol)(nil)
	_ protocol.Opener   = (*Protocol)(nil)
	_ protocol.Receiver = (*Protocol)(nil)
	_ protocol.Closer   = (*Protocol)(nil)
)

type Protocol struct {
	kafkaConfigMap *kafka.ConfigMap

	consumer             *kafka.Consumer
	consumerTopics       []string
	consumerRebalanceCb  kafka.RebalanceCb     // optional
	consumerPollTimeout  int                   // optional
	consumerErrorHandler func(err kafka.Error) //optional
	consumerMux          sync.Mutex

	producer             *kafka.Producer
	producerDeliveryChan chan kafka.Event // optional
	producerDefaultTopic string           // optional

	// receiver
	incoming chan *kafka.Message
}

func New(opts ...Option) (*Protocol, error) {
	p := &Protocol{
		consumerPollTimeout: 100,
		incoming:            make(chan *kafka.Message),
	}
	if err := p.applyOptions(opts...); err != nil {
		return nil, err
	}

	if p.kafkaConfigMap != nil {
		if p.consumerTopics != nil && p.consumer == nil {
			consumer, err := kafka.NewConsumer(p.kafkaConfigMap)
			if err != nil {
				return nil, err
			}
			p.consumer = consumer
		}
		if p.producerDefaultTopic != "" && p.producer == nil {
			producer, err := kafka.NewProducer(p.kafkaConfigMap)
			if err != nil {
				return nil, err
			}
			p.producer = producer
			p.producerDeliveryChan = make(chan kafka.Event)
		}
		if p.producer == nil && p.consumer == nil {
			return nil, fmt.Errorf("at least set one of the receiver or sender topic")
		}
	}

	if p.kafkaConfigMap == nil && p.producer == nil && p.consumer == nil {
		return nil, fmt.Errorf("at least one of the following to initialize the protocol: config, producer, or consumer")
	}

	return p, nil
}

func (p *Protocol) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(p); err != nil {
			return err
		}
	}
	return nil
}

func (p *Protocol) Send(ctx context.Context, in binding.Message, transformers ...binding.Transformer) (err error) {
	if p.producer == nil {
		return fmt.Errorf("the producer client must not be nil")
	}
	defer func() {
		_ = in.Finish(err)
	}()

	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.producerDefaultTopic,
			Partition: kafka.PartitionAny,
		},
	}

	if topic := cecontext.TopicFrom(ctx); topic != "" {
		kafkaMsg.TopicPartition.Topic = &topic
	}

	if messageKey := MessageKeyFrom(ctx); messageKey != "" {
		kafkaMsg.Key = []byte(messageKey)
	}

	err = WriteProducerMessage(ctx, in, kafkaMsg, transformers...)
	if err != nil {
		return err
	}

	err = p.producer.Produce(kafkaMsg, p.producerDeliveryChan)
	if err != nil {
		return err
	}
	e := <-p.producerDeliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}
	return nil
}

func (p *Protocol) OpenInbound(ctx context.Context) error {
	if p.consumer == nil {
		return fmt.Errorf("the consumer client must not be nil")
	}
	if p.consumerTopics == nil {
		return fmt.Errorf("the consumer topics must not be nil")
	}

	p.consumerMux.Lock()
	defer p.consumerMux.Unlock()
	logger := cecontext.LoggerFrom(ctx)

	// Query committed offsets for each partition
	if positions := TopicPartitionOffsetsFrom(ctx); positions != nil {
		if err := p.consumer.Assign(positions); err != nil {
			return err
		}
	}

	logger.Infof("Subscribing to topics: %v", p.consumerTopics)
	err := p.consumer.SubscribeTopics(p.consumerTopics, p.consumerRebalanceCb)
	if err != nil {
		return err
	}

	run := true
	for run {
		select {
		case <-ctx.Done():
			run = false
		default:
			ev := p.consumer.Poll(p.consumerPollTimeout)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				p.incoming <- e
			case kafka.Error:
				// Errors should generally be considered informational, the client will try to automatically recover.
				// But in here, we choose to terminate the application if all brokers are down.
				if p.consumerErrorHandler != nil {
					p.consumerErrorHandler(e)
				}
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}

	logger.Infof("Closing consumer %v", p.consumerTopics)
	return p.consumer.Close()
}

// Receive implements Receiver.Receive
func (p *Protocol) Receive(ctx context.Context) (binding.Message, error) {
	select {
	case m, ok := <-p.incoming:
		if !ok {
			return nil, io.EOF
		}
		msg := NewMessage(m)
		return msg, nil
	case <-ctx.Done():
		return nil, io.EOF
	}
}

func (p *Protocol) Close(ctx context.Context) error {
	if p.consumer != nil {
		return p.consumer.Close()
	}
	if p.producer != nil {
		p.producer.Close()
	}
	close(p.producerDeliveryChan)
	close(p.incoming)
	return nil
}
