package rabbitmq

import (
	"fmt"

	"github.com/streadway/amqp"
)

type RabbitClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

func newConnection(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to connecto to rabbitmq: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to open a channel: %v", err)
	}

	return conn, channel, nil
}

func NewRabbitClient(connectionURL string) (*RabbitClient, error) {
	conn, channel, err := newConnection(connectionURL)

	if err != nil {
		return nil, err
	}

	return &RabbitClient{
		conn:    conn,
		channel: channel,
		url:     connectionURL,
	}, nil
}

func (client *RabbitClient) ConsumeMessages(exchange, routingKey, queueName string) (<-chan amqp.Delivery, error) {
	if err := client.channel.ExchangeDeclare(
		exchange,
		"direct",
		true,
		true,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("Failed to declare exchange: %v", err)
	}

	queue, err := client.channel.QueueDeclare(queueName, true, true, false, false, nil)

	if err != nil {
		return nil, fmt.Errorf("Failed to declare queue: %v", err)
	}

	if err := client.channel.QueueBind(queue.Name, routingKey, exchange, false, nil); err != nil {
		return nil, fmt.Errorf("Failed to bind queue: %v", err)
	}

	messages, err := client.channel.Consume(
		queue.Name,
		"goapp",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to consume messages: %v", err)
	}

	return messages, nil
}

func (client *RabbitClient) Close() error {
	if err := client.channel.Close(); err != nil {
		return fmt.Errorf("Failed to close channel: %v", err)
	}

	if err := client.conn.Close(); err != nil {
		return fmt.Errorf("Failed to close connection: %v", err)
	}

	return nil
}
