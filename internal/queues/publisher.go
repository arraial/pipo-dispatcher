package queues

import (
	"log"

	config "github.com/spf13/viper"
	mq "github.com/wagslane/go-rabbitmq"
)

func Publisher(conn *mq.Conn) (*mq.Publisher, error) {

	publisher_options := mq.PublisherOptions{
		ExchangeOptions: mq.ExchangeOptions{
			Name:    config.GetString("player.queue.service.dispatcher.queue"),
			Declare: config.GetBool("player.queue.service.dispatcher.declare"),
			Durable: config.GetBool("player.queue.service.dispatcher.durable"),
			Args:    config.GetStringMap("player.queue.service.dispatcher.args"),
		},
	}

	publisher, err := mq.NewPublisher(
		conn, func(options *mq.PublisherOptions) {
			*options = publisher_options
		},
		mq.WithPublisherOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	// TODO review
	publisher.NotifyReturn(func(r mq.Return) {
		log.Printf("Message returned from server: %s", string(r.Body))
	})

	// TODO review
	publisher.NotifyPublish(func(c mq.Confirmation) {
		log.Printf("Message confirmed from server. tag: %v, ack: %v", c.DeliveryTag, c.Ack)
	})
	return publisher, err
}
