package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	ConnectionUrl string
}

type RabbitMQConnection struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewConnection(cfg RabbitMQConfig) (RabbitMQConnection, error) {
	conn, err := amqp.Dial(cfg.ConnectionUrl)
	if err != nil {
		return RabbitMQConnection{}, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return RabbitMQConnection{}, err
	}

	return RabbitMQConnection{Conn: conn, Channel: channel}, nil
}

func (r *RabbitMQConnection) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}
	if r.Conn != nil {
		r.Conn.Close()
	}
}

func (r *RabbitMQConnection) Subscribe(queueName string, prefetchCount int, withDQL bool) (<-chan amqp.Delivery, error) {
	args := amqp.Table{}

	if withDQL {
		args["x-dead-letter-exchange"] = queueName + "_dlx"
		args["x-dead-letter-routing-key"] = queueName + "_dlq"
		if err := r.applyDQL(queueName, args); err != nil {
			return nil, err
		}
	}

	_, err := r.Channel.QueueDeclare(queueName, true, false, false, false, args)
	if err != nil {
		return nil, err
	}

	err = r.Channel.Qos(prefetchCount, 0, false)
	if err != nil {
		return nil, err
	}

	messages, err := r.Channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *RabbitMQConnection) applyDQL(queueName string, args amqp.Table) error {
	err := r.Channel.ExchangeDeclare(
		queueName+"_dlx",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = r.Channel.QueueDeclare(
		queueName+"_dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = r.Channel.QueueBind(
		queueName+"_dlq",
		queueName+"_dlq",
		queueName+"_dlx",
		false,
		nil,
	)

	args["x-dead-letter-exchange"] = queueName + "_dlx"
	args["x-dead-letter-routing-key"] = queueName + "_dlq"

	return err
}
