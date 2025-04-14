package queues

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	amqp "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	audiosource "github.com/arraial/pipo-dispatcher/internal/audiosource"
	common "github.com/arraial/pipo-dispatcher/internal/common"
	models "github.com/arraial/pipo-dispatcher/models"
)

func CreateConsumer(ctx context.Context, ch chan *models.ProviderOperation) (subscriber *amqp.Subscriber, err error) {
	log := common.GetLogger()
	config := common.GetConfig()

	subscriber, err = amqp.NewSubscriber(
		consumerConfiguration(),
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		log.Fatalw("Unable to create consumer", "error", err)
	}

	messages, err := subscriber.Subscribe(ctx, config.GetString("queue.service.dispatcher.queue"))
	if err != nil {
		log.Fatalw("Unable to start consuming from topic", "error", err)
	}
	go process(messages, ch)
	log.Info("Consumer created")
	return
}

func process(messages <-chan *message.Message, ch chan<- *models.ProviderOperation) {
	handlers := []*audiosource.Handler{
		audiosource.NewSpotifyHandler(),
		audiosource.NewYoutubeHandler(),
		audiosource.NewYoutubeQueryHandler(),
	}
	manager := audiosource.NewSourceManager(handlers)
	log := common.GetLogger()
	log.Info("Consumer process created")
	for msg := range messages {
		var message models.MusicRequest
		if err := json.Unmarshal(msg.Payload, &message); err != nil {
			msg.Nack() // TODO check later if retry makes sense
		}
		log.Infow("Received message", "message", message)
		if _, err := manager.Handle(&message, ch); err != nil {
			log.Errorw("Error processing operation", "request", message)
			msg.Nack()
		} else {
			msg.Ack()
		}
	}
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
