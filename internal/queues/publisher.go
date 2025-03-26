package queues

import (
	"context"
	"encoding/json"
	"fmt"

	common "github.com/arraial/pipo-dispatcher/internal/common"
	"github.com/arraial/pipo-dispatcher/models"
	mq "github.com/wagslane/go-rabbitmq"
)

func createPublisher(conn *mq.Conn) (*mq.Publisher, error) {

	var log = common.GetLogger()
	var config = common.GetConfig()

	publisher, err := mq.NewPublisher(
		conn,
		mq.WithPublisherOptionsExchangeName(config.GetString("queue.service.transmuter.exchange.name")),
		mq.WithPublisherOptionsExchangeKind(config.GetString("queue.service.transmuter.exchange.type")),
		mq.WithPublisherOptionsExchangeDurable,
		mq.WithPublisherOptionsExchangeDeclare,
		mq.WithPublisherOptionsLogging,
		mq.WithPublisherOptionsLogger(mq.Logger(log)),
	)
	if err != nil {
		log.Fatalw("Unable to create publisher", "error", err)
	}
	return publisher, err
}

func StartPublisher(ctx context.Context, conn *mq.Conn, messages chan *models.ProviderOperation) (*mq.Publisher, error) {
	log := common.GetLogger()
	config := common.GetConfig()
	publisher, err := createPublisher(conn)
	if err != nil {
		log.Errorw("Error creating publisher", "error", err)
		return nil, err
	}
	go func(ctx context.Context, cons *mq.Publisher) {
		defer publisher.Close()
		log.Info("Starting publisher")
		for {
			select {
			case message, ok := <-messages:
				if !ok {
					log.Info("Channel closed. Stopping publisher")
					return
				}
				provider := fmt.Sprintf(
					"%s.%s.%s",
					config.GetString("queue.service.transmuter.routing_key"),
					message.Provider,
					message.Operation,
				)
				data, publisherErr := json.Marshal(message)
				if publisherErr != nil {
					log.Errorw("Unable to marshall message", "error", publisherErr)
					return
				}
				publisherErr = publisher.PublishWithContext(ctx, data, []string{provider})
				if publisherErr != nil {
					log.Errorw("Unable to publish message", "error", publisherErr)
					return
				}
			case <-ctx.Done():
				log.Info("Terminating publisher")
				return
			}
		}
	}(ctx, publisher)
	return publisher, err
}
