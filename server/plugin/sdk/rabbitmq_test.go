package sdk

import (
	"context"
	"fmt"
	"testing"

	"github.com/olive-io/gflow/server/config"
)

func TestNewRabbitmq(t *testing.T) {
	cfg := config.RabbitMQConfig{
		Username: "howlink",
		Password: "howlink",
		Host:     "192.168.141.128:5672",
	}

	rabbitmq, err := NewRabbitmq(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer rabbitmq.Close()

	ctx := context.TODO()
	topic := "hello"
	contentType := "application/json"
	//ch, cancel, err := rabbitmq.Receive(ctx, topic, contentType)
	//if err != nil {
	//	t.Fatalf("receive failed: %v", err)
	//}
	//defer cancel()

	msg := []byte(fmt.Sprint(`{"hello":"world"}`))
	//msg1 := []byte("")
	//done := make(chan struct{}, 1)
	//go func() {
	//	defer close(done)
	//	select {
	//	case d := <-ch:
	//		msg1 = d.Body
	//	}
	//}()

	if err = rabbitmq.Send(ctx, topic, contentType, msg); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	//<-done
	//
	//fmt.Println(string(msg1))
}
