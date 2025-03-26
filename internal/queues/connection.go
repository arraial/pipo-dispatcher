package queues

import (
	common "github.com/arraial/pipo-dispatcher/internal/common"
	mq "github.com/wagslane/go-rabbitmq"
)

func Connection(brokerUrl string) (*mq.Conn, error) {

	var log = common.GetLogger()

	conn, err := mq.NewConn(
		brokerUrl,
		mq.WithConnectionOptionsLogging,
		mq.WithConnectionOptionsLogger(mq.Logger(common.GetLogger())),
	)
	if err != nil {
		log.Fatal("Unable to create connection", "error", err)
	}
	return conn, err
}
