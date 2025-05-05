package main

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ConsClient struct {
	topic     *string
	partition int32
	kcons     *kafka.Consumer

	offserRemoteLow  int64
	offsetRemoteHigh int64
	offsetCurrent    int64
}

type offset int

const (
	Oldest offset = iota
	Current
	Latest
)

type commitBatch func() error

type message struct {
	topic     string
	partition int32
	key       []byte
	value     []byte
}

func (cons *consumer) InitPartitionClients(offset offset) ([]ConsClient, error) {
	var clients []ConsClient
	for _, client := range cons.clients {
		var toSet kafka.Offset
		switch offset {
		case Oldest:
			toSet = kafka.Offset(client.offserRemoteLow)
		case Current:
			toSet = kafka.Offset(client.offsetCurrent)
		case Latest:
			toSet = kafka.Offset(client.offsetRemoteHigh)
		}
		if _, err := client.kcons.CommitOffsets([]kafka.TopicPartition{
			{
				Topic:     client.topic,
				Partition: client.partition,
				Offset:    toSet,
			},
		}); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, nil
}

func (client ConsClient) GetBatchFromPartition(timeout time.Duration, size int) ([]message, commitBatch, error) {
	var messages []message
	var numToProcess int
	time.NewTimer(timeout)
	for i := 0; i < size; i++ {
		switch event := client.kcons.Poll(50).(type) {
		case *kafka.Message:
			messages = append(messages, message{
				topic:     *event.TopicPartition.Topic,
				partition: event.TopicPartition.Partition,
				key:       event.Key,
				value:     event.Value,
			})
		case kafka.Error:
			return nil, nil, event
		}
		numToProcess++
	}

	return messages, func() error {
		if _, err := client.kcons.CommitOffsets([]kafka.TopicPartition{
			{
				Topic:     client.topic,
				Partition: client.partition,
				Offset:    kafka.Offset(client.offsetCurrent + int64(numToProcess)),
			},
		}); err != nil {
			return err
		}
		return nil
	}, nil
}

// func (client ConsClient) SubscribePartitionWithOffset(ctx context.Context, part int32, offset int) error

// func (client ConsClient) SubscribeAll(ctx context.Context, part int32, offset int) error {
// 	client.cons.kcons.
// }
