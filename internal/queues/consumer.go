package queues

import (
	"context"

	common "github.com/arraial/pipo-dispatcher/internal/common"
	models "github.com/arraial/pipo-dispatcher/models"
	mq "github.com/wagslane/go-rabbitmq"
)

func createConsumer(conn *mq.Conn, concurrency int) (*mq.Consumer, error) {
	var log = common.GetLogger()
	var config = common.GetConfig()

	consumer, err := mq.NewConsumer(
		conn,
		config.GetString("queue.service.dispatcher.queue"),
		mq.WithConsumerOptionsConcurrency(concurrency),
		mq.WithConsumerOptionsQueueArgs(config.GetStringMap("queue.service.dispatcher.args")),
		mq.WithConsumerOptionsQueueDurable,
		mq.WithConsumerOptionsLogging,
		mq.WithConsumerOptionsLogger(mq.Logger(log)),
	)
	if err != nil {
		log.Fatalw("Unable to create consumer", "error", err)
	}
	log.Debug("Consumer created")
	return consumer, err
}

func StartConsumer(ctx context.Context, connection *mq.Conn, messages chan *models.ProviderOperation, consumers int) (*mq.Consumer, error) {
	log := common.GetLogger()
	consumer, consumerErr := createConsumer(connection, consumers)
	if consumerErr != nil {
		log.Errorw("Error creating consumer", "error", consumerErr)
		return nil, consumerErr
	}
	go func(ctx context.Context, cons *mq.Consumer) {
		// blocking call
		err := cons.Run(func(d mq.Delivery) mq.Action {
			if d.Body == nil {
				log.Errorw("Received empty message")
				return mq.NackDiscard
			}
			// TODO implement logic to create messages

			// mq.Ack, mq.NackDiscard, mq.NackRequeue
			return mq.Ack
		})
		if err != nil {
			log.Errorw("Error running consumer", "error", err)
			return
		}
	}(ctx, consumer)
	return consumer, consumerErr
}
