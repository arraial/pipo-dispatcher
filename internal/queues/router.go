package queues

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	amqp "github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	amqpMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	audiosource "github.com/arraial/pipo-dispatcher/internal/audiosource"
	common "github.com/arraial/pipo-dispatcher/internal/common"
	models "github.com/arraial/pipo-dispatcher/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

type structHandler struct {
	publisher *amqp.Publisher
	manager   *audiosource.SourceManager
}

func (s structHandler) Handler(msg *amqpMessage.Message) (err error) {
	tracer := otel.Tracer("dispatcher")
	log := common.GetLogger()
	ctx, span := tracer.Start(msg.Context(), "handler.createOperation", trace.WithSpanKind(trace.SpanKindConsumer))
	log.Infow("Received message", "message", msg)

	var request models.MusicRequest
	var wg sync.WaitGroup
	if err = json.Unmarshal(msg.Payload, &request); err != nil {
		return
	}
	span.AddEvent(
		"unmarshalled music request",
		trace.WithAttributes(attribute.Bool("request.shuffle", request.Shuffle)),
		trace.WithAttributes(attribute.String("request.serverId", request.Server_id)),
		trace.WithAttributes(attribute.StringSlice("request.query", request.Query)),
		trace.WithAttributes(attribute.String("request.uuid", request.UUID.String())),
	)

	operations, err := s.manager.Handle(&request)
	if err != nil {
		log.Errorw("Error processing request", "request", request)
		return
	}
	wg.Add(len(operations))
	span.AddEvent("music request processed")

	for _, op := range operations {
		span.AddEvent("launching operation publish routine")
		go func(ct context.Context, operation *models.ProviderOperation) {
			defer wg.Done()
			c, sp := tracer.Start(
				ct,
				"handler.publishOperation",
				trace.WithSpanKind(trace.SpanKindProducer),
				trace.WithAttributes(attribute.Bool("operation.shuffle", operation.Shuffle)),
				trace.WithAttributes(attribute.String("operation.serverId", operation.Server_id)),
				trace.WithAttributes(attribute.String("operation.query", operation.Query)),
				trace.WithAttributes(attribute.String("operation.uuid", operation.UUID.String())),
				trace.WithAttributes(attribute.String("operation.type", operation.Operation)),
				trace.WithAttributes(attribute.String("operation.provider", string(operation.Provider))),
			)
			defer sp.End()

			data, err := json.Marshal(operation)
			if err != nil {
				log.Errorw("Unable to marshall message", "error", err)
				sp.SetStatus(codes.Error, "message not published")
				sp.RecordError(err)
				return
			} else {
				newMessage := amqpMessage.NewMessage(watermill.NewUUID(), data)
				newMessage.Metadata = msg.Metadata
				middleware.SetCorrelationID(operation.UUID.String(), newMessage)
				newMessage.SetContext(c)
				sp.AddEvent("message created", trace.WithAttributes(attribute.String("uuid", newMessage.UUID)))
				log.Debugw("Publishing message", "message", newMessage)
				s.publisher.Publish(
					fmt.Sprintf(
						"%s.%s",
						operation.Provider,
						operation.Operation,
					),
					newMessage,
				)
				sp.AddEvent("message published")
				sp.SetStatus(codes.Ok, "message published")
			}
		}(ctx, op)
	}
	span.End()
	wg.Wait()
	return
}

func CreateRouter(ctx context.Context) (router *amqpMessage.Router, err error) {
	log := common.GetLogger()
	logger := NewAdaptedZap(log.Desugar())
	config := common.GetConfig()
	router, err = amqpMessage.NewRouter(amqpMessage.RouterConfig{}, logger)
	if err != nil {
		log.Errorw("Unable to create router", "error", err)
		return
	}

	router.AddMiddleware(func(h amqpMessage.HandlerFunc) amqpMessage.HandlerFunc {
		return func(message *amqpMessage.Message) ([]*amqpMessage.Message, error) {
			propagators := propagation.NewCompositeTextMapPropagator(
				propagation.Baggage{},
				propagation.TraceContext{},
			)
			b, err := baggage.Parse(message.Metadata.Get("baggage"))
			if err != nil {
				log.Errorw("Unable to parse baggage", "error", err)
			} else {
				middleware.SetCorrelationID(b.Member("uuid").Value(), message)
			}
			carrier := propagation.MapCarrier{
				"baggage":     message.Metadata.Get("baggage"),
				"traceparent": message.Metadata.Get("traceparent"),
			}
			message.SetContext(
				propagators.Extract(
					message.Context(),
					carrier,
				),
			)
			message.UUID = watermill.NewUUID()
			return h(message)
		}
	})

	router.AddMiddleware(func(h amqpMessage.HandlerFunc) amqpMessage.HandlerFunc {
		return func(message *amqpMessage.Message) ([]*amqpMessage.Message, error) {
			tracer := otel.Tracer("dispatcher")
			ctx, span := tracer.Start(
				message.Context(),
				"handler",
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(attribute.String("message.uuid", message.UUID)),
				trace.WithAttributes(attribute.String("message.baggage", message.Metadata.Get("baggage"))),
			)
			defer span.End()
			md, _ := metadata.FromIncomingContext(message.Context()) // TODO check for errors
			ctx = metadata.NewOutgoingContext(ctx, md)
			message.SetContext(ctx)
			return h(message)
		}
	})

	router.AddMiddleware(
		middleware.CorrelationID,
		middleware.Retry{
			MaxRetries:      3,
			InitialInterval: time.Millisecond * 100,
			Logger:          logger,
		}.Middleware,
		middleware.Recoverer,
	)

	publisher, err := CreatePublisher()
	if err != nil {
		log.Errorw("Unable to create publisher", "error", err)
		return
	}
	subscriber, err := CreateConsumer()
	if err != nil {
		log.Errorw("Unable to create consumer", "error", err)
		return
	}

	router.AddNoPublisherHandler(
		config.GetString("queue.service.dispatcher.exchange.name"),
		config.GetString("queue.service.dispatcher.exchange.name"),
		subscriber,
		structHandler{publisher, audiosource.NewDefaultSourceManager()}.Handler,
	)

	return
}
