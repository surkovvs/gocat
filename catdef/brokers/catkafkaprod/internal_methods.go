package catkafkaprod

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func (prod *producer) sendAsyncMsg(msg contextedMsg) error {
	select {
	case <-msg.ctx.Done():
		return msg.ctx.Err()
	case prod.msgs <- msg:
		return nil
	case <-prod.stopped:
		return ErrProducerIsStopped
	}
}

func (prod *producer) sendSyncMsg(msg contextedMsg) error {
	select {
	case <-msg.ctx.Done():
		return msg.ctx.Err()
	case prod.msgs <- msg:
		return <-prod.errChan
	case <-prod.stopped:
		return ErrProducerIsStopped
	}
}

func (prod *producer) produce(ctx context.Context, key, value []byte) error {
	if prod.topic == nil {
		return errors.New("topic has not been setted, set it in config, or use SyncProduceInTopic, SyncProduceInTopicPartition")
	}
	if prod.partitioner == nil {
		return errors.New("function for partition seleting has not been setted, set it with options or use SyncProduceInPartition, SyncProduceInTopicPartition")
	}

	part, err := prod.partitioner.PartHash(key)
	if err != nil {
		return err
	}
	ctxMsg := contextedMsg{
		ctx: ctx,
		msg: &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     prod.topic,
				Partition: int32(part % *prod.partNum[*prod.topic]),
			},
			Value: value,
			Key:   key,
		},
	}

	if prod.sync {
		return prod.sendSyncMsg(ctxMsg)
	}
	return prod.sendAsyncMsg(ctxMsg)
}

func (prod *producer) produceInPartition(ctx context.Context, part int32, key, value []byte) error {
	if prod.topic == nil {
		return errors.New("topic has not been setted, set it in config, or use SyncProduceInTopic, SyncProduceInTopicPartition")
	}
	ctxMsg := contextedMsg{
		ctx: ctx,
		msg: &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     prod.topic,
				Partition: part,
			},
			Value: key,
			Key:   value,
		},
	}

	if prod.sync {
		return prod.sendSyncMsg(ctxMsg)
	}
	return prod.sendAsyncMsg(ctxMsg)
}

func (prod *producer) produceInTopic(ctx context.Context, topic string, key, value []byte) error {
	if prod.partitioner == nil {
		return errors.New("function for partition seleting has not been setted, set it with options or use SyncProduceInPartition, SyncProduceInTopicPartition")
	}

	part, err := prod.partitioner.PartHash(key)
	if err != nil {
		return err
	}
	ctxMsg := contextedMsg{
		ctx: ctx,
		msg: &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: int32(part % *prod.partNum[topic]),
			},
			Value: value,
			Key:   key,
		},
	}

	if prod.sync {
		return prod.sendSyncMsg(ctxMsg)
	}
	return prod.sendAsyncMsg(ctxMsg)
}

func (prod *producer) produceInTopicPartition(ctx context.Context, topic string, part int32, key, value []byte) error {
	ctxMsg := contextedMsg{
		ctx: ctx,
		msg: &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: part,
			},
			Value: value,
			Key:   key,
		},
	}

	if prod.sync {
		return prod.sendSyncMsg(ctxMsg)
	}
	return prod.sendAsyncMsg(ctxMsg)
}

func (prod *producer) stop() error {
	if atomic.CompareAndSwapUint32(&prod.status, 0, 1) {
		toClose := prod.msgs
		prod.msgs = nil
		close(toClose)
		close(prod.stopped)
		return nil
	}
	return errors.New("producer already stopped")
}

func (prod *producer) isStopped() bool {
	return atomic.LoadUint32(&prod.status) == 1
}
