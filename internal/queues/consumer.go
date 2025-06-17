package queues

import (
	amqp "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	common "github.com/arraial/pipo-dispatcher/internal/common"
)

func CreateConsumer() (subscriber *amqp.Subscriber, err error) {
	log := common.GetLogger()

	subscriber, err = amqp.NewSubscriber(
		consumerConfiguration(),
		NewAdaptedZap(log.Desugar()),
	)
	if err != nil {
		log.Fatalw("Unable to create consumer", "error", err)
	}
	log.Info("Consumer created")
	return
}

func consumerConfiguration() amqp.Config {
	config := common.GetConfig()
	return amqp.Config{
		Connection: amqp.ConnectionConfig{
			AmqpURI: config.GetString("queue.broker.url"),
		},
		Marshaler: amqp.DefaultMarshaler{},
		Exchange: amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return config.GetString("queue.service.dispatcher.exchange.name")
			},
			Type:    config.GetString("queue.service.dispatcher.exchange.type"),
			Durable: config.GetBool("queue.service.dispatcher.exchange.durable"),
		},
		Queue: amqp.QueueConfig{
			GenerateName: amqp.GenerateQueueNameConstant(config.GetString("queue.service.dispatcher.queue.name")),
			Durable:      config.GetBool("queue.service.dispatcher.queue.durable"),
			Arguments:    config.GetStringMap("queue.service.dispatcher.queue.args"),
		},
		QueueBind: amqp.QueueBindConfig{
			GenerateRoutingKey: func(topic string) string {
				return config.GetString("queue.service.dispatcher.queue.routing_key")
			},
		},
		Consume: amqp.ConsumeConfig{
			Qos: amqp.QosConfig{
				PrefetchCount: 5,
			},
		},
		TopologyBuilder: &amqp.DefaultTopologyBuilder{},
	}
}
