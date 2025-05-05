package catkafkaprod

import (
	"context"
)

type ProdClient struct {
	topic     *string
	partition *int32
	prod      *producer
}

func (prod *producer) NewClient() ProdClient {
	return ProdClient{
		prod: prod,
	}
}

func (client ProdClient) StopProducer() error {
	return client.prod.stop()
}

func (client ProdClient) IsProducerStopped() bool {
	return client.prod.isStopped()
}

func (client ProdClient) Produce(ctx context.Context, key, value []byte) error {
	switch {
	case client.topic != nil && client.partition != nil:
		return client.prod.produceInTopicPartition(ctx, *client.topic, *client.partition, key, value)
	case client.topic != nil:
		return client.prod.produceInTopic(ctx, *client.topic, key, value)
	case client.partition != nil:
		return client.prod.produceInPartition(ctx, *client.partition, key, value)
	default:
		return client.prod.produce(ctx, key, value)
	}
}

func (client ProdClient) Topic(topic string) ProdClient {
	return ProdClient{
		topic:     &topic,
		partition: client.partition,
		prod:      client.prod,
	}
}

func (client ProdClient) Partition(part int32) ProdClient {
	return ProdClient{
		topic:     client.topic,
		partition: &part,
		prod:      client.prod,
	}
}
