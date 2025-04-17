package queues

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	common "github.com/arraial/pipo-dispatcher/internal/common"
)

func CreatePublisher() (publisher *amqp.Publisher, err error) {
	log := common.GetLogger()

	publisher, err = amqp.NewPublisher(publisherConfiguration(), NewAdaptedZap(log.Desugar()))
	if err != nil {
		log.Fatalw("Unable to create publisher", "error", err)
	}

	log.Info("Publisher created")
	return
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
