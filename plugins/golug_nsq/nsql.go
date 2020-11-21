package golug_nsq

import nsq "github.com/segmentio/nsq-go"

func init() {
	// Create a new consumer, looking up nsqd nodes from the listed nsqlookup
	// addresses, pulling messages from the 'world' channel of the 'hello' topic
	// with a maximum of 250 in-flight messages.
	consumer, _ := nsq.StartConsumer(nsq.ConsumerConfig{
		Topic:   "hello",
		Channel: "world",
		Lookup: []string{
			"nsqlookup-001.service.local:4161",
			"nsqlookup-002.service.local:4161",
			"nsqlookup-003.service.local:4161",
		},
		MaxInFlight: 250,
	})

	// Consume messages, the consumer automatically connects to the nsqd nodes
	// it discovers and handles reconnections if something goes wrong.
	for msg := range consumer.Messages() {
		// handle the message, then call msg.Finish or msg.Requeue
		// ...
		msg.Finish()
	}

	// Starts a new producer that publishes to the TCP endpoint of a nsqd node.
	// The producer automatically handles connections in the background.
	producer, _ := nsq.StartProducer(nsq.ProducerConfig{
		Topic:   "hello",
		Address: "localhost:4150",
	})

	// Publishes a message to the topic that this producer is configured for,
	// the method returns when the operation completes, potentially returning an
	// error if something went wrong.
	producer.Publish([]byte("Hello World!"))

	// Stops the producer, all in-flight requests will be canceled and no more
	// messages can be published through this producer.
	producer.Stop()
}
