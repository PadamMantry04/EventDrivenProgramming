package main

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/gofiber/fiber/v2"
)

type Comment struct {
	Text string `form:"text" json:"text"`
}

func main() {
	app := fiber.New()
	api := app.Group("/api/v1")
	api.Post("/comments", createComment) // basically defining a new route for the post route using the /api/v1 as the base route
	app.Listen(":3000")
}

func ConnectProducer(url []string) (sarama.SyncProducer, error) {
	// initial a new sarama.config
	// set appropriate params for sarama config
	// create a new connection using NewSyncProvider
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	conn, err := sarama.NewSyncProducer(url, config)
	if err != nil {
		panic(err)
	}
	return conn, nil
}

func PushCommentToQueue(topic string, message []byte) error {
	// connect to broker url
	// create a message in appropriate Producer format and then send it
	// once message is sent, return the topic, partition, offset
	brokersUrl := []string{"localhost:29092"}
	producer, err := ConnectProducer(brokersUrl)
	if err != nil {
		return err
	}

	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)

	return nil
}

// basically a callback which takes a fiber context and returns error.
func createComment(ctx *fiber.Ctx) error {
	// we have to retrieve the body of the context and then convert it to bytes to send to kafk
	cmt := new(Comment) // initialized a new comment
	if err := ctx.BodyParser(cmt); err != nil {
		panic(err)
	}
	// so the body has been parsed
	// now convert to bytes and send to kafka
	// once sent, display appropriate message back to user, whether success or failure
	cmtInBytes, err := json.Marshal(cmt)
	PushCommentToQueue("comments", cmtInBytes)
	if err != nil {
		ctx.Status(500).JSON(&fiber.Map{
			"success": "false",
			"message": "error creating comment",
		})
		return err
	}

	err = ctx.JSON(&fiber.Map{
		"success": "true",
		"message": "message pushed successfully",
		"comment": cmt,
	})
	return err

}
