/*
Copyright 2025 The gflow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sdk

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/olive-io/gflow/server/config"
)

type Rabbitmq struct {
	cfg  *config.RabbitMQConfig
	conn *amqp.Connection
}

func NewRabbitmq(cfg *config.RabbitMQConfig) (*Rabbitmq, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s/", cfg.Username, cfg.Password, cfg.Host)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect to RabbitMQ: %v", err)
	}

	mq := &Rabbitmq{
		cfg:  cfg,
		conn: conn,
	}

	return mq, nil
}

func (mq *Rabbitmq) Send(ctx context.Context, topic string, contentType string, data []byte) error {
	ch, err := mq.conn.Channel()
	if err != nil {
		return fmt.Errorf("create channel: %v", err)
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(topic, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("queue declare: %v", err)
	}

	msg := amqp.Publishing{
		ContentType: contentType,
		Body:        data,
	}
	err = ch.PublishWithContext(ctx, "", queue.Name, false, false, msg)
	if err != nil {
		return fmt.Errorf("publish message: %v", err)
	}
	return nil
}

func (mq *Rabbitmq) Receive(ctx context.Context, topic, contentType string) (<-chan amqp.Delivery, func(), error) {
	ch, err := mq.conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("create channel: %v", err)
	}
	cancel := func() { ch.Close() }

	queue, err := ch.QueueDeclare(topic, true, false, false, false, nil)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("queue declare: %v", err)
	}

	dch, err := ch.ConsumeWithContext(ctx, queue.Name, "", true, false, false, false, nil)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("consume: %v", err)
	}

	return dch, cancel, nil
}

func (mq *Rabbitmq) Close() error {
	return mq.conn.Close()
}
