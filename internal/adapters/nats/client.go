package natsadp

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const createOrUpdateStreamTimeout = 10 * time.Second

type NatsClient interface {
	Publish(ctx context.Context, subject string, data []byte) error
	Subscribe(
		ctx context.Context,
		streamName, durableName string,
		handler func(msg jetstream.Msg),
	) (jetstream.ConsumeContext, error)
	Close()
}

var _ NatsClient = (*NatsPubSubClient)(nil)

type NatsPubSubClient struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func New(url string, streamName string) (*NatsPubSubClient, error) {
	if url == "" {
		return nil, fmt.Errorf("nats url is required")
	}
	if streamName == "" {
		return nil, fmt.Errorf("nats stream name is required")
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("init jetstream: %w", err)
	}

	client := &NatsPubSubClient{nc: nc, js: js}

	ctx, cancel := context.WithTimeout(context.Background(), createOrUpdateStreamTimeout)
	defer cancel()

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"events.*"},
	})
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("create stream %s: %w", streamName, err)
	}

	return client, nil
}

func (c *NatsPubSubClient) Subscribe(
	ctx context.Context,
	streamName, durableName string,
	handler func(msg jetstream.Msg),
) (jetstream.ConsumeContext, error) {
	if streamName == "" {
		return nil, fmt.Errorf("stream name is required")
	}
	if durableName == "" {
		return nil, fmt.Errorf("consumer durable name is required")
	}
	if c == nil || c.js == nil {
		return nil, fmt.Errorf("nats client is not initialized")
	}

	consumer, err := c.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:   durableName,
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, err
	}

	consumeCtx, err := consumer.Consume(handler)
	if err != nil {
		return nil, err
	}

	return consumeCtx, nil
}

func (c *NatsPubSubClient) Publish(ctx context.Context, subject string, data []byte) error {
	if c == nil || c.js == nil {
		return fmt.Errorf("nats client is not initialized")
	}
	if subject == "" {
		return fmt.Errorf("subject is required")
	}
	if len(data) == 0 {
		return fmt.Errorf("event payload is required")
	}
	_, err := c.js.Publish(ctx, subject, data)
	return err
}

func (c *NatsPubSubClient) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}
