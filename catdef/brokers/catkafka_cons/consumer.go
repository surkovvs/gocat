package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/surkovvs/gocat/catcfg"
	"github.com/surkovvs/gocat/catlog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"golang.org/x/sync/errgroup"
)

type consumer struct {
	logger  catlog.Logger
	topic   *string
	kcons   *kafka.Consumer
	cfg     *kafka.ConfigMap
	clients map[int32]ConsClient
}

func New(cfg *catcfg.Config, tag string) *consumer {
	consCfg := cfg.Kafka[tag]
	toDefault(consCfg)
	cons := &consumer{
		logger:  cfg.Logger,
		cfg:     &consCfg,
		clients: make(map[int32]ConsClient),
	}
	return cons
}

// TODO: implement auto commit
func toDefault(cfg kafka.ConfigMap) {
	cfg["enable.auto.commit"] = "false"
	cfg["enable.auto.offset.store"] = "false"
}

func (cons *consumer) Init(ctx context.Context) error {
	if err := cons.topicCatch(); err != nil {
		return fmt.Errorf("catching topic: %w", err)
	}

	kcons, err := kafka.NewConsumer(cons.cfg)
	if err != nil {
		return fmt.Errorf("new confluent consumer: %w", err)
	}
	cons.kcons = kcons

	if err := cons.logging(cons.kcons); err != nil {
		return fmt.Errorf("logging: %w", err)
	}

	if err := cons.fetchMeta(ctx); err != nil {
		return fmt.Errorf("fetch meta: %w", err)
	}

	return nil
}

func (cons *consumer) logging(kclient *kafka.Consumer) error {
	logEnabled, err := cons.cfg.Get("go.logs.channel.enable", false)
	if err != nil {
		return err
	}
	if logEnabled.(bool) {
		if cons.logger == nil {
			return errors.New("logger has not been setted")
		}

		logLevel := 7
		logLevelVal, err := cons.cfg.Get("log_level", nil)
		if err != nil {
			return err
		}
		if logLevelVal != nil {
			logLevel = logLevelVal.(int)
		}
		go func() {
			for event := range kclient.Logs() {
				if logLevel < event.Level {
					continue
				}
				switch event.Level {
				case 7:
					cons.logger.Debug(event.Message,
						"name", event.Name,
						"tag", event.Tag)
				case 5, 6:
					cons.logger.Info(event.Message,
						"name", event.Name,
						"tag", event.Tag)
				case 4:
					cons.logger.Warn(event.Message,
						"name", event.Name,
						"tag", event.Tag)
				case 0, 1, 2, 3:
					cons.logger.Error(event.Message,
						"name", event.Name,
						"tag", event.Tag)
				}
			}
		}()
	}

	return nil
}

func (cons *consumer) fetchMeta(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		meta, err := cons.kcons.GetMetadata(cons.topic, true, 500)
		if err != nil {
			return fmt.Errorf("get metadata: %w", err)
		}
		for _, part := range meta.Topics[*cons.topic].Partitions {
			if egCtx.Err() != nil {
				return err
			}

			low, high, err := cons.kcons.QueryWatermarkOffsets(*cons.topic, part.ID, 500)
			if err != nil {
				return fmt.Errorf("query offsets: %w", err)
			}

			kcons, err := kafka.NewConsumer(cons.cfg)
			if err != nil {
				return fmt.Errorf("new client consumer: %w", err)
			}

			if err := cons.logging(kcons); err != nil {
				return fmt.Errorf("logging client: %w", err)
			}

			if err := kcons.Assign([]kafka.TopicPartition{
				{
					Topic:     cons.topic,
					Partition: part.ID,
					Offset:    kafka.OffsetStored,
				},
			}); err != nil {
				return fmt.Errorf("client assign: %w", err)
			}

			committedTP, err := kcons.Committed([]kafka.TopicPartition{
				{
					Topic:     cons.topic,
					Partition: part.ID,
				},
			}, 500)
			if err != nil {
				return fmt.Errorf("client request committed: %w", err)
			}

			cons.clients[part.ID] = ConsClient{
				topic:            cons.topic,
				partition:        part.ID,
				kcons:            kcons,
				offserRemoteLow:  low,
				offsetRemoteHigh: high,
				offsetCurrent:    int64(committedTP[0].Offset),
			}
		}
		return nil
	})
	return eg.Wait()
}

func (cons *consumer) topicCatch() error {
	val, err := cons.cfg.Get("topic", nil)
	if err != nil {
		return err
	}
	if val != nil {
		defer delete(*cons.cfg, "topic")
		topic, ok := val.(string)
		if !ok {
			return errors.New("incorrect value for topic key, must be string")
		}
		cons.topic = &topic
		if cons.logger != nil {
			cons.logger.Debug("topic has been setted for consumer",
				"topic", cons.topic)
		}
	}
	return nil
}

func (cons consumer) Shutdown(ctx context.Context) error {
	var errs []error
	if err := cons.kcons.Close(); err != nil {
		errs = append(errs, err)
	}
	for _, client := range cons.clients {
		if err := client.kcons.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
