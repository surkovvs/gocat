package catkafkaprod

import (
	"context"
	"errors"
	"fmt"

	"github.com/surkovvs/gocat/catcfg"
	"github.com/surkovvs/gocat/catlog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var ErrProducerIsStopped = errors.New("producer has been stopped")

type contextedMsg struct {
	ctx context.Context
	msg *kafka.Message
}

type producer struct {
	logger      catlog.Logger
	topic       *string
	partitioner partitioner
	partNum     map[string]*uint32
	cfg         *kafka.ConfigMap
	kprod       *kafka.Producer
	msgs        chan contextedMsg
	logs        chan kafka.LogEvent
	errChan     chan error
	stopped     chan struct{} // on closed means in stop process
	fushTimeout int
	status      uint32
	sync        bool
}

func New(cfg *catcfg.Config, tag string, sync bool, opts ...prodOptions) *producer {
	prodCfg := cfg.Kafka[tag]
	prod := &producer{
		logger:  cfg.Logger,
		cfg:     &prodCfg,
		msgs:    make(chan contextedMsg),
		errChan: make(chan error),
		stopped: make(chan struct{}),
		sync:    sync,
	}

	for _, opt := range opts {
		opt(prod)
	}

	return prod
}

func (prod *producer) Init(ctx context.Context) error {
	if err := prod.logging(); err != nil {
		return fmt.Errorf("logging: %w", err)
	}

	if err := prod.topicCatch(); err != nil {
		return fmt.Errorf("catching topic: %w", err)
	}

	if err := prod.flushTomeoutCatch(); err != nil {
		return fmt.Errorf("catching timeout for flush: %w", err)
	}

	producer, err := kafka.NewProducer(prod.cfg)
	if err != nil {
		return fmt.Errorf("new confluent producer: %w", err)
	}
	prod.kprod = producer

	if err := prod.fetchMeta(); err != nil {
		return fmt.Errorf("fetch meta: %w", err)
	}

	if prod.topic != nil {
		_, ok := prod.partNum[*prod.topic]
		if !ok {
			topics := make([]string, 0, len(prod.partNum))
			for name := range prod.partNum {
				topics = append(topics, name)
			}
			return fmt.Errorf("unexisted topic in cluster has been setted to producer, topcis in cluster: %v", topics)
		}
	}

	return nil
}

func (prod *producer) logging() error {
	logEnabled, err := prod.cfg.Get("go.logs.channel.enable", false)
	if err != nil {
		return err
	}
	if logEnabled.(bool) {
		if prod.logger == nil {
			return errors.New("logger has not been setted")
		}

		prod.logs = make(chan kafka.LogEvent)
		if err := prod.cfg.SetKey("go.logs.channel", prod.logs); err != nil {
			return err
		}

		logLevel := 7
		logLevelVal, err := prod.cfg.Get("log_level", nil)
		if err != nil {
			return err
		}
		if logLevelVal != nil {
			logLevel = logLevelVal.(int)
		}
		go func() {
			for event := range prod.logs {
				if logLevel < event.Level {
					continue
				}
				switch event.Level {
				case 7:
					prod.logger.Debug(event.Message)
				case 5, 6:
					prod.logger.Info(event.Message)
				case 4:
					prod.logger.Warn(event.Message)
				case 0, 1, 2, 3:
					prod.logger.Error(event.Message)
				}
			}
		}()
	}

	return nil
}

func (prod *producer) topicCatch() error {
	val, err := prod.cfg.Get("topic", nil)
	if err != nil {
		return err
	}
	if val != nil {
		defer delete(*prod.cfg, "topic")
		topic, ok := val.(string)
		if !ok {
			return errors.New("incorrect value for topic key, must be string")
		}
		prod.topic = &topic
		if prod.logger != nil {
			prod.logger.Debug("topic has been setted for producer",
				"topic", prod.topic)
		}
	}
	return nil
}

func (prod *producer) flushTomeoutCatch() error {
	if !prod.sync {
		val, err := prod.cfg.Get("flush.timeout.ms", nil)
		if err != nil {
			return err
		}
		if val != nil {
			fTo, ok := val.(int)
			if !ok {
				return errors.New("incorrect value for flush.timeout.ms key, must be int")
			}
			prod.fushTimeout = fTo
			if prod.logger != nil {
				prod.logger.Debug("timeout for flushing has been setted for producer",
					"timeout", prod.fushTimeout)
			}
		}
	}
	delete(*prod.cfg, "flush.timeout.ms")
	return nil
}

func (prod *producer) fetchMeta() error {
	meta, err := prod.kprod.GetMetadata(nil, true, 500)
	if err != nil {
		return fmt.Errorf("get metadata: %w", err)
	}
	prod.partNum = make(map[string]*uint32, len(meta.Topics))
	for name, topicMeta := range meta.Topics {
		if topicMeta.Error.IsFatal() ||
			topicMeta.Error.IsTimeout() ||
			topicMeta.Error.IsRetriable() {
			return fmt.Errorf("topic meta error: %w", topicMeta.Error)
		}
		pNum := uint32(len(topicMeta.Partitions))
		prod.partNum[name] = &pNum
	}
	return nil
}

func (prod *producer) Run(ctx context.Context) error {
	if prod.sync {
		return prod.syncRunning(ctx)
	}
	return prod.asyncRunning(ctx)
}

func (prod *producer) syncRunning(ctx context.Context) error {
	delivery := make(chan kafka.Event)
	for {
		select {
		case <-prod.stopped:
			if prod.logger != nil {
				prod.logger.Debug("producer has been stopped by stop method")
			}

			close(delivery)
			return nil
		case <-ctx.Done():
			if prod.logger != nil {
				prod.logger.Debug("producer has been stopped by ctx done in Run method")
			}

			close(delivery)
			return nil
		case ctxMsg := <-prod.msgs:
			if prod.isStopped() {
				continue
			}
			if err := prod.kprod.Produce(ctxMsg.msg, delivery); err != nil {
				return err
			}
			select {
			case <-ctxMsg.ctx.Done():
				prod.errChan <- ctxMsg.ctx.Err()
				prod.handleEventOnCtxDone(<-delivery)
			case event := <-delivery:
				prod.errChan <- event.(*kafka.Message).TopicPartition.Error
			}
		}
	}
}

// TODO: add in option as custom func
func (prod *producer) handleEventOnCtxDone(event kafka.Event) {
	msg := event.(*kafka.Message)

	if prod.logger != nil {
		prod.logger.Warn("handled event on sync message sending with exeeded timeout",
			"topic", msg.TopicPartition.Topic,
			"partition", msg.TopicPartition.Partition,
			"offset", msg.TopicPartition.Offset,
			"key", string(msg.Key),
			"error", msg.TopicPartition.Error)
	}
}

func (prod *producer) asyncRunning(ctx context.Context) error {
	go func() {
		for event := range prod.kprod.Events() {
			msg, ok := event.(*kafka.Message)
			if !ok {
				_ = prod.stop()
				if prod.logger != nil {
					prod.logger.Error("unexpected event type",
						"event_value", event.String())
				}
			}
			if msg.TopicPartition.Error != nil {
				if prod.logger != nil {
					prod.logger.Warn("failed to deliver message",
						"topic", msg.TopicPartition.Topic,
						"partition", msg.TopicPartition.Partition,
						"offset", msg.TopicPartition.Offset,
						"key", string(msg.Key),
						"error", msg.TopicPartition.Error)
				}
			} else {
				if prod.logger != nil {
					prod.logger.Debug("successfully produced record",
						"topic", msg.TopicPartition.Topic,
						"partition", msg.TopicPartition.Partition,
						"offset", msg.TopicPartition.Offset,
						"key", string(msg.Key))
				}
			}
		}
	}()

	lastUnflushed := 0
	for {
		select {
		case <-prod.stopped:
			if prod.logger != nil {
				prod.logger.Debug("producer has been stopped by stop method")
			}
			return nil
		case <-ctx.Done():
			if prod.logger != nil {
				prod.logger.Debug("producer has been stopped by ctx done in Run method")
			}
			return nil
		case ctxMsg := <-prod.msgs:
			if prod.isStopped() {
				continue
			}
			if err := prod.kprod.Produce(ctxMsg.msg, nil); err != nil {
				if err.Error() == kafka.ErrQueueFull.String() {
					unflushed := prod.kprod.Flush(prod.fushTimeout)
					if unflushed != 0 && prod.logger != nil {
						if unflushed > lastUnflushed {
							prod.logger.Info("got un-flushed events",
								"unflushed_curr", unflushed)
						} else {
							prod.logger.Warn("count of un-flushed events is growing",
								"unflushed_last", lastUnflushed,
								"unflushed_curr", unflushed)
						}
					}
					lastUnflushed = unflushed
					if err := prod.kprod.Produce(ctxMsg.msg, nil); err != nil {
						return err
					}
					continue
				}
				return err
			}
		}
	}
}

func (prod *producer) Shutdown(ctx context.Context) error {
	if prod.logger != nil {
		prod.logger.Debug("producer shutdown has been started")
	}
	unflushed := prod.kprod.Flush(1000)
	prod.kprod.Close()
	if prod.logs != nil {
		close(prod.logs)
	}
	if unflushed != 0 {
		return fmt.Errorf("un-flushed events num: %d", unflushed)
	}
	return nil
}
