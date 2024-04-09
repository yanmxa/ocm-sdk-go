package kafka_confluent

import (
	"bytes"
	"context"
	"strconv"
	"strings"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

//TODO: the implementation will be removed once the this pr is merged: https://github.com/cloudevents/sdk-go/pull/988

const (
	prefix         = "ce-"
	contentTypeKey = "Content-Type"
)

// the following property will be added to the cloudevents extension by the receiver
const (
	KafkaOffsetKey    = "kafkaoffset"
	KafkaPartitionKey = "kafkapartition"
	KafkaTopicKey     = "kafkatopic"
	KafkaMessageKey   = "kafkamessagekey"
)

var specs = spec.WithPrefix(prefix)

// Message represents a Kafka message.
// This message *can* be read several times safely
type Message struct {
	internal   *kafka.Message
	properties map[string][]byte
	format     format.Format
	version    spec.Version
}

// Check if Message implements binding.Message
var (
	_ binding.Message               = (*Message)(nil)
	_ binding.MessageMetadataReader = (*Message)(nil)
)

func NewMessage(msg *kafka.Message) *Message {
	var contentType, contentVersion string
	properties := make(map[string][]byte, len(msg.Headers)+3)
	for _, header := range msg.Headers {
		k := strings.ToLower(string(header.Key))
		if k == strings.ToLower(contentTypeKey) {
			contentType = string(header.Value)
		}
		if k == specs.PrefixedSpecVersionName() {
			contentVersion = string(header.Value)
		}
		properties[k] = header.Value
	}

	// add the kafka message key, topic, partition and partition key to the properties
	properties[prefix+KafkaOffsetKey] = []byte(strconv.FormatInt(int64(msg.TopicPartition.Offset), 10))
	properties[prefix+KafkaPartitionKey] = []byte(strconv.FormatInt(int64(msg.TopicPartition.Partition), 10))
	properties[prefix+KafkaTopicKey] = []byte(*msg.TopicPartition.Topic)
	if msg.Key != nil {
		properties[prefix+KafkaMessageKey] = msg.Key
	}

	message := &Message{
		internal:   msg,
		properties: properties,
	}
	if ft := format.Lookup(contentType); ft != nil {
		message.format = ft
	} else if v := specs.Version(contentVersion); v != nil {
		message.version = v
	}

	return message
}

func (m *Message) ReadEncoding() binding.Encoding {
	if m.version != nil {
		return binding.EncodingBinary
	}
	if m.format != nil {
		return binding.EncodingStructured
	}
	return binding.EncodingUnknown
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.format != nil {
		return encoder.SetStructuredEvent(ctx, m.format, bytes.NewReader(m.internal.Value))
	}
	return binding.ErrNotStructured
}

func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) error {
	if m.version == nil {
		return binding.ErrNotBinary
	}

	var err error
	for k, v := range m.properties {
		if strings.HasPrefix(k, prefix) {
			attr := m.version.Attribute(k)
			if attr != nil {
				err = encoder.SetAttribute(attr, string(v))
			} else {
				err = encoder.SetExtension(strings.TrimPrefix(k, prefix), string(v))
			}
		} else if k == strings.ToLower(contentTypeKey) {
			err = encoder.SetAttribute(m.version.AttributeFromKind(spec.DataContentType), string(v))
		}
		if err != nil {
			return err
		}
	}

	if m.internal.Value != nil {
		err = encoder.SetData(bytes.NewBuffer(m.internal.Value))
	}
	return err
}

func (m *Message) Finish(error) error {
	return nil
}

func (m *Message) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	attr := m.version.AttributeFromKind(k)
	if attr == nil {
		return nil, nil
	}
	return attr, m.properties[attr.PrefixedName()]
}

func (m *Message) GetExtension(name string) interface{} {
	return m.properties[prefix+name]
}
