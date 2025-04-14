package queues

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	amqpMessage "github.com/ThreeDotsLabs/watermill/message"
	common "github.com/arraial/pipo-dispatcher/internal/common"
	models "github.com/arraial/pipo-dispatcher/models"
)

func CreatePublisher(ctx context.Context, messages <-chan *models.ProviderOperation) (publisher *amqp.Publisher, err error) {
	log := common.GetLogger()

	publisher, err = amqp.NewPublisher(publisherConfiguration(), watermill.NewStdLogger(false, false))
	if err != nil {
		log.Fatalw("Unable to create publisher", "error", err)
	}

	go publishMessages(publisher, messages)
	log.Info("Publisher created")
	return
}

func publishMessages(pub amqpMessage.Publisher, messages <-chan *models.ProviderOperation) {
	log := common.GetLogger()
	log.Info("Starting publisher")
	for message := range messages {
		provider := fmt.Sprintf(
			"%s.%s",
			message.Provider,
			message.Operation,
		)
		data, err := json.Marshal(message)
		if err != nil {
			log.Errorw("Unable to marshall message", "error", err)
			continue
		}
		msg := amqpMessage.NewMessage(watermill.NewUUID(), data)
		log.Infow("Publishing message", "message", msg, "provider", provider)
		err = pub.Publish(provider, msg)
		if err != nil {
			log.Errorw("Unable to publish message", "error", err)
		}
	}
}

func publisherConfiguration() amqp.Config {
	config := common.GetConfig()
	return amqp.Config{
		Connection: amqp.ConnectionConfig{
			AmqpURI: config.GetString("queue.broker.url"),
		},
		Marshaler: amqp.DefaultMarshaler{},
		Exchange: amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return config.GetString("queue.service.transmuter.exchange.name")
			},
			Type:    config.GetString("queue.service.transmuter.exchange.type"),
			Durable: config.GetBool("queue.service.transmuter.exchange.durable"),
		},
		Publish: amqp.PublishConfig{
			GenerateRoutingKey: func(topic string) string {
				return fmt.Sprintf(
					"%s.%s",
					common.GetConfig().GetString("queue.service.transmuter.exchange.routing_key"),
					topic,
				)
			},
		},
		TopologyBuilder: &amqp.DefaultTopologyBuilder{},
	}
}
