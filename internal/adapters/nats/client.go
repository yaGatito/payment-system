package natsadp

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsClient interface {
	Publish(ctx context.Context, subject string, data []byte) error
	Subscribe(ctx context.Context, streamName, durableName string, handler func(msg jetstream.Msg)) (jetstream.ConsumeContext, error)
	Close()
}

var _ NatsClient = (*NatsPubSubClient)(nil)

type NatsPubSubClient struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func New(url string, streamName string) (*NatsPubSubClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, err
	}

	client := &NatsPubSubClient{nc: nc, js: js}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"events.*"},
	})
	if err != nil {
		nc.Close()
		return nil, err
	}

	return client, nil
}

func (c *NatsPubSubClient) Subscribe(ctx context.Context, streamName, durableName string, handler func(msg jetstream.Msg)) (jetstream.ConsumeContext, error) {
	consumer, err := c.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:   durableName,
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, err
	}

	consumeCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		handler(msg)
		msg.Ack()
	})
	if err != nil {
		return nil, err
	}

	return consumeCtx, nil
}

func (c *NatsPubSubClient) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := c.js.Publish(ctx, subject, data)
	return err
}

func (c *NatsPubSubClient) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}
