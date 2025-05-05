package main

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func (cons *consumer) defaultRB(c *kafka.Consumer, event kafka.Event) error {
	switch ev := event.(type) {
	case kafka.AssignedPartitions:
		cons.logger.Info("rebalance, assigned",
			"protocol", c.GetRebalanceProtocol(),
			"partitions_num", len(ev.Partitions),
			"partitions", ev.Partitions,
		)

		// The application may update the start .Offset of each assigned
		// partition and then call Assign(). It is optional to call Assign
		// in case the application is not modifying any start .Offsets. In
		// that case we don't, the library takes care of it.
		// It is called here despite not modifying any .Offsets for illustrative
		// purposes.
		err := c.Assign(ev.Partitions)
		if err != nil {
			return err
		}

	case kafka.RevokedPartitions:
		cons.logger.Info("rebalance, revoked",
			"protocol", c.GetRebalanceProtocol(),
			"partitions_num", len(ev.Partitions),
			"partitions", ev.Partitions,
		)

		// Usually, the rebalance callback for `RevokedPartitions` is called
		// just before the partitions are revoked. We can be certain that a
		// partition being revoked is not yet owned by any other consumer.
		// This way, logic like storing any pending offsets or committing
		// offsets can be handled.
		// However, there can be cases where the assignment is lost
		// involuntarily. In this case, the partition might already be owned
		// by another consumer, and operations including committing
		// offsets may not work.
		if c.AssignmentLost() {
			// Our consumer has been kicked out of the group and the
			// entire assignment is thus lost.
			cons.logger.Warn("Assignment lost involuntarily, commit may fail")
		}

		// Since enable.auto.commit is unset, we need to commit offsets manually
		// before the partition is revoked.
		commitedOffsets, err := c.Commit()

		if err != nil && err.(kafka.Error).Code() != kafka.ErrNoOffset {
			cons.logger.Error("failed to commit offsets",
				"error", err,
			)
			return err
		}
		cons.logger.Info("commited",
			"offsets", commitedOffsets,
		)

		// Similar to Assign, client automatically calls Unassign() unless the
		// callback has already called that method. Here, we don't call it.

	default:
		cons.logger.Warn("unexpected event type",
			"event", event,
		)
	}

	return nil
}
